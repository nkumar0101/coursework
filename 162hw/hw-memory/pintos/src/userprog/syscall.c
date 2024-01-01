#include "userprog/syscall.h"
#include <stdio.h>
#include <string.h>
#include <syscall-nr.h>
#include "threads/interrupt.h"
#include "threads/thread.h"
#include "threads/vaddr.h"
#include "filesys/filesys.h"
#include "filesys/file.h"
#include "threads/palloc.h"
#include "userprog/pagedir.h"

static void syscall_handler(struct intr_frame*);

void syscall_init(void) { intr_register_int(0x30, 3, INTR_ON, syscall_handler, "syscall"); }

void syscall_exit(int status) {
  printf("%s: exit(%d)\n", thread_current()->name, status);
  thread_exit();
}

/*
 * This does not check that the buffer consists of only mapped pages; it merely
 * checks the buffer exists entirely below PHYS_BASE.
 */
static void validate_buffer_in_user_region(const void* buffer, size_t length) {
  uintptr_t delta = PHYS_BASE - buffer;
  if (!is_user_vaddr(buffer) || length > delta)
    syscall_exit(-1);
}

/*
 * This does not check that the string consists of only mapped pages; it merely
 * checks the string exists entirely below PHYS_BASE.
 */
static void validate_string_in_user_region(const char* string) {
  uintptr_t delta = PHYS_BASE - (const void*)string;
  if (!is_user_vaddr(string) || strnlen(string, delta) == delta)
    syscall_exit(-1);
}

static int syscall_open(const char* filename) {
  struct thread* t = thread_current();
  if (t->open_file != NULL)
    return -1;

  t->open_file = filesys_open(filename);
  if (t->open_file == NULL)
    return -1;

  return 2;
}

static int syscall_write(int fd, void* buffer, unsigned size) {
  struct thread* t = thread_current();
  if (fd == STDOUT_FILENO) {
    putbuf(buffer, size);
    return size;
  } else if (fd != 2 || t->open_file == NULL)
    return -1;

  return (int)file_write(t->open_file, buffer, size);
}

static int syscall_read(int fd, void* buffer, unsigned size) {
  struct thread* t = thread_current();
  if (fd != 2 || t->open_file == NULL)
    return -1;

  return (int)file_read(t->open_file, buffer, size);
}

static void syscall_close(int fd) {
  struct thread* t = thread_current();
  if (fd == 2 && t->open_file != NULL) {
    file_close(t->open_file);
    t->open_file = NULL;
  }
}

static uint8_t* syscall_sbrk(int increment) {
  struct thread* t = thread_current();
  //printf("Brk addr: %08x\n", t->heap_brk_address);
  uint32_t old_heap_brk = t->heap_brk_address;
  uint32_t new_brk_addr = t->heap_brk_address;
  new_brk_addr += increment;

  if (increment == 0) {
    return old_heap_brk;
  }

  if (increment > 0) {
    uint32_t start_page_num = old_heap_brk / PGSIZE;
    if (old_heap_brk % PGSIZE) {
      start_page_num += 1;
    }
    uint32_t end_page_num = new_brk_addr / PGSIZE;
    if (new_brk_addr % PGSIZE) {
      end_page_num += 1;
    }
    for (uint32_t i = start_page_num; i < end_page_num; i++) {
      uint32_t kpage = palloc_get_page(PAL_USER | PAL_ZERO);
      if (kpage == NULL) {
        //printf("increment > 0, kpage not found --> i:%08x\n", i);
        for (uint32_t j = start_page_num; j < i; j++) {
          uint32_t kpage = pagedir_get_page(t->pagedir, j*PGSIZE);
          pagedir_clear_page(t->pagedir, j*PGSIZE);
          palloc_free_page(kpage);
        }
        return (void*) -1;

      }
      
      if (!(pagedir_get_page(t->pagedir, i*PGSIZE) == NULL && pagedir_set_page(t->pagedir, i*PGSIZE, kpage, true))) {
        //printf("increment > 0 --> kpage:%08x, i:%08x\n", kpage, i);
        palloc_free_page(kpage);
        for (uint32_t j = start_page_num; j < i; j++) {
          uint32_t kpage = pagedir_get_page(t->pagedir, j*PGSIZE);
          pagedir_clear_page(t->pagedir, j*PGSIZE);
          palloc_free_page(kpage);
        }
        return (void*) -1;
      }
    }
  }

  if (increment < 0) {
    uint32_t start_page_num = new_brk_addr / PGSIZE;
    if (new_brk_addr % PGSIZE) {
      start_page_num += 1;
    }
    uint32_t end_page_num = old_heap_brk / PGSIZE;
    if (old_heap_brk % PGSIZE) {
      end_page_num += 1;
    }

    for (uint32_t i = start_page_num; i < end_page_num; i++) {
      uint32_t kpage = pagedir_get_page(t->pagedir, i*PGSIZE);
      pagedir_clear_page(t->pagedir, i*PGSIZE);
      palloc_free_page(kpage);
    }
  }  
  //printf("old heap brk:%08x, current brk:%08x\n", old_heap_brk, new_brk_addr);
  t->heap_brk_address = new_brk_addr;
  //printf("returning heap brk addr:%08x\n", t->heap_brk_address);
  return old_heap_brk;
}

