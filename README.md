# gosh - A Cross-Platform Shell

<img src="images/gosh_logo.png" alt="Gosh Logo" width="50%" style="display: block; margin: 20px auto;">

A powerful command-line shell written in Go with sh-style scripting support, designed for cross-platform compatibility. Works seamlessly on Windows, macOS, and Linux. Gosh, it's good to have a reliable shell everywhere!

## Features

### Built-in Commands
Gosh provides a comprehensive set of built-in commands for shell operations. See the [Command Reference](doc/command_reference.md) for detailed documentation of all available commands.

### Cross-Platform File Operations
Gosh includes built-in implementations of essential file operations that work consistently across all platforms:
- **File Listing**: `ls` (Unix-style) and `dir` (Windows-style) with support for long format, hidden files, and human-readable sizes
- **File Display**: `cat` (Unix) and `type` (Windows) for viewing file contents
- **File Operations**: `cp`/`copy`, `mv`/`move`, `rm`/`del` for copying, moving, and removing files and directories
- **Directory Operations**: `mkdir`/`md`, `rmdir`/`rd` for creating and removing directories
- **Screen Management**: `clear` (Unix) and `cls` (Windows) for clearing the terminal

Both Unix-style and Windows-style command names are supported, making the shell familiar to users from any platform!

### Shell Features
- **Interactive Mode**: Run `./gosh` for an interactive shell session
- **Command Mode**: Run `./gosh -c "command"` to execute a single command and exit
- **Script Mode**: Run `./gosh script.sh` to execute shell scripts
- **Control Structures**: If statements and case statements supported in both interactive and script modes
- **Functions**: Define and use shell functions with positional parameters
- **Command Chaining**: Execute multiple commands sequentially with `;`
- **Unix Pipes**: Connect commands with `|` to chain operations
- **Input/Output Redirection**: Support for `<`, `>`, and `>>` operators
- **Background Execution**: Run commands in background with `&`
- **Line Continuation**: Split long commands across multiple lines using `\`
- **Environment Variables**: Set and use custom environment variables
- **Command Aliases**: Create shortcuts for frequently used commands
- **Command History**: Track previously executed commands
- **Quote Handling**: Support for single and double quotes in arguments
- **Comment Support**: Lines starting with `#` are treated as comments in scripts
- **Variable Expansion**: Expand environment variables with `$VAR` and `${VAR}` syntax
- **Command Substitution**: Execute commands and use their output with `$(command)` and `` `command` ``
- **Tilde Expansion**: Expand `~` to home directory and user paths
- **Glob Patterns**: Use wildcards for filename matching
- **Signal Handling**: Ctrl-C interrupts running commands without exiting the shell

For detailed information on scripting and advanced features, see the [Scripting Guide](doc/scripting_guide.md) and [Advanced Topics](doc/advanced_topics.md).

## Documentation

- [Cross-Platform Features](doc/cross_platform.md) - Learn about dual command support and platform compatibility
- [Tab Completion Test Guide](doc/tab_completion.md) - Test and verify tab completion features
- [Command Reference](doc/command_reference.md) - List of all built-in commands and their usage
- [Scripting Guide](doc/scripting_guide.md) - Learn how to write shell scripts using gosh
- [Advanced Topics](doc/advanced_topics.md) - Explore more advanced features and techniques

## Building

Build gosh for your platform:

```bash
# Build for current platform
go build -o gosh

# Cross-compile for different platforms
go build -o gosh.exe                    # Windows
GOOS=linux go build -o gosh-linux       # Linux
GOOS=darwin go build -o gosh-macos      # macOS
```

The resulting binary will include all cross-platform built-in commands and work consistently across different operating systems.

## Usage

### Interactive Mode
```bash
./gosh
```
Gosh, that was easy!

### Command Mode
```bash
./gosh -c "command"
```
Execute a single command and exit. Perfect for automation and scripting!

Examples:
```bash
./gosh -c "echo 'Hello World'"
./gosh -c "ls -la | grep .go"
./gosh -c "pwd; echo 'Current directory listed above'"
./gosh -c "ls; echo 'Files listed'; pwd"
```

### Script Mode
```bash
./gosh script.sh
```
By gosh, scripting has never been simpler!

**Windows Note**: Gosh automatically detects and executes shell scripts on Windows, even though they're not native Windows executables. Scripts with `.sh` extensions or shell shebang lines work seamlessly across all platforms.

