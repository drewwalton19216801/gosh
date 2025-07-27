#!/usr/bin/env gosh

# Test script for function support in gosh

# Define a simple function
greet() {
    echo "Hello, $1!"
    echo "You passed $# arguments"
}

# Define a function that uses multiple parameters
show_params() {
    echo "Function name: $0"
    echo "First parameter: $1"
    echo "Second parameter: $2"
    echo "All parameters: $@"
    echo "Parameter count: $#"
}

# Define a function that calls another function
welcome() {
    echo "Welcome to gosh!"
    greet "$1"
    echo "Function name: $0"
}

# Test the functions
echo "Testing function support:"
echo

echo "1. Simple greeting:"
greet "World"
echo

echo "2. Parameter display:"
show_params "Alice" "Bob" "Charlie"
echo

echo "3. Nested function call:"
welcome "User"
echo

echo "4. Function with multiple arguments:"
greet "Alice" "Bob" "Charlie"
echo

echo "5. List all functions:"
declare -f
echo

echo "6. Show function definitions:"
declare
echo

echo "Function tests completed!"