/* Call directory-only syscalls on file objects. These should fail and return false. */

#include <syscall.h>
#include "tests/lib.h"
#include "tests/main.h"

char buf[100];

void test_main(void) {
  int fd1;
  int fd2;
  char* file_name = "a_file";

  CHECK(create(file_name, 20), "create \"%s\"", file_name);
  CHECK(mkdir("b_directory"), "mkdir \"b_directory\"");
  CHECK((fd1 = open(file_name)) > 1, "open \"%s\"", file_name);
  CHECK((fd2 = open("b_directory")) > 1, "open \"b_directory\"");
  CHECK(!isdir(fd1), "isdir \"a_file\"");
  CHECK(isdir(fd2), "isdir \"b_directory\"");
  CHECK(!chdir("a_file"), "chdir \"a_file\" (must return false)");
  CHECK(!readdir(fd1,"BCS"), "readdir \"a_file\" (must return false)");
  CHECK(read(fd2, buf, 10) == -1, "read \"b_directory\" (must return -1)");
  CHECK(write(fd2, buf, 10) == -1, "write \"b_directory\" (must return -1)");
}m