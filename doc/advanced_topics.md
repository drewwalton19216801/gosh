# Advanced Topics

This document explores advanced features and techniques for power users of gosh.

## Architecture Overview

Gosh is organized into multiple modules for maximum readability and maintainability:

- `main.go` - Entry point and argument handling
- `shell.go` - Core shell structure and main loop
- `parser.go` - Command line parsing and tokenization
- `executor.go` - Command execution logic
- `expansion.go` - Variable and command expansion
- `builtins.go` - Built-in command implementations
- `utils.go` - Utility functions

## Advanced Variable Expansion

### Complex Variable Substitution
```bash
# Nested variable expansion with proper declarations
export BASE_DIR="/home/user"          # Environment variable
local PROJECT_DIR="${BASE_DIR}/projects" # Local variable
export CURRENT_PROJECT="${PROJECT_DIR}/gosh" # Environment variable

echo "Working in: $CURRENT_PROJECT"
```

### Variable Expansion in Command Substitution
```bash
# Proper variable declarations
local SEARCH_TERM="error"              # Local search term
export LOG_FILE="/var/log/system.log"  # Environment variable for log file

# Use variables within command substitution
local ERROR_COUNT="$(grep "$SEARCH_TERM" "$LOG_FILE" | wc -l)"
echo "Found $ERROR_COUNT occurrences"
```

## Advanced Glob Patterns

### Character Classes
```bash
# Match files starting with uppercase letters
ls [A-Z]*

# Match files with specific extensions
ls *.[ch]  # .c or .h files

# Exclude patterns (when supported by underlying system)
ls !(*.tmp)  # All files except .tmp files
```

### Complex Glob Combinations
```bash
# Multiple patterns
ls *.{go,md,txt}

# Nested directory patterns
find . -name "*.go" -type f
```

## Line Continuation

