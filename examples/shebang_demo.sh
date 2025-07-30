#!/usr/bin/env gosh
# Shebang demonstration script
# This script shows how shebang support works in gosh

echo "=== Gosh Shebang Demonstration ==="
echo "This script has shebang: #!/usr/bin/env gosh"
echo ""

echo "Platform detection:"
if command -v uname >/dev/null 2>&1; then
    local PLATFORM=$(uname -s)
    echo "Running on Unix-like system: $PLATFORM"
    echo "On this platform, gosh respects shebangs and would run this script with gosh."
    echo "Scripts with #!/bin/bash would run with bash, #!/bin/zsh with zsh, etc."
    echo "Scripts without shebangs get a warning: 'By gosh, I'll run it with gosh and hope for the best!'"
else
    echo "Running on Windows"
    echo "On Windows, all scripts run through gosh regardless of shebang."
    echo "Shebangs are recognized but not enforced (Windows doesn't support them natively)."
fi

echo ""
echo "Script arguments: $@"
echo "Current working directory: $(pwd)"

echo ""
echo "=== Testing some gosh features ==="

# Test file operations
if [ -f "README.md" ]; then
    echo "✓ Found README.md - file test works"
else
    echo "✗ README.md not found"
fi

# Test variables
export TEST_VAR="Hello from gosh!"
echo "✓ Environment variable: $TEST_VAR"

local LOCAL_VAR="Local value"
echo "✓ Local variable: $LOCAL_VAR"

# Test command substitution
echo "✓ Command substitution: Current time is $(date)"

echo ""
echo "Gosh, that demonstrates shebang support nicely!"