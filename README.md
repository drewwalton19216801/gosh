# gosh - A Simple Shell

A basic command-line shell written in Go with sh-style scripting support, designed for macOS compatibility. Gosh, it's good to have a reliable shell!

## Features

### Built-in Commands
- `exit [code]` - Exit the shell with optional exit code
- `cd [dir]` - Change directory (defaults to home directory)
- `pwd` - Print working directory
- `echo [args...]` - Print arguments
- `env` - Show environment variables
- `export VAR=value` - Set environment variable
- `unset VAR` - Unset environment variable
- `alias name=command` - Create command alias
- `unalias name` - Remove alias
- `history` - Show command history
- `which cmd` - Show path to command
- `help` - Show help information

### Shell Features
- **Interactive Mode**: Run `./gosh` for an interactive shell session
- **Command Mode**: Run `./gosh -c "command"` to execute a single command and exit
- **Script Mode**: Run `./gosh script.sh` to execute shell scripts
- **Command Chaining**: Execute multiple commands sequentially with `;`
  - `cmd1; cmd2; cmd3` - Run commands one after another
  - `ls; pwd; echo "Done"` - List files, show directory, then print message
  - Works with pipes and redirection: `ls > files.txt; cat files.txt | wc -l`
- **Unix Pipes**: Connect commands with `|` to chain operations
  - `cmd1 | cmd2` - Pass output of cmd1 as input to cmd2
  - `cmd1 | cmd2 | cmd3` - Chain multiple commands together
  - Works with input/output redirection: `cat file.txt | grep pattern > results.txt`
- **Input/Output Redirection**: 
  - `cmd < input.txt` - Redirect input from file
  - `cmd > output.txt` - Redirect output to file (overwrite)
  - `cmd >> output.txt` - Redirect output to file (append)
- **Background Execution**: `cmd &` - Run command in background
- **Environment Variables**: Set and use custom environment variables
- **Command Aliases**: Create shortcuts for frequently used commands
- **Command History**: Track previously executed commands
- **Quote Handling**: Support for single and double quotes in arguments
- **Comment Support**: Lines starting with `#` are treated as comments in scripts

## Building

```bash
go build -o gosh
```

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

## Architecture

Gosh, we've organized the shell into multiple modules for maximum readability:

- `main.go` - Entry point and argument handling
- `shell.go` - Core shell structure and main loop
- `parser.go` - Command line parsing and tokenization
- `executor.go` - Command execution logic
- `builtins.go` - Built-in command implementations
- `utils.go` - Utility functions

## Compatibility

Designed and tested for macOS. Gosh, the shell uses Go's standard library for cross-platform compatibility where possible - oh my gosh, it's portable!

## License

This project is open source and available under the MIT License.