# -*- perl -*-
use strict;
use warnings;
use tests::tests;
use tests::random;
check_expected (IGNORE_EXIT_CODES => 1, [<<'EOF']);
(dir-mania) begin
(dir-mania) creating /0/0/0/0 through /3/2/1/0...
(dir-mania) open "/0/2/0/0"
(dir-mania) close "/0/2/0/0"
(dir-mania) create "0/1/1/a"
(dir-mania) chdir into non-existent directory should fail.
(dir-mania) chdir into existing directory should pass.
(dir-mania) open "../../././0/../0/./../0/1/././1/./../1/a"
(dir-mania) end
EOF
pass;