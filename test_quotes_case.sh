#!/usr/bin/env gosh
# Test script for case statement with quotes bug

echo "Testing case statement with quotes"

# Test 1: Variable with quotes should match quoted pattern
export TEST_VAR="hello world"
echo "Test 1: Testing with TEST_VAR='$TEST_VAR'"

case $TEST_VAR in
"hello world")
    echo "  -> Matched quoted pattern!"
    ;;
'hello world')
    echo "  -> Matched single quoted pattern!"
    ;;
*)
    echo "  -> No match found for: $TEST_VAR"
    ;;
esac

echo

# Test 2: Variable without quotes should match quoted pattern
export TEST_VAR2="hello world"
echo "Test 2: Testing with TEST_VAR2='$TEST_VAR2'"

case $TEST_VAR2 in
"hello world")
    echo "  -> Matched quoted pattern!"
    ;;
*)
    echo "  -> No match found for: $TEST_VAR2"
    ;;
esac

echo

# Test 3: Variable with spaces should work with patterns
export FILENAME="my file.txt"
echo "Test 3: Testing with FILENAME='$FILENAME'"

case $FILENAME in
"*.txt")
    echo "  -> Matched quoted wildcard pattern!"
    ;;
*.txt)
    echo "  -> Matched unquoted wildcard pattern!"
    ;;
*)
    echo "  -> No match found for: $FILENAME"
    ;;
esac

echo "Case statement quote tests completed!"