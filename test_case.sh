#!/usr/bin/env gosh
# Test script for case statement functionality

echo Testing case statement in gosh

# Test 1: Basic case statement with string matching
export TEST_VAR=apple
echo Test 1: Testing with TEST_VAR=$TEST_VAR

case $TEST_VAR in
apple)
    echo Found an apple!
    echo Apples are red or green
    ;;
banana)
    echo Found a banana!
    echo Bananas are yellow
    ;;
orange)
    echo Found an orange!
    echo Oranges are orange
    ;;
*)
    echo Unknown fruit: $TEST_VAR
    ;;
esac

echo

# Test 2: Case statement with multiple patterns
export COLOR=red
echo Test 2: Testing with COLOR=$COLOR

case $COLOR in
red|crimson|scarlet)
    echo That is a shade of red!
    ;;
blue|navy|azure)
    echo That is a shade of blue!
    ;;
green|emerald|lime)
    echo That is a shade of green!
    ;;
*)
    echo Unknown color: $COLOR
    ;;
esac

echo

# Test 3: Case statement with no match
export NUMBER=42
echo Test 3: Testing with NUMBER=$NUMBER

case $NUMBER in
1)
    echo One
    ;;
2)
    echo Two
    ;;
3)
    echo Three
    ;;
esac

echo Case statement tests completed!