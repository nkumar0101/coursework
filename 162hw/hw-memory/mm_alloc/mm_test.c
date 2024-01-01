#include <assert.h>
#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>

/* Function pointers to hw3 functions */
void* (*mm_malloc)(size_t);
void* (*mm_realloc)(void*, size_t);
void (*mm_free)(void*);

static void* try_dlsym(void* handle, const char* symbol) {
  char* error;
  void* function = dlsym(handle, symbol);
  if ((error = dlerror())) {
    fprintf(stderr, "%s\n", error);
    exit(EXIT_FAILURE);
  }
  return function;
}

static void load_alloc_functions() {
  void* handle = dlopen("hw3lib.so", RTLD_NOW);
  if (!handle) {
    fprintf(stderr, "%s\n", dlerror());
    exit(EXIT_FAILURE);
  }

  mm_malloc = try_dlsym(handle, "mm_malloc");
  mm_realloc = try_dlsym(handle, "mm_realloc");
  mm_free = try_dlsym(handle, "mm_free");
}


int main() {
  load_alloc_functions();

  int* data = mm_malloc(sizeof(int));
  assert(data != NULL);
  data[0] = 0x162;
  mm_free(data);
  int* new_data = mm_malloc(sizeof(int));
  if (*new_data != 0 || new_data != data) {
    puts("not successful");
    return 0;
  }
  int* dataA = mm_malloc(100*sizeof(int));
  //printf("Data A:%08x\n",dataA);
  int* dataB = mm_malloc(101*sizeof(int));
  //printf("Data B:%08x\n",dataB);
  mm_free(dataA);
  int* dataA0 = mm_malloc(20*sizeof(int));
  int* dataA1 = mm_malloc(20*sizeof(int));
  //printf("Data A0:%08x, Data A1:%08x\n",dataA0, dataA1);
  if (dataA > dataA0 || dataA0 > dataB) {
    puts("A0 not within A and B");
    return 0;
  }
  if (dataA > dataA1 || dataA1 > dataB) {
    puts("A1 not within A and B");
    return 0;
  }

  mm_free(dataB);
  mm_free(dataA0);
  mm_free(dataA1);
  puts("malloc test successful!");


  
}