static uint8_t* copy_syscall_sbrk(int increment) {
  struct thread* t = thread_current();
  //printf("Brk addr: %08x\n", t->heap_brk_address);
  uint32_t old_heap_brk = t->heap_brk_address;
  uint32_t curr_brk_addr = t->heap_brk_address;
  curr_brk_addr += increment;
  uint32_t new_page_addr = curr_brk_addr & ~PGMASK;
  uint32_t start_page_addr = t->heap_brk_address & ~PGMASK;
  //printf("new page num:%08x, start page num:%08x\n", new_page_addr, start_page_addr);
  if (t->heap_start == t->heap_brk_address) {
    if (increment <= 0) {
      //printf("increment < 0 --> %d\n", increment);
      return (void*) -1;
    }
    uint32_t kpage = palloc_get_page(PAL_USER | PAL_ZERO);
    if (kpage == NULL) {
      //printf("heap start no page --> i:%08x\n", t->heap_start);
      return (void*) -1;
    }
    
    
    if (!(pagedir_get_page(t->pagedir, t->heap_start) == NULL && pagedir_set_page(t->pagedir, t->heap_start, kpage, true))) {
      //printf("heap start --> kpage:%08x, i:%08x\n", kpage, t->heap_start);
      palloc_free_page(kpage);
      
      return (void*) -1;
    }
  }
  // if (increment < 0) {
  //   printf("increment: %d, new page addr:%08x, start page addr: %08x\n", increment, new_page_addr, start_page_addr);
  // }
  if (new_page_addr != start_page_addr) {
    if (increment > 0) {
      //allocate new pages and activate them
      for (uint32_t i = start_page_addr + PGSIZE; i <= new_page_addr; i+= PGSIZE) {
        uint32_t kpage = palloc_get_page(PAL_USER | PAL_ZERO);
        if (kpage == NULL) {
          //printf("increment > 0, kpage not found --> i:%08x\n", i);
          for (uint32_t j = old_heap_brk; j < i; j+= PGSIZE) {
            uint32_t kpage = pagedir_get_page(t->pagedir, j);
            pagedir_clear_page(t->pagedir, j);
            palloc_free_page(kpage);
          }
          return (void*) -1;

        }
        
        if (!(pagedir_get_page(t->pagedir, i) == NULL && pagedir_set_page(t->pagedir, i, kpage, true))) {
          //printf("increment > 0 --> kpage:%08x, i:%08x\n", kpage, i);
          palloc_free_page(kpage);
          for (uint32_t j = old_heap_brk; j < i; j+= PGSIZE) {
            uint32_t kpage = pagedir_get_page(t->pagedir, j);
            pagedir_clear_page(t->pagedir, j);
            palloc_free_page(kpage);
          }
          return (void*) -1;
        }
        //t->heap_brk_address = i + (PGSIZE - 1);
      }
    } else {
      //deallocate unused pages
      
      for (uint32_t i = new_page_addr; i <= (start_page_addr + PGSIZE); i+= PGSIZE) {
        uint32_t kpage = pagedir_get_page(t->pagedir, i);
        pagedir_clear_page(t->pagedir, i);
        palloc_free_page(kpage);
      }
    }
  }
  //printf("old heap brk:%08x, current brk:%08x\n", old_heap_brk, curr_brk_addr);
  t->heap_brk_address = curr_brk_addr;

  //printf("returning heap brk addr:%08x\n", t->heap_brk_address);
  return old_heap_brk;
}

static void syscall_handler(struct intr_frame* f) {
  uint32_t* args = (uint32_t*)f->esp;
  struct thread* t = thread_current();
  t->in_syscall = true;

  validate_buffer_in_user_region(args, sizeof(uint32_t));
  switch (args[0]) {
    case SYS_EXIT:
      validate_buffer_in_user_region(&args[1], sizeof(uint32_t));
      syscall_exit((int)args[1]);
      break;

    case SYS_OPEN:
      validate_buffer_in_user_region(&args[1], sizeof(uint32_t));
      validate_string_in_user_region((char*)args[1]);
      f->eax = (uint32_t)syscall_open((char*)args[1]);
      break;

    case SYS_WRITE:
      validate_buffer_in_user_region(&args[1], 3 * sizeof(uint32_t));
      validate_buffer_in_user_region((void*)args[2], (unsigned)args[3]);
      f->eax = (uint32_t)syscall_write((int)args[1], (void*)args[2], (unsigned)args[3]);
      break;

    case SYS_READ:
      validate_buffer_in_user_region(&args[1], 3 * sizeof(uint32_t));
      validate_buffer_in_user_region((void*)args[2], (unsigned)args[3]);
      f->eax = (uint32_t)syscall_read((int)args[1], (void*)args[2], (unsigned)args[3]);
      break;

    case SYS_CLOSE:
      validate_buffer_in_user_region(&args[1], sizeof(uint32_t));
      syscall_close((int)args[1]);
      break;

    case SYS_SBRK:
      validate_buffer_in_user_region(&args[1], sizeof(uint32_t));
      f->eax = (uint32_t)syscall_sbrk((int)args[1]);
      break;

    default:
      printf("Unimplemented system call: %d\n", (int)args[0]);
      break;
  }

  t->in_syscall = false;
}
