#include <ctype.h>
#include <errno.h>
#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <signal.h>
#include <sys/wait.h>
#include <termios.h>
#include <unistd.h>

#include "tokenizer.h"

/* Convenience macro to silence compiler warnings about unused function parameters. */
#define unused __attribute__((unused))

/* Whether the shell is connected to an actual terminal or not. */
bool shell_is_interactive;

/* File descriptor for the shell input */
int shell_terminal;

/* Terminal mode settings for the shell */
struct termios shell_tmodes;

/* Process group id for the shell */
pid_t shell_pgid;

int cmd_exit(struct tokens* tokens);
int cmd_help(struct tokens* tokens);
int cmd_cd(struct tokens* tokens);
int cmd_pwd(struct tokens* tokens);
int fork_child(char** argv, int input_fd, int output_fd, int* pgid);
void install_sighandler(bool ignore);

/* Built-in command functions take token array (see parse.h) and return int */
typedef int cmd_fun_t(struct tokens* tokens);

/* Built-in command struct and lookup table */
typedef struct fun_desc {
  cmd_fun_t* fun;
  char* cmd;
  char* doc;
} fun_desc_t;

fun_desc_t cmd_table[] = {
    {cmd_help, "?", "show this help menu"},
    {cmd_exit, "exit", "exit the command shell"},
    {cmd_pwd, "pwd", "prints the current working directory"},
    {cmd_cd, "cd", "changes the current working directory to given directory"}
};

/* Prints a helpful description for the given command */
int cmd_help(unused struct tokens* tokens) {
  for (unsigned int i = 0; i < sizeof(cmd_table) / sizeof(fun_desc_t); i++)
    printf("%s - %s\n", cmd_table[i].cmd, cmd_table[i].doc);
  return 1;
}

/* Exits this shell */
int cmd_exit(unused struct tokens* tokens) { exit(0); }

/* Prints the current working directory to standard output*/
int cmd_pwd(unused struct tokens* tokens) {
  char* buf = getcwd(NULL, 0);
  printf("%s\n", buf);
  free(buf);
  return 1;
}

/* Changes the current working directory to the given directory.*/
int cmd_cd(struct tokens* tokens) {
  // for (int i = 0; i < tokens_get_length(tokens); i++) {
  //   printf("%s\n", tokens_get_token(tokens, i));
  // }
  if (tokens_get_length(tokens) != 2) {
    printf("Incorrect syntax. cd requires exactly 1 argument.");
    return -1;
  }
  int val = chdir(tokens_get_token(tokens, 1));
  if (val != 0) {
    perror("Error");
    return -1;
  }
  return 1;
  
}

int run_userprog(struct tokens* tokens) {
  //signal(SIGINT, SIG_IGN);
  int i;
  int output_fd = -1;
  int pipe_fds[2];
  int ifd = -1;
  int ofd = -1;
  int start_loc = 0;
  int pgid = -1; 
  char **argv = malloc(sizeof(char *) * (tokens_get_length(tokens) + 1));
  for (i = 0; i < tokens_get_length(tokens); i++) {
    argv[i] = tokens_get_token(tokens, i);
    if (strcmp(argv[i],"<") == 0) {
      char* input_file = tokens_get_token(tokens, i + 1);
      ifd = open(input_file, O_RDONLY); 
      if (ifd == -1) {
        perror("Error");
        exit(0);
      }
      argv[i] = NULL;
      
    }
    else if (strcmp(argv[i],">") == 0) {
      char* output_file = tokens_get_token(tokens, i + 1);
      output_fd = open(output_file, O_WRONLY | O_CREAT | O_TRUNC, S_IRWXU);
      if (output_fd == -1) {
        perror("Error");
        exit(0);
      }
      argv[i] = NULL;
      
    }
    else if (strcmp(argv[i],"|") == 0) {
      argv[i] = NULL;
      if (pipe(pipe_fds) == -1) {
        perror("Pipe error");
        exit(0);
      }
    
      ofd = pipe_fds[1];
      fork_child(&argv[start_loc], ifd, ofd, &pgid);
      ifd = pipe_fds[0];
      start_loc = i + 1;
     }
  }
  argv[i] = NULL;
  fork_child(&argv[start_loc], ifd, output_fd, &pgid);
  waitid(P_PGID, pgid, NULL, WEXITED | WSTOPPED);
  tcsetpgrp(0, getpid());
  //while(waitpid(-1, NULL, WUNTRACED) > 0);
  return 1;


}

