# -*- perl -*-
use strict;
use warnings;
use tests::tests;
use tests::random;
check_expected (IGNORE_EXIT_CODES => 1, [<<'EOF']);
(dir-not-dir) begin
(dir-not-dir) create "a_file"
(dir-not-dir) mkdir "b_directory"
(dir-not-dir) open "a_file"
(dir-not-dir) open "b_directory"
(dir-not-dir) isdir "a_file"
(dir-not-dir) isdir "b_directory"
(dir-not-dir) chdir "a_file" (must return false)
(dir-not-dir) readdir "a_file" (must return false)
(dir-not-dir) read "b_directory" (must return -1)
(dir-not-dir) write "b_directory" (must return -1)
(dir-not-dir) end
EOF
pass;