/*
 * mm_alloc.c
 */

#include "mm_alloc.h"

#include <stdlib.h>
#include <unistd.h>
#include <stdio.h>
#include <string.h>
#include <inttypes.h>
#include <stdbool.h>


typedef struct metadata{
  struct metadata* prev;
  struct metadata* next;
  bool free;
  size_t size;
  uint8_t buffer[0];
} metadata_t;

static metadata_t* mm_head;

void* mm_malloc(size_t size) {
  //TODO: Implement malloc
  // if mm_head is null 
    //call sbrk(size) --> returns start of heap (mm head can point to this)
    // set prev and next to null, 
    //free = false
    //size = size
    // 
  if (size == 0) {
    return NULL;
  }
  
  if (mm_head == NULL) {
    mm_head = sbrk(size + sizeof(metadata_t));
    if (mm_head != -1) {
      mm_head->prev = NULL;
      mm_head->next = NULL;
      mm_head->free = false;
      mm_head->size = size;
      return mm_head->buffer;
    }
  } else {
    metadata_t* block = mm_head;
    metadata_t* mm_tail;
    while(block != NULL) {
      //printf("block:%08x, free:%d, size:%d\n",block, block->free, block->size);
      if (block->free && (block->size >= size)) {
        break;
      }
      mm_tail = block;
      block = block->next;
    }

    if (block != NULL) {
      block->free = false;
      memset(block->buffer, 0, block->size);
      if ((int)(block->size - (size + sizeof(metadata_t))) > 0) {
        //printf("Splitting the block, block->size:%ld, size:%ld\n", block->size, size);
        metadata_t* new_block = block->buffer + size;
        new_block->prev = block;
        new_block->next = block->next;
        block->next = new_block;
        new_block->size = block->size - (size + sizeof(metadata_t));
        block->size = size;
        new_block->free = true;
      }
      
      return block->buffer;
    }

    block = sbrk(size + sizeof(metadata_t));
    if (block != -1) {
      block->free = false;
      block->size = size;
      block->prev = mm_tail;
      block->next = NULL;
      mm_tail->next = block;
      return block->buffer;
    }
  }
  return NULL;
}

void* mm_realloc(void* ptr, size_t size) {
  //TODO: Implement realloc
  if (ptr == NULL) {
    return mm_malloc(size);
  }
  metadata_t* block = mm_head;
  while(block != NULL) {
    if (ptr == block->buffer) {
      break;
    }
    block = block->next;
  }
  if (block == NULL) {
    return NULL;
  }
  void* new_ptr = mm_malloc(size);
  if (new_ptr != NULL) {
    if (block->size < size) {
      memcpy(new_ptr, block->buffer, block->size);
    } else {
      memcpy(new_ptr, block->buffer, size);
    }
  }
  mm_free(ptr);
  
  return new_ptr;
}

void mm_free(void* ptr) {
  //TODO: Implement free
  if(ptr == NULL) {
    return;
  }
  metadata_t* block = mm_head;
  // metadata_t* mm_tail;
  while(block != NULL) {
    if (ptr == block->buffer) {
      if (block->free) {
        break;
      }
      block->free = true;

      if(block->next && block->next->free) {
        block->size += block->next->size + sizeof(metadata_t);
        block->next = block->next->next;
        if (block->next) {
          block->next->prev = block;
        }
      }

      if(block->prev && block->prev->free) {
        block->prev->next = block->next;
        if (block->next) {
          block->next->prev = block->prev;
        }
        block->prev->size += block->size + sizeof(metadata_t);
      }
      break;
    }
    block = block->next;
  }
}
