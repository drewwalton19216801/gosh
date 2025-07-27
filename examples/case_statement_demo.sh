#!/usr/bin/env gosh
# Case Statement Demo for gosh shell
# This script demonstrates various case statement features

echo "=== gosh Case Statement Demo ==="
echo

# Example 1: Basic string matching
echo "Example 1: Basic String Matching"
export OS=linux
echo "Operating System: $OS"

case $OS in
linux)
    echo "  -> Linux detected! Using apt package manager."
    ;;
darwin)
    echo "  -> macOS detected! Using brew package manager."
    ;;
windows)
    echo "  -> Windows detected! Using chocolatey package manager."
    ;;
*)
    echo "  -> Unknown OS: $OS"
    ;;
esac
echo

# Example 2: Multiple patterns with OR operator
echo "Example 2: Multiple Patterns (OR operator)"
export BROWSER=firefox
echo "Browser: $BROWSER"

case $BROWSER in
chrome|chromium|google-chrome)
    echo "  -> Chromium-based browser detected"
    ;;
firefox|mozilla)
    echo "  -> Mozilla Firefox detected"
    ;;
safari|webkit)
    echo "  -> WebKit-based browser detected"
    ;;
*)
    echo "  -> Unknown browser: $BROWSER"
    ;;
esac
echo

# Example 3: Wildcard pattern matching
echo "Example 3: Wildcard Pattern Matching"
export FILENAME=report.pdf
echo "File: $FILENAME"

case $FILENAME in
*.txt)
    echo "  -> Text file - opening with text editor"
    ;;
*.pdf)
    echo "  -> PDF file - opening with PDF viewer"
    ;;
*.jpg|*.png|*.gif)
    echo "  -> Image file - opening with image viewer"
    ;;
*.mp3|*.wav|*.flac)
    echo "  -> Audio file - opening with media player"
    ;;
*)
    echo "  -> Unknown file type: $FILENAME"
    ;;
esac
echo

# Example 4: Variable expansion in patterns
echo "Example 4: Variable Expansion in Patterns"
export EXPECTED_VERSION=1.2.3
export CURRENT_VERSION=1.2.3
echo "Expected: $EXPECTED_VERSION, Current: $CURRENT_VERSION"

case $CURRENT_VERSION in
$EXPECTED_VERSION)
    echo "  -> Version matches! System is up to date."
    ;;
*)
    echo "  -> Version mismatch! Update required."
    ;;
esac
echo

# Example 5: Numeric ranges (using patterns)
echo "Example 5: HTTP Status Code Handling"
export HTTP_STATUS=404
echo "HTTP Status: $HTTP_STATUS"

case $HTTP_STATUS in
200)
    echo "  -> Success: OK"
    ;;
3*)
    echo "  -> Redirection: $HTTP_STATUS"
    ;;
4*)
    echo "  -> Client Error: $HTTP_STATUS"
    ;;
5*)
    echo "  -> Server Error: $HTTP_STATUS"
    ;;
*)
    echo "  -> Unknown status code: $HTTP_STATUS"
    ;;
esac
echo

echo "=== Demo Complete ==="
echo "Case statements in gosh support:"
echo "  - Exact string matching"
echo "  - Multiple patterns with | (OR operator)"
echo "  - Wildcard patterns (*, ?, [abc])"
echo "  - Variable expansion in patterns"
echo "  - Default case with * pattern"