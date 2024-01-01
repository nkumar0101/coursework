/**
 * The MapReduce coordinator.
 */

#include "coordinator.h"

#ifndef SIG_PF
#define SIG_PF void (*)(int)
#endif

/* Global coordinator state. */
coordinator* state;

extern void coordinator_1(struct svc_req*, SVCXPRT*);
static bool is_done(GHashTable* ht, int total);

/* Set up and run RPC server. */
int main(int argc, char** argv) {
  register SVCXPRT* transp;

  pmap_unset(COORDINATOR, COORDINATOR_V1);

  transp = svcudp_create(RPC_ANYSOCK);
  if (transp == NULL) {
    fprintf(stderr, "%s", "cannot create udp service.");
    exit(1);
  }
  if (!svc_register(transp, COORDINATOR, COORDINATOR_V1, coordinator_1, IPPROTO_UDP)) {
    fprintf(stderr, "%s", "unable to register (COORDINATOR, COORDINATOR_V1, udp).");
    exit(1);
  }

  transp = svctcp_create(RPC_ANYSOCK, 0, 0);
  if (transp == NULL) {
    fprintf(stderr, "%s", "cannot create tcp service.");
    exit(1);
  }
  if (!svc_register(transp, COORDINATOR, COORDINATOR_V1, coordinator_1, IPPROTO_TCP)) {
    fprintf(stderr, "%s", "unable to register (COORDINATOR, COORDINATOR_V1, tcp).");
    exit(1);
  }

  coordinator_init(&state);

  svc_run();
  fprintf(stderr, "%s", "svc_run returned");
  exit(1);
  /* NOTREACHED */
}

/* EXAMPLE RPC implementation. */
int* example_1_svc(int* argp, struct svc_req* rqstp) {
  static int result;

  result = *argp + 1;

  return &result;
}

/* SUBMIT_JOB RPC implementation. */
int* submit_job_1_svc(submit_job_request* argp, struct svc_req* rqstp) {
  static int result;

  printf("Received submit job request\n");

  /* TODO */
  if (get_app(argp->app).name == NULL) {
    result = -1;
    return &result;
  }
  
  job* new_j = (job*) malloc(sizeof(job));
  new_j->job_id = state->next_job_id++;
  new_j->done = false;
  new_j->failed = false;

  u_int number_of_files = argp->files.files_len;
  new_j->files.files_len = number_of_files;
  new_j->files.files_val = (path*) malloc(sizeof(path) * number_of_files);
  for (int i = 0; i < number_of_files; i++) {
    new_j->files.files_val[i] = strdup(argp->files.files_val[i]);
  }

  new_j->output_dir = strdup(argp->output_dir);
  new_j->app = strdup(argp->app);
  new_j->n_reduce = argp->n_reduce;

  u_int args_len = argp->args.args_len;
  new_j->args.args_len = args_len;
  new_j->args.args_val = (char*) malloc(args_len);
  memcpy((void*) new_j->args.args_val, (void*) argp->args.args_val, args_len);

  new_j->map_tasks = g_hash_table_new_full(g_direct_hash, g_direct_equal, NULL, NULL);
  new_j->reduce_tasks = g_hash_table_new_full(g_direct_hash, g_direct_equal, NULL, NULL);
  new_j->map_tasks_done = false;

  g_hash_table_insert(state->job_map, GINT_TO_POINTER(new_j->job_id), new_j);
  state->job_queue = g_list_append(state->job_queue, GINT_TO_POINTER(new_j->job_id));
  result = new_j->job_id;

  /* Do not modify the following code. */
  /* BEGIN */
  struct stat st;
  if (stat(argp->output_dir, &st) == -1) {
    mkdirp(argp->output_dir);
  }

  return &result;
  /* END */
}

/* POLL_JOB RPC implementation. */
poll_job_reply* poll_job_1_svc(int* argp, struct svc_req* rqstp) {
  static poll_job_reply result;

  printf("Received poll job request\n");

  /* TODO */

  if (argp == NULL) {
    result.done = false;
    result.failed = false;
    result.invalid_job_id = true;
    return &result;
  }
  job* lookup = g_hash_table_lookup(state->job_map, GINT_TO_POINTER(*argp));
  if (lookup == NULL) {
    result.done = false;
    result.failed = false;
    result.invalid_job_id = true;
  } else {
    result.done = lookup->done;
    result.failed = lookup->failed;
    result.invalid_job_id = false;
  }

  return &result;
}

