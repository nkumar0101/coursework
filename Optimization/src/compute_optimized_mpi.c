#include <omp.h>
#include <x86intrin.h>

#include "compute.h"

// Computes the dot product of vec1 and vec2, both of size n
__m256i dot(uint32_t n, int32_t *vec1, int32_t *vec2)  {
  // TODO: implement dot product of vec1 and vec2, both of size n

  //__m128i temp = _mm_loadu_si128((__m128i *) (vals+i));
  // __m128i combine = _mm_and_si128(temp, _mm_cmpgt_epi32(temp, _127));
  // zero_sum_vector = _mm_add_epi32(zero_sum_vector, combine);

  int dot_prod = 0;
  __m256i sum = _mm256_setzero_si256();

  #ifdef DEBUG 
  for (int i = 0; i < n; i++){
    printf("%d ", vec1[i]);
  }
  printf("\n");

  for (int i = 0; i < n; i++){
    printf("%d ", vec2[i]);
  }
  printf("\n");

  #endif

  for (int i = 0; i < n/8*8; i+=8) {
    __m256i v1 = _mm256_loadu_si256((__m256i *) (vec1 + i));
    __m256i v2 = _mm256_loadu_si256((__m256i *) (vec2 + i));
    __m256i v3 = _mm256_mullo_epi32(v1, v2);
    sum = _mm256_add_epi32(sum, v3);
  }

  int32_t tail_1[8] = {0, 0, 0, 0, 0, 0, 0, 0};
  int32_t tail_2[8] = {0, 0, 0, 0, 0, 0, 0, 0};

  for (int i = n/8*8, j = 0; i < n; i++, j++) {
     tail_1[j] = vec1[i];
     tail_2[j] = vec2[i];
  }

  __m256i t1 = _mm256_loadu_si256((__m256i *) (tail_1));
  __m256i t2 = _mm256_loadu_si256((__m256i *) (tail_2));
  __m256i t3 = _mm256_mullo_epi32(t1, t2);
  sum = _mm256_add_epi32(sum, t3);

  return sum;
}

matrix_t* create_matrix(uint32_t rows, uint32_t cols) {
  matrix_t* new = malloc(sizeof(matrix_t));
  if (new == NULL) {
    return NULL;
  }
  //return null if malloc failed

  new->rows = rows;
  new->cols = cols;
  new->data = malloc(rows * cols * sizeof(uint32_t));

  if (new->data == NULL) {
    free(new);
    return NULL;
  }

  return new;
}

matrix_t* reverse(matrix_t* m) {

  // reverse the order of elements in m, put it in a new matrix
  // make a new matrix
  matrix_t* rev = create_matrix(m->rows, m->cols);
  
  if (rev == NULL) {
    return NULL;
  }
  uint32_t rows_cols = m->rows * m->cols;
  // starting from the last element, for loop to the first element, adding each element into the matrix

  for(int i = rows_cols - 1; i >= 0; i--) {
    rev->data[i] = m->data[rows_cols - i - 1];
  }

  return rev;
}

// Computes the convolution of two matrices
int convolve(matrix_t *a_matrix, matrix_t *b_matrix, matrix_t **output_matrix) {
  // TODO: convolve matrix a and matrix b, and store the resulting matrix in
  // output_matrix

  matrix_t* rev_b = reverse(b_matrix);
  matrix_t* result = create_matrix(a_matrix->rows - b_matrix->rows + 1, a_matrix->cols - b_matrix->cols + 1);

  //printf("a: %d %d b: %d %d\n", a_matrix->rows, a_matrix->cols, b_matrix->rows, b_matrix->cols);
  if (result == NULL) {
    return -1;
  }

  
  #pragma omp parallel for
  for (int j = 0; j < a_matrix->rows - b_matrix->rows + 1; j++) {
    #pragma omp parallel for
    for (int i = 0; i < a_matrix->cols - b_matrix->cols + 1; i++) {
      int result_pos = j * (a_matrix->cols - b_matrix->cols + 1) + i;
      __m256i result_data = _mm256_setzero_si256();
      //#pragma omp parallel for
      for (int r = 0; r < b_matrix->rows; r++) {
        int a_pos = a_matrix->cols * (r + j);
        int b_pos = b_matrix->cols * (r);
        __m256i dot_val = dot(b_matrix->cols, a_matrix->data + a_pos + i, rev_b->data + b_pos);
        //#pragma omp critical
        result_data = _mm256_add_epi32(result_data, dot_val);
        //printf("%d %d %d \n", a_pos, b_pos, i);
      } 
      
      int32_t tmp_arr[8];
      _mm256_storeu_si256((__m256i *) tmp_arr, result_data);


      result->data[result_pos] = tmp_arr[0] + tmp_arr[1] + tmp_arr[2] +tmp_arr[3] + tmp_arr[4] + tmp_arr[5] + tmp_arr[6] + tmp_arr[7];
      //printf("%d %d %d %d %d %d\n", i, j, a_pos, b_pos, result_pos, result_data);
    }
    
  }
  
  *output_matrix = result;
  free(rev_b->data);
  free(rev_b);
  //free reverse matrix->data, and reversed matrix
  return 0;
}

// Executes a task
int execute_task(task_t *task) {
  matrix_t *a_matrix, *b_matrix, *output_matrix;

  if (read_matrix(get_a_matrix_path(task), &a_matrix))
    return -1;
  if (read_matrix(get_b_matrix_path(task), &b_matrix))
    return -1;

  if (convolve(a_matrix, b_matrix, &output_matrix))
    return -1;

  if (write_matrix(get_output_matrix_path(task), output_matrix))
    return -1;

  free(a_matrix->data);
  free(b_matrix->data);
  free(output_matrix->data);
  free(a_matrix);
  free(b_matrix);
  free(output_matrix);
  return 0;
}
