# Scripting Guide

This guide covers how to write shell scripts using gosh, including syntax, features, and best practices.

## Getting Started with Gosh Scripts

### Creating Your First Script

Gosh scripts are text files containing shell commands. They work seamlessly across all platforms, including Windows.

#### Basic Script Structure
```bash
#!/usr/bin/env gosh
# This is a comment

echo "Hello from gosh!"
echo "Current directory: $(pwd)"
```

#### Platform Compatibility
Gosh provides excellent cross-platform script execution:

- **Windows**: Scripts with `.sh` extension or shell shebangs are automatically detected and executed through gosh
- **macOS/Linux**: Scripts work natively with standard Unix conventions
- **Cross-platform commands**: Use gosh's dual command support for maximum compatibility

#### Running Scripts

**On all platforms:**
```bash
# Make script executable (Unix/macOS)
chmod +x myscript.sh
./myscript.sh

# Or run directly with gosh
gosh myscript.sh

# Command mode
gosh -c "myscript.sh arg1 arg2"
```

**Windows-specific notes:**
- No need to make scripts executable
- Scripts are automatically detected by extension (`.sh`) or shebang
- Both forward slashes and backslashes work in paths

#### Shebang Lines
Gosh recognizes various shebang formats and handles them differently based on the platform:

**On Unix/Linux/macOS:**
- Shebangs are respected and scripts are executed with the specified interpreter
- If the interpreter is not found, gosh returns an error
- Scripts without shebangs show a warning and run in gosh

