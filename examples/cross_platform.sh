#!/usr/bin/env gosh
# This script works identically on Windows, macOS, and Linux

# Detect platform in a way that works with gosh
PLATFORM=$(uname -s 2>/dev/null)
if [ -z "$PLATFORM" ]; then
    PLATFORM="Windows"
fi
echo "Running on: $PLATFORM"
echo "Script arguments: $@"

# Use cross-platform commands
ls -la          # Works on all platforms
echo "Files listed above"

# Environment variables work the same way
export MY_VAR="cross-platform value"
echo "MY_VAR: $MY_VAR"