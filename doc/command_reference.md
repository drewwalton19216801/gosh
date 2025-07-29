# Command Reference

This document provides a comprehensive reference for all built-in commands available in gosh.

## Built-in Commands

### Core Shell Commands

### `exit [code]`
Exit the shell with optional exit code.

**Usage:**
```bash
exit          # Exit with code 0
exit 1        # Exit with code 1
```

### `cd [dir]`
Change directory. Defaults to home directory if no argument provided.

**Usage:**
```bash
cd            # Change to home directory
cd /path/to/dir  # Change to specific directory
cd ..         # Change to parent directory
```

### `pwd`
Print the current working directory.

**Usage:**
```bash
pwd
```

### `echo [args...]`
Print arguments to standard output.

**Usage:**
```bash
echo "Hello World"
echo $MY_VAR
echo "Multiple" "arguments"
```

## Cross-Platform File Operations

### `ls [options] [path...]` / `dir [options] [path...]`
List directory contents. Both Unix-style (`ls`) and Windows-style (`dir`) commands are supported.

**Options:**
- `-l` - Long format (detailed listing with permissions, size, date)
- `-a` - Show all files including hidden files (starting with .)
- `-h` - Human-readable file sizes (with -l option)

**Usage:**
```bash
ls                    # List current directory
ls -l                 # Long format listing
ls -la                # Long format with hidden files
ls -lah               # Long format, hidden files, human-readable sizes
dir                   # Windows-style listing
dir -l /path/to/dir   # Long format for specific directory
```

### `cat [file...]` / `type [file...]`
Display file contents. Both Unix-style (`cat`) and Windows-style (`type`) commands are supported.

**Usage:**
```bash
cat file.txt          # Display file contents
cat file1.txt file2.txt  # Display multiple files
type document.txt     # Windows-style file display
```

### `cp [options] source dest` / `copy [options] source dest`
Copy files and directories. Both Unix-style (`cp`) and Windows-style (`copy`) commands are supported.

**Options:**
- `-r` - Recursive copy (for directories)

**Usage:**
```bash
cp file.txt backup.txt        # Copy file
cp -r dir1 dir2              # Copy directory recursively
copy document.txt backup/    # Windows-style copy to directory
```

### `mv source dest` / `move source dest`
Move or rename files and directories. Both Unix-style (`mv`) and Windows-style (`move`) commands are supported.

**Usage:**
```bash
mv oldname.txt newname.txt   # Rename file
mv file.txt /path/to/dest/   # Move file to directory
move folder1 folder2         # Windows-style move/rename
```

### `rm [options] file...` / `del [options] file...`
Remove files and directories. Both Unix-style (`rm`) and Windows-style (`del`) commands are supported.

**Options:**
- `-r` - Recursive removal (for directories)
- `-f` - Force removal (ignore nonexistent files)

**Usage:**
```bash
rm file.txt              # Remove file
rm -r directory          # Remove directory recursively
rm -rf temp/             # Force remove directory
del document.txt         # Windows-style file removal
```

### `mkdir [options] dir...` / `md [options] dir...`
Create directories. Both Unix-style (`mkdir`) and Windows-style (`md`) commands are supported.

**Options:**
- `-p` - Create parent directories as needed

**Usage:**
```bash
mkdir newdir             # Create directory
mkdir -p path/to/newdir  # Create with parent directories
md folder                # Windows-style directory creation
```

### `rmdir dir...` / `rd dir...`
Remove empty directories. Both Unix-style (`rmdir`) and Windows-style (`rd`) commands are supported.

**Usage:**
```bash
rmdir emptydir           # Remove empty directory
rd folder                # Windows-style directory removal
```

### `clear` / `cls`
Clear the terminal screen. Both Unix-style (`clear`) and Windows-style (`cls`) commands are supported.

**Usage:**
```bash
clear                    # Unix-style screen clear
cls                      # Windows-style screen clear
```

## Environment and Shell Management

### `env`
Show all environment variables.

**Usage:**
```bash
env
```

### `export VAR=value`
Set an environment variable that will be available to child processes.

**Usage:**
```bash
export MY_VAR="Hello"
export PATH="$PATH:/new/path"
```

### `local VAR=value`
Set a local variable that exists only within the current shell session (not exported to child processes).

**Usage:**
```bash
local TEMP_VAR="temporary value"
local COUNT=42
local RESULT="$(date +%Y%m%d)"
```

### `unset VAR`
Unset (remove) an environment variable.

**Usage:**
```bash
unset MY_VAR
```

### `alias name=command`
Create a command alias.

**Usage:**
```bash
alias ll="ls -la"
alias grep="grep --color=auto"
```

### `unalias name`
Remove an alias.

**Usage:**
```bash
unalias ll
```

### `history`
Show command history.

**Usage:**
```bash
history
```

### `which cmd`
Show the path to a command.

**Usage:**
```bash
which ls
which python
```

### `help`
Show help information.

**Usage:**
```bash
help
```

### `functions`
List all defined functions.

**Usage:**
```bash
functions
```

### `type name`
Show information about a command, function, or alias.

**Usage:**
```bash
type greet      # Check if greet is a function
type ls         # Show type of ls command
type ll         # Check if ll is an alias
```

## Command Features

### Command Chaining
Execute multiple commands sequentially with `;`:
```bash
cmd1; cmd2; cmd3
ls; pwd; echo "Done"
```

### Unix Pipes
Connect commands with `|` to chain operations:
```bash
cmd1 | cmd2
ls | grep ".go" | wc -l
```

### Input/Output Redirection
- `cmd < input.txt` - Redirect input from file
- `cmd > output.txt` - Redirect output to file (overwrite)
- `cmd >> output.txt` - Redirect output to file (append)

### Background Execution
Run commands in background with `&`:
```bash
sleep 10 &
```

### Quote Handling
Support for single and double quotes:
```bash
echo "Hello World"
echo 'Single quotes'
```

### Comment Support
Lines starting with `#` are treated as comments in scripts:
```bash
# This is a comment
echo "This is not a comment"
```

### Functions
Define and use shell functions for code reusability:
```bash
# Define a function
greet() {
    echo "Hello, $1!"
}

# Call the function
greet "World"

# Functions support positional parameters
show_info() {
    echo "Function: $0"
    echo "Args: $@"
    echo "Count: $#"
}

show_info arg1 arg2
```

### Variable Expansion
Expand variables and command substitution:
```bash
# Environment variables (available to child processes)
export NAME="World"
echo "Hello, $NAME!"

# Local variables (shell-only)
local TEMP_DIR="/tmp/myapp"
echo "Using temp dir: $TEMP_DIR"

# Command substitution
local CURRENT_DIR="$(pwd)"
export BUILD_DATE="$(date +%Y%m%d)"
echo "Current dir: $CURRENT_DIR"
echo "Files: `ls | wc -l`"
```

### Command Substitution
Use command output in other commands:
```bash
echo "Today is $(date)"
echo "Found $(ls *.txt | wc -l) text files"
```

### Tilde Expansion
Expand `~` to home directories:
```bash
ls ~/Documents
cd ~user/projects
```

### Glob Patterns
Use wildcards for filename matching:
```bash
ls *.go
echo [Hh]ello*
rm temp_?.txt
```