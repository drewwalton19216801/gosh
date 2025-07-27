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