/* GET_TASK RPC implementation. */
get_task_reply* get_task_1_svc(void* argp, struct svc_req* rqstp) {
  static get_task_reply result;

  printf("Received get task request\n");
  result.file = "";
  result.output_dir = "";
  result.app = "";
  result.wait = true;
  result.args.args_len = 0;

  /* TODO */

  bool searching = true;
  for (GList* elem = state->job_queue; elem && searching; elem = elem->next) {
    int job_id = GPOINTER_TO_INT(elem->data);
    job* j = g_hash_table_lookup(state->job_map, GINT_TO_POINTER(job_id));
    if (j == NULL) {
      printf("get task failed: job %d in list but not found in map???\n", job_id);
      return &result;
    }
    if (j->map_tasks_done) {
      // get reduce task
      for (int i = 0; i < j->n_reduce && searching; i++) {
        if (!g_hash_table_contains(j->reduce_tasks, GINT_TO_POINTER(i))) {
          result.job_id = j->job_id;
          result.task = i;
          result.output_dir = j->output_dir;
          result.app = j->app;
          result.n_map = j->files.files_len;
          result.n_reduce = j->n_reduce;
          result.reduce = true;
          result.wait = false;
          result.args.args_len = j->args.args_len;
          result.args.args_val = j->args.args_val;

          task* new_task = (task*) malloc(sizeof(task));
          new_task->status = 0;
          new_task->time_assigned = time(NULL);
          g_hash_table_insert(j->reduce_tasks, GINT_TO_POINTER(i), new_task);
          searching = false;
        } else {
          task* t = g_hash_table_lookup(j->reduce_tasks, GINT_TO_POINTER(i));
          if (t->status == 1 || time(NULL) - t->time_assigned < TASK_TIMEOUT_SECS)
            continue;

          result.job_id = j->job_id;
          result.task = i;
          result.output_dir = j->output_dir;
          result.app = j->app;
          result.n_map = j->files.files_len;
          result.n_reduce = j->n_reduce;
          result.reduce = true;
          result.wait = false;
          result.args.args_len = j->args.args_len;
          result.args.args_val = j->args.args_val;

          t->time_assigned = time(NULL);
          searching = false;
        }
      }
    } else {
      // get map task
      for (int i = 0; i < j->files.files_len && searching; i++) {
        if (!g_hash_table_contains(j->map_tasks, GINT_TO_POINTER(i))) {
          result.job_id = j->job_id;
          result.task = i;
          result.file = j->files.files_val[i];
          result.output_dir = j->output_dir;
          result.app = j->app;
          result.n_map = j->files.files_len;
          result.n_reduce = j->n_reduce;
          result.reduce = false;
          result.wait = false;
          result.args.args_len = j->args.args_len;
          result.args.args_val = j->args.args_val;

          task* new_task = (task*) malloc(sizeof(task));
          new_task->status = 0;
          new_task->time_assigned = time(NULL);
          g_hash_table_insert(j->map_tasks, GINT_TO_POINTER(i), new_task);
          searching = false;
        } else {
          task* t = g_hash_table_lookup(j->map_tasks, GINT_TO_POINTER(i));
          if (t->status == 1 || time(NULL) - t->time_assigned < TASK_TIMEOUT_SECS)
            continue;

          result.job_id = j->job_id;
          result.task = i;
          result.file = j->files.files_val[i];
          result.output_dir = j->output_dir;
          result.app = j->app;
          result.n_map = j->files.files_len;
          result.n_reduce = j->n_reduce;
          result.reduce = false;
          result.wait = false;
          result.args.args_len = j->args.args_len;
          result.args.args_val = j->args.args_val;
          
          t->time_assigned = time(NULL);
          searching = false;
        }
      }
    }
  }

  return &result;
}

/* FINISH_TASK RPC implementation. */
void* finish_task_1_svc(finish_task_request* argp, struct svc_req* rqstp) {
  static char* result;

  printf("Received finish task request\n");

  /* TODO */

  job* j = g_hash_table_lookup(state->job_map, GINT_TO_POINTER(argp->job_id));
  if (j == NULL) {
    printf("finish request failed: job %d not found\n", argp->job_id);
    return (void*)&result;
  }

  if (argp->success == false) {
    j->done = true;
    j->failed = true;
    state->job_queue = g_list_remove(state->job_queue, GINT_TO_POINTER(j->job_id));
  } else {
    if (argp->reduce) {
      task* t = g_hash_table_lookup(j->reduce_tasks, GINT_TO_POINTER(argp->task));
      t->status = 1;
      if (is_done(j->reduce_tasks, j->n_reduce)) {
        j->done = true;
        j->failed = false;
        state->job_queue = g_list_remove(state->job_queue, GINT_TO_POINTER(j->job_id));
      }
    } else {
      task* t = g_hash_table_lookup(j->map_tasks, GINT_TO_POINTER(argp->task));
      t->status = 1;
      if (is_done(j->map_tasks, j->files.files_len)) {
        j->map_tasks_done = true;
      }
    }
  }

  return (void*)&result;
}

static bool is_done(GHashTable* ht, int total) {
  GList* tasks = g_hash_table_get_keys(ht);
  int count = 0;
  for (GList* elem = tasks; elem; elem = elem->next) {
    int task_id = GPOINTER_TO_INT(elem->data);
    task* t = g_hash_table_lookup(ht, GINT_TO_POINTER(task_id));
    if (t->status != 1) return false;
    count++;
  }
  return count == total;
}

/* Initialize coordinator state. */
void coordinator_init(coordinator** coord_ptr) {
  *coord_ptr = malloc(sizeof(coordinator));

  coordinator* coord = *coord_ptr;

  /* TODO */
  coord->next_job_id = 0;
  coord->job_map = g_hash_table_new_full(g_direct_hash, g_direct_equal, NULL, NULL);
  coord->job_queue = NULL;
}
