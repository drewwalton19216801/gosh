# If Statements in gosh

The `gosh` shell supports if statements for conditional execution. If statements provide a way to execute different commands based on the success or failure of test conditions.

## Syntax

```bash
if command
then
    commands...
else
    commands...
fi
```

### With elif branches:

```bash
if command1
then
    commands...
elif command2
then
    commands...
else
    commands...
fi
```

## Features

### Basic If-Then-Else

```bash
if test 5 -gt 3
then
    echo "5 is greater than 3"
else
    echo "Math is broken!"
fi
```

### Using the test Command

The most common way to create conditions is using the `test` command:

```bash
# Numeric comparisons
if test $number -eq 42
then
    echo "The answer to everything!"
fi

# String comparisons
if test $name = "gosh"
then
    echo "Hello, gosh!"
fi

# File tests
if test -f myfile.txt
then
    echo "File exists"
fi
```

### Function Argument Checking

```bash
my_function() {
    if test $# -eq 0
    then
        echo "Usage: my_function <arg>"
        return 1
    else
        echo "Processing: $1"
    fi
}
```

### Multiple Conditions with elif

```bash
check_grade() {
    if test $1 -ge 90
    then
        echo "Grade A"
    elif test $1 -ge 80
    then
        echo "Grade B"
    elif test $1 -ge 70
    then
        echo "Grade C"
    else
        echo "Grade F"
    fi
}
```

## Test Command Operators

### Numeric Comparisons
- `-eq` - equal to
- `-ne` - not equal to
- `-lt` - less than
- `-le` - less than or equal to
- `-gt` - greater than
- `-ge` - greater than or equal to

### String Comparisons
- `=` - equal to
- `!=` - not equal to

### File Tests
- `-f file` - file exists and is a regular file
- `-d dir` - directory exists
- `-e path` - path exists
- `-r file` - file is readable
- `-w file` - file is writable
- `-x file` - file is executable
- `-s file` - file exists and is not empty

## Important Notes

1. **Multi-line support**: If statements are supported in both script files and interactive mode. In interactive mode, the shell will automatically detect when you start an if statement and continue reading lines until you type `fi`.

2. **Multi-line format**: If statements must use the multi-line format with separate `then`, `else`, and `fi` keywords on their own lines.

3. **Nested if statements**: gosh fully supports nested if statements. You can place if statements inside other if statements, and the parser will correctly match each `if` with its corresponding `fi`.

4. **Condition evaluation**: The condition is based on the exit status of the command. Exit status 0 means success (true), non-zero means failure (false).

5. **Variable expansion**: Variables in conditions are expanded before evaluation.

6. **Return statements**: Functions can use `return` statements within if blocks to exit early.

## Examples

### File Backup Script

```bash
#!/usr/local/bin/gosh

backup_file() {
    if test $# -eq 0
    then
        echo "Usage: backup_file <filename>"
        return 1
    fi
    
    if test -f $1
    then
        cp $1 $1.backup
        echo "Backup created: $1.backup"
    else
        echo "Error: File $1 does not exist"
        return 1
    fi
}

backup_file $1
```

### System Information Script

```bash
#!/usr/local/bin/gosh

if test $(uname) = Darwin
then
    echo "Running on macOS"
    echo "System version: $(sw_vers -productVersion)"
elif test $(uname) = Linux
then
    echo "Running on Linux"
    echo "Kernel version: $(uname -r)"
else
    echo "Unknown operating system: $(uname)"
fi
```

### Nested If Statements

```bash
#!/usr/local/bin/gosh

check_directory() {
    if test -d $1
    then
        echo "Directory $1 exists"
        if test -w $1
        then
            echo "Directory is writable"
            if test -r $1
            then
                echo "Directory is readable"
            else
                echo "Directory is not readable"
            fi
        else
            echo "Directory is read-only"
        fi
    else
        echo "Directory $1 does not exist"
    fi
}

check_directory $1
```

## Comparison with Other Shells

The gosh if statement implementation is compatible with POSIX shell syntax and behaves similarly to bash and other common shells, supporting:

- Standard if-then-else-fi syntax
- elif branches for multiple conditions
- Integration with the test command
- Variable expansion in conditions
- Return statements within if blocks

This makes it easy to port existing shell scripts that use if statements to gosh.