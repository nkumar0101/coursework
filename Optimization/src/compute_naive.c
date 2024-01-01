#include "compute.h"

// Computes the dot product of vec1 and vec2, both of size n
int dot(uint32_t n, int32_t *vec1, int32_t *vec2) {
  // TODO: implement dot product of vec1 and vec2, both of size n
  int dot_prod = 0;
  for (int i = 0; i < n; i++) {
    dot_prod += vec1[i] * vec2[i];
  }
  return dot_prod;
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

  if (result == NULL) {
    return -1;
  }

  int result_pos = 0;
  int a_pos = 0, b_pos = 0;

  for (int j = 0; j < a_matrix->rows - b_matrix->rows + 1; j++) {
    a_pos = a_matrix->cols * j;
    b_pos = 0;
    for (int i = 0; i < a_matrix->cols - b_matrix->cols + 1; i++) {
      
      result->data[result_pos] = 0;
      for (int r = 0; r < b_matrix->rows; r++) {
        result->data[result_pos] += dot(b_matrix->cols, a_matrix->data + a_pos + i, rev_b->data + b_pos);
        //printf("%d %d %d \n", a_pos, b_pos, i);
        a_pos += a_matrix->cols; // start of the next row for matrix a
        b_pos += b_matrix->cols; // start of the next row for matrix b
      } 
      
      a_pos = a_matrix->cols * j;
      b_pos = 0;
      result_pos ++;
      
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
