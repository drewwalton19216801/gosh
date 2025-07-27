# Case Statements in gosh

The `gosh` shell supports case statements for conditional execution based on pattern matching. Case statements provide a clean way to handle multiple conditional branches based on the value of a variable.

## Syntax

```bash
case $variable in
pattern1)
    commands...
    ;;
pattern2|pattern3)
    commands...
    ;;
*)
    default commands...
    ;;
esac
```

## Features

### Basic String Matching

```bash
export OS=linux
case $OS in
linux)
    echo "Linux system detected"
    ;;
darwin)
    echo "macOS system detected"
    ;;
windows)
    echo "Windows system detected"
    ;;
*)
    echo "Unknown operating system"
    ;;
esac
```

### Multiple Patterns (OR operator)

Use the `|` operator to match multiple patterns in a single case:

```bash
export BROWSER=firefox
case $BROWSER in
chrome|chromium|google-chrome)
    echo "Chromium-based browser"
    ;;
firefox|mozilla)
    echo "Mozilla Firefox"
    ;;
safari|webkit)
    echo "WebKit-based browser"
    ;;
*)
    echo "Unknown browser"
    ;;
esac
```

### Wildcard Pattern Matching

Case statements support shell wildcards:

- `*` - matches any string
- `?` - matches any single character
- `[abc]` - matches any character in the set
- `[a-z]` - matches any character in the range

```bash
export FILENAME=document.pdf
case $FILENAME in
*.txt)
    echo "Text file"
    ;;
*.pdf)
    echo "PDF file"
    ;;
*.jpg|*.png|*.gif)
    echo "Image file"
    ;;
*)
    echo "Unknown file type"
    ;;
esac
```

### Variable Expansion in Patterns

Patterns can include variable references that are expanded before matching:

```bash
export EXPECTED_VERSION=1.2.3
export CURRENT_VERSION=1.2.3

case $CURRENT_VERSION in
$EXPECTED_VERSION)
    echo "Version matches!"
    ;;
*)
    echo "Version mismatch!"
    ;;
esac
```

### Default Case

The `*` pattern serves as a catch-all default case and should typically be placed last:

```bash
case $VALUE in
specific_value)
    echo "Matched specific value"
    ;;
*)
    echo "No specific match found"
    ;;
esac
```

## Important Notes

1. **Multi-line support**: Case statements are supported in both script files and interactive mode. In interactive mode, the shell will automatically detect when you start a case statement and continue reading lines until you type `esac`.

2. **First match wins**: The case statement executes the commands for the first matching pattern and then exits.

3. **Pattern termination**: Each pattern block must end with `;;` (double semicolon).

4. **Case sensitivity**: Pattern matching is case-sensitive.

5. **Variable expansion**: Variables in both the test value and patterns are expanded before matching.

## Examples

See `examples/case_statement_demo.sh` for a comprehensive demonstration of case statement features.

## Comparison with Other Shells

The gosh case statement implementation is compatible with POSIX shell syntax and behaves similarly to bash and other common shells, supporting:

- Standard pattern matching syntax
- Multiple patterns with `|`
- Wildcard patterns (`*`, `?`, `[]`)
- Variable expansion
- Default case with `*`

This makes it easy to port existing shell scripts that use case statements to gosh.