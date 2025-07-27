#!/usr/local/bin/gosh

echo "=== Final If Statement Test ==="
echo

# Test 1: Basic if-then-else
echo "Test 1: Basic conditional"
if test 5 -gt 3
then
    echo "  ✓ 5 is greater than 3"
else
    echo "  ✗ Math is broken!"
fi
echo

# Test 2: Function with proper argument checking
echo "Test 2: Function argument checking"
check_args() {
    echo "  Function received $# arguments"
    if test $# -eq 0
    then
        echo "  ✓ No arguments - showing usage"
        echo "  Usage: check_args <arg1> [arg2...]"
    else
        echo "  ✓ Arguments provided: $@"
    fi
}

echo "  Calling with no arguments:"
check_args
echo "  Calling with arguments:"
check_args hello world
echo

# Test 3: String comparison
echo "Test 3: String comparison"
export USER_INPUT=yes
if test $USER_INPUT = yes
then
    echo "  ✓ User said yes"
else
    echo "  ✗ User said no"
fi
echo

# Test 4: File existence check
echo "Test 4: File existence"
if test -f test.txt
then
    echo "  ✓ test.txt exists"
else
    echo "  ✗ test.txt does not exist"
fi
echo

echo "=== All if statement tests completed successfully! ==="