int fork_child(char** argv, int input_fd, int output_fd, int *pgid) {
  //printf("Main proc: Argv[0]: %s Input fd: %d Output fd: %d\n", argv[0], input_fd, output_fd);
  
  pid_t child = fork();
  if (child > 0) { //parent
    // waitpid(child, NULL, 0);
    if (*pgid == -1) {
      *pgid = child;
    }
    setpgid(child, *pgid);
    
    tcsetpgrp(0, *pgid);
    
    if (input_fd >= 0) {
      close(input_fd);
    } else {
      //signal(SIGTTOU, SIG_IGN);
    }
    if (output_fd >= 0) {
      close(output_fd);
    }
  } else if (child == 0) {
    if (*pgid == -1) {
      *pgid = getpid();
    }
    setpgid(0, *pgid);
    install_sighandler(false);
    //signal(SIGTTOU, SIG_DFL);
    // signal(SIGTTIN, SIG_DFL);
    //signal(SIGTSTP, SIG_DFL);
    //signal(SIGQUIT, SIG_DFL);
    //signal(SIGINT, SIG_DFL);
    if (input_fd >= 0) {
      close(STDIN_FILENO);
      dup2(input_fd, STDIN_FILENO);
    } 
    if (output_fd >= 0) {
      close(STDOUT_FILENO);
      dup2(output_fd, STDOUT_FILENO);
    }
    //printf("Child proc: Argv[0]: %s Input fd: %d Output fd: %d\n", argv[0], input_fd, output_fd);
    if (execv(argv[0], argv) == -1) {
      char* strng = getenv("PATH");
      char* str_tkn;
      char buffer[4096*2 + 1];
      while ((str_tkn = strtok_r(strng, ":", &strng))) {                
        snprintf(buffer,4096*2 + 1, "%s/%s", str_tkn, argv[0]);
        //printf("%s\n", buffer);
        if (execv(buffer, argv) != -1) {
          break;
        }
      }
    }
    
    exit(0);
  } else {
    perror("Error");
    return -1;
  }
  

  return 0;
}

static const int all_signals[] = {
    SIGINT,
    SIGSEGV,
    SIGQUIT,
    SIGTSTP,
    SIGCONT,
    SIGTTIN,
    SIGTTOU,
    SIGTERM
};

void install_sighandler(bool ignore) {
  //printf("pgid: %d pgrp: %d\n", getpgid(0), getpgrp());
  struct sigaction sig_act;
  struct sigaction old_act;
  int i = 0;

  memset(&sig_act, 0, sizeof sig_act);
  sig_act.sa_flags = 0;
  sigemptyset(&sig_act.sa_mask);
  if (ignore) {
    sig_act.sa_handler = SIG_IGN;
  } else {
    sig_act.sa_handler = SIG_DFL;
    sig_act.sa_flags = SA_RESETHAND;
  }
  
  

  do {
      //printf("i: %d\n", i);
      if (sigaction(all_signals[i], &sig_act, &old_act)) {
          perror("Signal Error");
          exit(0);
      }
      //printf("i: %d handler: %ld flag: %d \n", i, (unsigned long) old_act.sa_handler, old_act.sa_flags);
  } while (all_signals[i++] != SIGTERM);
}


/* Looks up the built-in command, if it exists. */
int lookup(char cmd[]) {
  for (unsigned int i = 0; i < sizeof(cmd_table) / sizeof(fun_desc_t); i++)
    if (cmd && (strcmp(cmd_table[i].cmd, cmd) == 0))
      return i;
  return -1;
}

/* Intialization procedures for this shell */
void init_shell() {
  /* Our shell is connected to standard input. */
  shell_terminal = STDIN_FILENO;

  /* Check if we are running interactively */
  shell_is_interactive = isatty(shell_terminal);

  if (shell_is_interactive) {
    /* If the shell is not currently in the foreground, we must pause the shell until it becomes a
     * foreground process. We use SIGTTIN to pause the shell. When the shell gets moved to the
     * foreground, we'll receive a SIGCONT. */
    while (tcgetpgrp(shell_terminal) != (shell_pgid = getpgrp()))
      kill(-shell_pgid, SIGTTIN);

    /* Saves the shell's process id */
    shell_pgid = getpid();

    /* Take control of the terminal */
    tcsetpgrp(shell_terminal, shell_pgid);

    /* Save the current termios to a variable, so it can be restored later. */
    tcgetattr(shell_terminal, &shell_tmodes);
  }
}

int main(unused int argc, unused char* argv[]) {
  init_shell();

  static char line[4096];
  int line_num = 0;

  /* Please only print shell prompts when standard input is not a tty */
  if (shell_is_interactive)
    fprintf(stdout, "%d: ", line_num);

  install_sighandler(true);  
  while (fgets(line, 4096, stdin)) {
    /* Split our line into words. */
    struct tokens* tokens = tokenize(line);

    /* Find which built-in function to run. */
    int fundex = lookup(tokens_get_token(tokens, 0));

    if (fundex >= 0) {
      cmd_table[fundex].fun(tokens);
    } else {
      /* REPLACE this to run commands as programs. */
      //fprintf(stdout, "This shell doesn't know how to run programs.\n");
      run_userprog(tokens);
    }

    if (shell_is_interactive)
      /* Please only print shell prompts when standard input is not a tty */
      fprintf(stdout, "%d: ", ++line_num);

    /* Clean up memory */
    tokens_destroy(tokens);
  }

  return 0;
}
