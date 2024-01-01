/* Tests functionality of our directory path finding function. */

#include <syscall.h>
#include "tests/lib.h"
#include "tests/main.h"
#include "tests/filesys/extended/mk-tree.h"

void test_main(void) {
    make_tree(4, 3, 2, 1);
    CHECK(create("/0/1/1/a", 512), "create \"0/1/1/a\"");
    CHECK(!chdir("0/0/0/0/0/0"), "chdir into non-existent directory should fail.");
    CHECK(chdir("/0/2/"), "chdir into existing directory should pass.");
    CHECK(open("../../././0/../0/./../0/1/././1/./../1/a") > 1, "open \"../../././0/../0/./../0/1/././1/./../1/a\"");
}