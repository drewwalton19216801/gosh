#!/usr/local/bin/gosh

test_func() {
    echo "Function received: $1"
}

echo "Script received: $1"
test_func $1