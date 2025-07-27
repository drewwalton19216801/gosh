#!/usr/bin/env gosh
# Test script for gosh shell

echo "Starting gosh test script..."
echo "Current directory:"
pwd

echo "Setting environment variable:"
local TEST_VAR=hello_world
echo "TEST_VAR is set to: $TEST_VAR"

echo "Creating alias:"
alias ll="ls -la"

echo "Testing echo command:"
echo "Hello from gosh shell!"

echo "Showing environment variables:"
env | grep TEST_VAR

echo "Script completed successfully!"