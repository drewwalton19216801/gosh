# Command Reference

This document provides a comprehensive reference for all built-in commands available in gosh.

## Built-in Commands

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

### `env`
Show all environment variables.

**Usage:**
```bash
env
```

### `export VAR=value`
Set an environment variable.

**Usage:**
```bash
export MY_VAR="Hello"
export PATH="$PATH:/new/path"
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
export NAME="World"
echo "Hello, $NAME!"
echo "Current dir: $(pwd)"
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