**On Windows:**
- Shebangs are recognized but not enforced (Windows doesn't support shebangs natively)
- All scripts run through gosh regardless of shebang

**Supported Shebang Formats:**
```bash
#!/usr/bin/env gosh    # Recommended for cross-platform gosh scripts
#!/bin/gosh            # Direct path (if gosh is in /bin)
#!/bin/bash            # Runs with bash on Unix/Linux/macOS
#!/usr/bin/env bash    # Uses env to find bash in PATH
#!/bin/zsh             # Runs with zsh on Unix/Linux/macOS
#!/usr/bin/env zsh     # Uses env to find zsh in PATH
#!/bin/sh              # Runs with sh on Unix/Linux/macOS
```

**Examples:**
```bash
# This script will run with bash on Unix/Linux/macOS, gosh on Windows
#!/bin/bash
echo "Running in: $0"

# This script will run with gosh on all platforms
#!/usr/bin/env gosh
echo "Running in gosh!"

# This script will show a warning on Unix/Linux/macOS, then run with gosh
# No shebang line
echo "No shebang - gosh will show a warning and run anyway!"
```

## Basic Script Structure

### Example Script
```bash
#!/usr/bin/env gosh
# Example gosh script - gosh, this is neat!

echo "Hello from gosh!"
pwd

# Proper variable declarations
export MY_VAR="test"  # Environment variable
local TEMP_COUNT=42   # Local variable

echo "MY_VAR is: $MY_VAR"
echo "Temp count: $TEMP_COUNT"

# Pipe examples - gosh, pipes are powerful!
echo "testing pipe functionality" | wc -w
ls -la | head -5
cat /etc/passwd | grep root | wc -l

# Gosh darn it, that's some fine scripting!
```

## Variable Declarations

Gosh requires explicit variable declarations using either `local` or `export` commands. Bare variable assignments (e.g., `VAR=value`) are not allowed.

### Local Variables
Use `local` for variables that should only exist within the current shell session:
```bash
# Local variables (shell-only)
local TEMP_DIR="/tmp/myapp"
local COUNT=42
local RESULT="$(date +%Y%m%d)"

echo "Temp directory: $TEMP_DIR"
echo "Count: $COUNT"
```

### Environment Variables
Use `export` for variables that should be available to child processes:
```bash
# Environment variables (inherited by child processes)
export MY_VAR="Hello"
export PATH="$PATH:/new/path"
export BUILD_ENV="production"

echo "MY_VAR: $MY_VAR"
echo "Current PATH: $PATH"
```

### Key Differences
- **`local`**: Variables exist only in the current shell session
- **`export`**: Variables are passed to child processes and subshells
- **Error**: Bare assignments like `VAR=value` will result in an error

## Variable Expansion

Gosh supports several forms of variable expansion:

### Basic Variable Expansion
```bash
export NAME="World"
local GREETING="Hello"
echo "$GREETING, $NAME!"
echo "Path: ${HOME}/Documents"
```

## Command Substitution

Execute commands and use their output in your scripts:

### Using `$(command)` Syntax
```bash
# Direct command substitution in output
echo "Current directory: $(pwd)"
echo "Today is $(date +%Y-%m-%d)"
echo "Found $(ls *.txt | wc -l) text files in $(pwd)"

# Storing command output in variables
local CURRENT_DIR="$(pwd)"
export BUILD_DATE="$(date +%Y-%m-%d)"
local FILE_COUNT="$(ls *.txt | wc -l)"

echo "Working in: $CURRENT_DIR"
echo "Build date: $BUILD_DATE"
echo "Text files: $FILE_COUNT"
```

### Using Backticks
```bash
# Direct command substitution in output
echo "Files: `ls | wc -l`"
echo "Current user: `whoami`"

# Storing command output in variables
local FILE_COUNT="`ls | wc -l`"
export CURRENT_USER="`whoami`"

echo "Total files: $FILE_COUNT"
echo "Running as: $CURRENT_USER"
```

## Tilde Expansion

Gosh expands tilde (`~`) to represent home directories with cross-platform path separator support:

```bash
# Expand to current user's home
ls ~/Documents
cp file.txt ~/backup/
cd ~

# Windows-style paths with backslashes (Windows only)
cd ~\Documents
ls ~\backup\
copy file.txt ~\backup\

# Expand to specific user's home
cd ~user/projects
ls ~admin/logs

# Windows-style user paths
cd ~user\projects
ls ~admin\logs
```

**Cross-Platform Compatibility:**
- Both `~/path` and `~\path` work on Windows
- Unix/Linux systems use forward slashes: `~/path`
- Windows systems support both: `~/path` and `~\path`
- Path separator style is preserved in expansions

## Glob Patterns

Use wildcards for filename matching:

```bash
# Match all .go files
ls *.go

# Match single character
rm temp_file_?.txt

# Character classes
echo [Hh]ello*
ls [abc]*
```

## Input/Output Redirection

### Input Redirection
```bash
# Read from file
cat < input.txt
sort < unsorted.txt
```

### Output Redirection
```bash
# Overwrite file
echo "Hello" > output.txt
ls -la > file_list.txt

# Append to file
echo "World" >> output.txt
date >> log.txt
```

### Combining Redirection
```bash
# Sort a file and save result
sort < input.txt > sorted.txt

# Pipe with redirection
ls | grep ".go" > go_files.txt
```

## Pipes and Command Chaining

### Using Pipes
```bash
# Chain commands with pipes
ls | grep ".go" | wc -l
cat file.txt | sort | uniq
ps aux | grep gosh | head -5
```

### Command Chaining with Semicolons
```bash
# Execute commands sequentially
echo "First command"; echo "Second command"; pwd
ls; echo "Files listed"; pwd

# Combine with pipes and redirection
ls *.go > go_files.txt; cat go_files.txt | wc -l
```

## Background Execution

Run commands in the background:

```bash
# Run command in background
sleep 10 &
long_running_process &

# Continue with other commands
echo "Background job started"
ps aux | grep sleep
```

## Comments

Use comments to document your scripts:

```bash
#!/usr/bin/env gosh

# This is a comment
echo "This is not a comment"  # Inline comment

# Multi-line comments
# can span multiple lines
# like this
```

## Functions

Gosh supports shell functions for code reusability and organization:

### Function Definition
```bash
# Basic function syntax
function_name() {
    # function body
    echo "Hello from function"
}

# Alternative syntax
function function_name {
    echo "Hello from function"
}
```

### Calling Functions
```bash
# Define a function
greet() {
    echo "Hello, $1!"
}

# Call the function
greet "World"  # Output: Hello, World!
greet "Gosh"   # Output: Hello, Gosh!
```

### Positional Parameters
Functions can access arguments through positional parameters:

```bash
show_params() {
    echo "Function name: $0"
    echo "First argument: $1"
    echo "Second argument: $2"
    echo "Number of arguments: $#"
    echo "All arguments: $@"
    echo "All arguments as string: $*"
}

show_params arg1 arg2 arg3
```

### Practical Examples

#### File Processing Function
```bash
backup_file() {
    if [ $# -eq 0 ]; then
        echo "Usage: backup_file <filename>"
        return 1
    fi
    
    # Local variables (function scope)
    local file="$1"
    local backup_name="${file}.backup.$(date +%Y%m%d)"
    
    cp "$file" "$backup_name"
    echo "Backed up $file to $backup_name"
}

# Usage
backup_file important.txt
```

#### System Information Function
```bash
system_info() {
    echo "=== System Information ==="
    echo "Current directory: $(pwd)"
    echo "Current user: $(whoami)"
    echo "Date: $(date)"
    echo "Disk usage: $(df -h . | tail -1 | awk '{print $5}')"
}

system_info
```

### Nested Function Calls
```bash
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

process_file() {
    local file="$1"
    log_message "Processing file: $file"
    
    if [ -f "$file" ]; then
        log_message "File exists, processing..."
        wc -l "$file"
        log_message "File processed successfully"
    else
        log_message "Error: File not found"
    fi
}

process_file "example.txt"
```

### Function Management
```bash
# List all defined functions
functions

# Check if something is a function
type greet
type ls  # Shows it's not a function
```

### Complete Example: Project Management Script
```bash
#!/usr/bin/env gosh
# Project management utilities

# Initialize project structure
init_project() {
    local project_name="$1"
    if [ -z "$project_name" ]; then
        echo "Usage: init_project <project_name>"
        return 1
    fi
    
    mkdir -p "$project_name"/{src,docs,tests}
    echo "# $project_name" > "$project_name/README.md"
    echo "Project $project_name initialized!"
}

# Count lines of code
count_code() {
    echo "Lines of code in project:"
    find . -name "*.go" -o -name "*.sh" | xargs wc -l | tail -1
}

# Clean temporary files
clean_temp() {
    echo "Cleaning temporary files..."
    find . -name "*.tmp" -o -name "*.log" | xargs rm -f
    echo "Cleanup complete!"
}

# Usage examples
init_project "my-gosh-project"
count_code
clean_temp
```

## Aliases in Scripts

Create and use aliases within scripts:

```bash
# Create aliases
alias ll="ls -la"
alias grep="grep --color=auto"

# Use aliases
ll
ps aux | grep gosh

# Remove aliases
unalias ll
```

## Best Practices

1. **Use shebang lines** for executable scripts
2. **Add comments** to explain complex operations
3. **Quote variables** to handle spaces: `"$MY_VAR"`
4. **Use meaningful variable names**
5. **Test scripts** before deploying
6. **Handle errors** appropriately
7. **Use consistent indentation** for readability

## Example: Complete Script

```bash
#!/usr/bin/env gosh
# File backup script

# Set variables with proper declarations
export BACKUP_DIR="$HOME/backups"           # Environment variable
local DATE="$(date +%Y%m%d)"                # Local variable
export BACKUP_FILE="backup_$DATE.tar.gz"    # Environment variable

# Create backup directory if it doesn't exist
echo "Creating backup directory: $BACKUP_DIR"
mkdir -p "$BACKUP_DIR"

# Create backup
echo "Creating backup: $BACKUP_FILE"
tar -czf "$BACKUP_DIR/$BACKUP_FILE" *.txt *.md

# List backup contents
echo "Backup created successfully!"
ls -la "$BACKUP_DIR" | grep "$DATE"

# Show backup size
echo "Backup size: $(du -h "$BACKUP_DIR/$BACKUP_FILE" | cut -f1)"
```

This script demonstrates many gosh features including variable expansion, command substitution, pipes, and redirection in a practical example.