### Example Script
```bash
#!/usr/bin/env gosh
# Example gosh script - showcasing cross-platform features!

echo "Hello from gosh!"
pwd

# Cross-platform file operations
echo "=== File Listing (Unix-style) ==="
ls -la

echo "=== File Listing (Windows-style) ==="
dir -l

# Create test directory and files
mkdir -p test_project/src
echo "console.log('Hello World');" > test_project/src/app.js
echo "# Test Project" > test_project/README.md

# Show file contents using both styles
echo "=== File Contents (Unix-style) ==="
cat test_project/README.md

echo "=== File Contents (Windows-style) ==="
type test_project/src/app.js

# Copy and move operations
cp test_project/README.md test_project/README_backup.md
copy test_project/src/app.js test_project/app_copy.js

# Environment variables
export MY_VAR=test
echo "MY_VAR is: $MY_VAR"

# Local variables (shell-only)
local TEMP_VAR="temporary value"
echo "TEMP_VAR is: $TEMP_VAR"

# Pipe examples - gosh, pipes are powerful!
echo "testing pipe functionality" | wc -w
ls -la | head -5

# Clean up
rm -rf test_project

# Clear screen (choose your style!)
# clear    # Unix-style
# cls      # Windows-style

echo "Gosh darn it, that's some fine cross-platform scripting!"
```

## Quick Examples

```bash
# Basic usage
ls -la
echo "Hello, World!"

# Cross-platform file operations - use either style!
ls -l                    # Unix-style file listing
dir -l                   # Windows-style file listing
cat README.md            # Unix-style file display
type README.md           # Windows-style file display
cp file.txt backup.txt   # Unix-style copy
copy file.txt backup.txt # Windows-style copy

# Pipes and redirection
ls | grep ".go" > go_files.txt
echo "Hello" >> output.txt

# Variables and expansion
export MY_VAR="Hello"          # Environment variable
local TEMP_VAR="temporary"     # Local variable (shell-only)
echo "$MY_VAR, $(whoami)!"
echo "Temp: $TEMP_VAR"

# Command chaining
echo "Starting..."; pwd; echo "Done!"

# Cross-platform directory operations
mkdir -p projects/gosh     # Create nested directories
cp -r src/ backup/         # Copy directory recursively
rm -rf temp/               # Remove directory and contents
clear                      # Clear screen (Unix-style)
cls                        # Clear screen (Windows-style)

# Line continuation for long commands
echo "This is a very long command" \
  "that spans multiple lines" \
  "for better readability"
```

For comprehensive examples and tutorials, see the [Scripting Guide](doc/scripting_guide.md).

## Signal Handling

Gosh implements proper signal handling to provide a better user experience:

- **Ctrl-C (SIGINT)**: Interrupts the currently running command but does not exit the shell
- **Ctrl-C while idle**: Shows a new prompt but does not exit the shell
- The shell remains active and returns to the prompt in all cases
- This allows you to stop long-running commands without losing your shell session
- The shell can only be exited using the `exit` command or EOF (Ctrl-D)

### Testing Signal Handling

To test the signal handling functionality:

1. Start the shell: `./gosh`
2. Press Ctrl-C while idle - the shell should show a new prompt but not exit
3. Run a long command: `sleep 10`
4. Press Ctrl-C to interrupt the sleep command
5. The shell should return to the prompt without exiting
6. Press Ctrl-C again while idle - should show new prompt, not exit
7. Use `exit` to properly exit the shell

Gosh, that's convenient!

## Architecture

Gosh is organized into multiple modules for maximum readability and maintainability. For detailed architecture information, see the [Advanced Topics](doc/advanced_topics.md) documentation.

## Compatibility

**Cross-Platform Support**: Gosh is designed and tested to work seamlessly across multiple operating systems:
- **Windows**: Full support with both Unix-style (`ls`, `cat`, `cp`) and Windows-style (`dir`, `type`, `copy`) commands, plus automatic shell script execution
- **macOS**: Native support with all Unix-style commands
- **Linux**: Complete compatibility with standard Unix commands

**Shell Script Execution**: Gosh automatically handles shell scripts on all platforms:
- Scripts with `.sh` extension work on Windows without modification
- Shebang lines (`#!/bin/sh`, `#!/usr/bin/env gosh`, etc.) are properly recognized
- No "not a valid Win32 application" errors on Windows

The shell automatically handles platform-specific differences like file paths, permissions, and command conventions. Gosh, it's truly portable!

## Bugs

There are probably many bugs in gosh. Please report any bugs or issues to the [GitHub issue tracker](https://github.com/drewwalton19216801/gosh/issues).


## License

This project is open source and available under the MIT License.