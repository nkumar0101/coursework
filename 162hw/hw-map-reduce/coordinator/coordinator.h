/**
 * The MapReduce coordinator.
 */

#ifndef H1_H__
#define H1_H__
#include "../rpc/rpc.h"
#include "../lib/lib.h"
#include "../app/app.h"
#include "job.h"
#include <glib.h>
#include <stdio.h>
#include <stdbool.h>
#include <time.h>
#include <sys/time.h>
#include <unistd.h>
#include <stdlib.h>
#include <rpc/pmap_clnt.h>
#include <string.h>
#include <memory.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <sys/types.h>
#include <sys/stat.h>

typedef struct {
  int next_job_id;
  GHashTable* job_map; 
  GList* job_queue;
} coordinator;

typedef struct {
  int job_id;
  bool done;
  bool failed;

  
  struct {
    u_int files_len; 
    path *files_val;
  } files;

  path output_dir;
  char *app;
  int n_reduce;

  struct {
    u_int args_len;
    char *args_val;
  } args;

  GHashTable* map_tasks;
  GHashTable* reduce_tasks;
  bool map_tasks_done;
} job;

typedef struct {
  int status; // 0 or 1
  time_t time_assigned;
} task;


void coordinator_init(coordinator** coord_ptr);
#endif