Gosh supports line continuation using backslashes (`\`) at the end of lines, allowing you to split long commands across multiple lines for better readability.

### Basic Line Continuation
```bash
# Split a long echo command
echo "This is a very long message" \
  "that spans multiple lines" \
  "for better readability"

# Long command with many arguments
ls -la \
  *.go \
  *.md \
  *.txt
```

### Line Continuation in Pipelines
```bash
# Multi-stage pipeline
cat large_file.txt | \
  grep "pattern" | \
  sort | \
  uniq -c | \
  head -20
```

### Line Continuation in Command Chains
```bash
# Sequential commands
echo "Step 1: Preparing..."; \
  mkdir -p output; \
  echo "Step 2: Processing..."; \
  touch output/result.txt; \
  echo "Step 3: Complete!"
```

### Interactive vs Script Usage
Line continuation works in both interactive mode and script files:
- **Interactive**: Use backslash at end of line, press Enter, and continue on the next line with a `> ` prompt
- **Scripts**: Use backslash at end of line and continue the command on subsequent lines

### Best Practices
- Use line continuation to improve readability of complex commands
- Indent continued lines for visual clarity
- Avoid unnecessary spaces after the backslash
- Be consistent with your continuation style within scripts

## Advanced Piping Techniques

### Multi-stage Data Processing
```bash
# Complex data pipeline with line continuation
cat data.csv | \
  grep -v "^#" | \
  cut -d',' -f2,3 | \
  sort | \
  uniq -c | \
  sort -nr | \
  head -10
```

### Combining Pipes with Redirection
```bash
# Process data and save both intermediate and final results
cat input.txt | \
  tee intermediate.txt | \
  sort | \
  tee sorted.txt | \
  uniq > final.txt
```

## Advanced Command Chaining

### Conditional Execution Patterns
```bash
# Sequential execution with status checking
command1; echo "Command 1 completed"; command2; echo "Command 2 completed"

# Complex workflows with line continuation
echo "Starting process..."; \
  mkdir -p temp; \
  cd temp; \
  echo "Working in $(pwd)"; \
  touch file.txt; \
  ls -la; \
  cd ..; \
  echo "Process completed"
```

## Environment Management

### Variable Declaration Requirements
Gosh enforces explicit variable declarations to improve script clarity and prevent accidental variable creation:

```bash
# ✅ Correct: Use explicit declarations
local TEMP_VAR="temporary"     # Local to current shell
export GLOBAL_VAR="global"     # Available to child processes

# ❌ Error: Bare assignments are not allowed
# TEMP_VAR="temporary"  # This will cause an error
```

### Advanced Environment Variable Techniques
```bash
# Save and restore environment with proper declarations
local OLD_PATH="$PATH"          # Save current PATH locally
export PATH="/custom/bin:$PATH" # Modify PATH for child processes
# ... do work with custom PATH ...
export PATH="$OLD_PATH"         # Restore original PATH
unset OLD_PATH                   # Clean up
```

### Dynamic Environment Setup
```bash
# Set environment based on conditions with proper declarations
export ENVIRONMENT="development"                    # Environment variable
local CONFIG_FILE="config_${ENVIRONMENT}.json"     # Local variable
export LOG_LEVEL="debug"                           # Environment variable

echo "Using config: $CONFIG_FILE"
echo "Log level: $LOG_LEVEL"
```

## Advanced Aliasing

### Complex Aliases
```bash
# Aliases with multiple commands
alias backup="tar -czf backup_$(date +%Y%m%d).tar.gz"
alias logcheck="tail -f /var/log/system.log | grep ERROR"

# Aliases with pipes
alias psgrep="ps aux | grep"
alias netstat="netstat -tuln | grep LISTEN"
```

### Temporary Aliases
```bash
# Create temporary aliases for a session
alias temp_alias="echo 'This is temporary'"
temp_alias
unalias temp_alias
```

## Advanced Scripting Patterns

### Error Handling Patterns
```bash
#!/usr/bin/env gosh

# Function-like command grouping
echo "Starting backup process..."

# Check if source directory exists
if [ -d "$SOURCE_DIR" ]; then
    echo "Source directory found: $SOURCE_DIR"
else
    echo "Error: Source directory not found: $SOURCE_DIR"
    exit 1
fi

# Perform backup with status checking
tar -czf "backup.tar.gz" "$SOURCE_DIR"
echo "Backup completed with status: $?"
```

### Configuration Management
```bash
#!/usr/bin/env gosh
# Configuration-driven script

# Load configuration
export CONFIG_FILE="${HOME}/.gosh_config"
if [ -f "$CONFIG_FILE" ]; then
    # Source configuration (if it were supported)
    echo "Loading configuration from $CONFIG_FILE"
else
    echo "Using default configuration"
    export DEFAULT_EDITOR="nano"
    export DEFAULT_PAGER="less"
fi
```

## Performance Optimization

### Efficient Command Patterns
```bash
# Avoid unnecessary command substitution
# Instead of: echo "Files: $(ls | wc -l)"
# Use when possible: ls | wc -l

# Minimize pipe stages
# Instead of: cat file | grep pattern | wc -l
# Use: grep pattern file | wc -l
```

### Memory-Efficient Processing
```bash
# Process large files efficiently
# Stream processing instead of loading entire files
tail -f large_log.txt | grep "ERROR" | head -100

# Use appropriate tools for the job
grep "pattern" large_file.txt > matches.txt  # Better than cat | grep
```

## Integration with External Tools

### Working with JSON Data
```bash
# Process JSON with external tools (when available)
echo '{"name":"gosh","version":"1.0"}' | \
  python3 -c "import json,sys; print(json.load(sys.stdin)['name'])"
```

### System Integration
```bash
# System monitoring scripts
echo "System Status Report - $(date)"
echo "=================================="
echo "Disk Usage:"
df -h | head -5
echo ""
echo "Memory Usage:"
free -h 2>/dev/null || vm_stat
echo ""
echo "Load Average:"
uptime
```

## Debugging and Troubleshooting

### Debug Mode Techniques
```bash
#!/usr/bin/env gosh
# Enable debug output
export DEBUG=1

if [ "$DEBUG" = "1" ]; then
    echo "Debug: Starting script execution"
fi

# Trace command execution
echo "Executing: ls -la"
ls -la

if [ "$DEBUG" = "1" ]; then
    echo "Debug: Command completed"
fi
```

### Error Logging
```bash
# Log errors to file
export ERROR_LOG="/tmp/gosh_errors.log"

# Redirect errors
command_that_might_fail 2>> "$ERROR_LOG"

# Check for errors
if [ -s "$ERROR_LOG" ]; then
    echo "Errors occurred, check $ERROR_LOG"
fi
```

## Best Practices for Advanced Usage

1. **Modular Scripts**: Break complex scripts into smaller, reusable components
2. **Error Handling**: Always consider what happens when commands fail
3. **Resource Management**: Be mindful of memory and CPU usage in scripts
4. **Security**: Never expose sensitive data in command lines or environment variables
5. **Portability**: Write scripts that work across different systems when possible
6. **Documentation**: Comment complex logic and provide usage examples
7. **Testing**: Test scripts with various inputs and edge cases

## Platform-Specific Considerations

### macOS Compatibility
Gosh is designed and tested for macOS, taking advantage of:
- BSD-style command utilities
- macOS-specific paths and conventions
- Native macOS terminal features

### Cross-Platform Scripting
```bash
# Detect operating system
export OS_TYPE=$(uname -s)
case "$OS_TYPE" in
    "Darwin")
        echo "Running on macOS"
        export OPEN_CMD="open"
        ;;
    "Linux")
        echo "Running on Linux"
        export OPEN_CMD="xdg-open"
        ;;
    *)
        echo "Unknown operating system: $OS_TYPE"
        ;;
esac
```

## Future Enhancements

Areas for potential future development:
- Enhanced tab completion
- Job control improvements
- Additional built-in commands
- Configuration file support
- Plugin system
- Enhanced error reporting

Gosh continues to evolve while maintaining its core philosophy of simplicity and reliability!