# Scripting Guide

This guide covers how to write shell scripts using gosh, including syntax, features, and best practices.

## Getting Started

### Script Mode
Run gosh scripts by passing the script file as an argument:
```bash
./gosh script.sh
```

### Shebang Line
Start your scripts with a shebang line to make them executable:
```bash
#!/usr/bin/env gosh
```

Then make the script executable:
```bash
chmod +x script.sh
./script.sh
```

## Basic Script Structure

### Example Script
```bash
#!/usr/bin/env gosh
# Example gosh script - gosh, this is neat!

echo "Hello from gosh!"
pwd
export MY_VAR=test
echo "MY_VAR is: $MY_VAR"

# Pipe examples - gosh, pipes are powerful!
echo "testing pipe functionality" | wc -w
ls -la | head -5
cat /etc/passwd | grep root | wc -l

# Gosh darn it, that's some fine scripting!
```

## Variable Expansion

Gosh supports several forms of variable expansion:

### Basic Variable Expansion
```bash
export NAME="World"
echo "Hello, $NAME!"
echo "Path: ${HOME}/Documents"
```

### Environment Variables
```bash
# Set environment variables
export MY_VAR="Hello"
export PATH="$PATH:/new/path"

# Use environment variables
echo $MY_VAR
echo "Current PATH: $PATH"
```

## Command Substitution

Execute commands and use their output in your scripts:

### Using `$(command)` Syntax
```bash
echo "Current directory: $(pwd)"
echo "Today is $(date +%Y-%m-%d)"
echo "Found $(ls *.txt | wc -l) text files in $(pwd)"
```

### Using Backticks
```bash
echo "Files: `ls | wc -l`"
echo "Current user: `whoami`"
```

## Tilde Expansion

Gosh expands tilde (`~`) to represent home directories:

```bash
# Expand to current user's home
ls ~/Documents
cp file.txt ~/backup/
cd ~

# Expand to specific user's home
cd ~user/projects
ls ~admin/logs
```

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

# Set variables
export BACKUP_DIR="$HOME/backups"
export DATE=$(date +%Y%m%d)
export BACKUP_FILE="backup_$DATE.tar.gz"

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