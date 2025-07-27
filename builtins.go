package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BuiltinFunc represents a built-in command function
type BuiltinFunc func(*Shell, *Command) error

// builtins maps command names to their implementation functions
var builtins map[string]BuiltinFunc

// init initializes the builtins map
func init() {
	builtins = map[string]BuiltinFunc{
		"exit":     cmdExit,
		"cd":       cmdCd,
		"pwd":      cmdPwd,
		"echo":     cmdEcho,
		"env":      cmdEnv,
		"export":   cmdExport,
		"local":    cmdLocal,
		"unset":    cmdUnset,
		"alias":    cmdAlias,
		"unalias":  cmdUnalias,
		"history":  cmdHistory,
		"help":     cmdHelp,
		"which":    cmdWhich,
		"case":     cmdCase,
		"declare":  cmdDeclare,
		"test":     cmdTest,
		"[":        cmdTest,
		"return":   cmdReturn,
		"if":       cmdIf,
		"then":     cmdThen,
		"else":     cmdElse,
		"elif":     cmdElif,
		"fi":       cmdFi,
	}
}

// cmdExit implements the exit command
func cmdExit(s *Shell, cmd *Command) error {
	code := 0
	if len(cmd.Args) > 0 {
		if c, err := strconv.Atoi(cmd.Args[0]); err == nil {
			code = c
		}
	}
	s.Exit(code)
	return nil
}

// cmdCd implements the cd command
func cmdCd(s *Shell, cmd *Command) error {
	var dir string
	if len(cmd.Args) == 0 {
		// Go to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get home directory: %v", err)
		}
		dir = home
	} else {
		dir = cmd.Args[0]
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %v", err)
	}

	return nil
}

// cmdPwd implements the pwd command
func cmdPwd(s *Shell, cmd *Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %v", err)
	}
	fmt.Println(wd)
	return nil
}

// cmdEcho implements the echo command
func cmdEcho(s *Shell, cmd *Command) error {
	// Remove quotes from arguments and process escape sequences
	var processedArgs []string
	for _, arg := range cmd.Args {
		processedArg := arg
		// Remove quotes if present
		if len(arg) >= 2 && ((arg[0] == '"' && arg[len(arg)-1] == '"') || (arg[0] == '\'' && arg[len(arg)-1] == '\'')) {
			processedArg = arg[1:len(arg)-1]
		}
		// Process escape sequences
		processedArg = processEscapeSequences(processedArg)
		processedArgs = append(processedArgs, processedArg)
	}
	output := strings.Join(processedArgs, " ")
	fmt.Println(output)
	return nil
}

// processEscapeSequences processes common escape sequences in strings
func processEscapeSequences(s string) string {
	result := strings.Builder{}
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '"':
				result.WriteByte('"')
				i++ // Skip the next character
			case '\'':
				result.WriteByte('\'')
				i++ // Skip the next character
			case '\\':
				result.WriteByte('\\')
				i++ // Skip the next character
			case 'n':
				result.WriteByte('\n')
				i++ // Skip the next character
			case 't':
				result.WriteByte('\t')
				i++ // Skip the next character
			case 'r':
				result.WriteByte('\r')
				i++ // Skip the next character
			default:
				// Unknown escape sequence, keep the backslash
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

// cmdEnv implements the env command
func cmdEnv(s *Shell, cmd *Command) error {
	// Print system environment
	for _, env := range os.Environ() {
		fmt.Println(env)
	}
	// Print shell-specific environment
	for key, value := range s.env {
		fmt.Printf("%s=%s\n", key, value)
	}
	return nil
}

// cmdExport implements the export command
func cmdExport(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("export: missing variable assignment")
	}

	for _, arg := range cmd.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("export: invalid assignment: %s", arg)
		}
		key, value := parts[0], parts[1]
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		s.env[key] = value
		os.Setenv(key, value)
	}

	return nil
}

// cmdLocal implements the local command
func cmdLocal(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("local: missing variable assignment")
	}

	for _, arg := range cmd.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("local: invalid assignment: %s", arg)
		}
		key, value := parts[0], parts[1]
		
		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}
		
		// Handle command substitution if present
		if strings.Contains(value, "$(") {
			expandedValue, err := s.expandToken(value)
			if err != nil {
				return fmt.Errorf("error expanding variable value: %v", err)
			}
			if len(expandedValue) == 1 {
				value = expandedValue[0]
			}
		}
		
		// Set the variable in the shell environment only (not system environment)
		s.env[key] = value
	}

	return nil
}

// cmdUnset implements the unset command
func cmdUnset(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("unset: missing variable or function name")
	}

	// Check for -f flag to unset functions
	if len(cmd.Args) >= 2 && cmd.Args[0] == "-f" {
		for i := 1; i < len(cmd.Args); i++ {
			delete(s.functions, cmd.Args[i])
		}
		return nil
	}

	// Default behavior: unset variables
	for _, arg := range cmd.Args {
		// Check if it's a function first
		if _, exists := s.functions[arg]; exists {
			delete(s.functions, arg)
		} else {
			// Unset as variable
			delete(s.env, arg)
			os.Unsetenv(arg)
		}
	}

	return nil
}

// cmdAlias implements the alias command
func cmdAlias(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		// List all aliases
		for name, value := range s.aliases {
			fmt.Printf("alias %s='%s'\n", name, value)
		}
		return nil
	}

	for _, arg := range cmd.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("alias: invalid assignment: %s", arg)
		}
		name, value := parts[0], parts[1]
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		s.aliases[name] = value
	}

	return nil
}

// cmdUnalias implements the unalias command
func cmdUnalias(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("unalias: missing alias name")
	}

	for _, arg := range cmd.Args {
		delete(s.aliases, arg)
	}

	return nil
}

// cmdHistory implements the history command
func cmdHistory(s *Shell, cmd *Command) error {
	for i, line := range s.history {
		fmt.Printf("%4d  %s\n", i+1, line)
	}
	return nil
}

// cmdHelp implements the help command
func cmdHelp(s *Shell, cmd *Command) error {
	fmt.Println("gosh - A simple shell written in Go")
	fmt.Println("")
	fmt.Println("Built-in commands:")
	fmt.Println("  exit [code]     - Exit the shell with optional exit code")
	fmt.Println("  cd [dir]        - Change directory (defaults to home)")
	fmt.Println("  pwd             - Print working directory")
	fmt.Println("  echo [args...]  - Print arguments")
	fmt.Println("  env             - Show environment variables")
	fmt.Println("  export VAR=val  - Set environment variable")
	fmt.Println("  local VAR=val   - Set local variable (shell only)")
	fmt.Println("  unset VAR       - Unset environment variable or function")
	fmt.Println("  unset -f FUNC   - Unset function")
	fmt.Println("  alias name=cmd  - Create command alias")
	fmt.Println("  unalias name    - Remove alias")
	fmt.Println("  declare         - List all defined functions")
	fmt.Println("  declare -f      - List function names only")
	fmt.Println("  history         - Show command history")
	fmt.Println("  which cmd       - Show path to command")
	fmt.Println("  help            - Show this help")
	fmt.Println("")
	fmt.Println("Features:")
	fmt.Println("  - Tab completion: Press TAB to complete commands and file paths")
	fmt.Println("  - Command chaining: cmd1; cmd2; cmd3")
	fmt.Println("  - Unix pipes: cmd1 | cmd2 | cmd3")
	fmt.Println("  - Input/output redirection: cmd < input.txt > output.txt")
	fmt.Println("  - Background execution: cmd &")
	fmt.Println("  - Script execution: gosh script.sh")
	fmt.Println("  - Case statements: case $var in pattern) commands ;; esac")
	fmt.Println("  - User-defined functions: function_name() { commands; }")
	fmt.Println("  - Function parameters: $1, $2, ..., $#, $0")
	fmt.Println("  - Environment variables and aliases")
	fmt.Println("  - Variable expansion: $VAR, ${VAR}")
	fmt.Println("  - Command substitution: $(cmd), `cmd`")
	fmt.Println("  - Tilde expansion: ~, ~/path, ~user/path")
	fmt.Println("  - Glob patterns: *.txt, file?.log, [abc]*")
	return nil
}

// cmdWhich implements the which command
func cmdWhich(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("which: missing command name")
	}

	for _, cmdName := range cmd.Args {
		// Check if it's a user-defined function
		if _, exists := s.functions[cmdName]; exists {
			fmt.Printf("%s: user-defined function\n", cmdName)
			continue
		}

		// Check if it's a built-in
		if _, exists := builtins[cmdName]; exists {
			fmt.Printf("%s: shell builtin\n", cmdName)
			continue
		}

		// Check if it's an alias
		if alias, exists := s.aliases[cmdName]; exists {
			fmt.Printf("%s: aliased to %s\n", cmdName, alias)
			continue
		}

		// Look for external command
		path, err := s.resolvePath(cmdName)
		if err != nil {
			fmt.Printf("%s: not found\n", cmdName)
		} else {
			// Get absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Println(path)
			} else {
				fmt.Println(absPath)
			}
		}
	}

	return nil
}

// cmdCase implements the case command (for interactive use)
func cmdCase(s *Shell, cmd *Command) error {
	return fmt.Errorf("case statements are only supported in scripts, not interactive mode")
}

// cmdDeclare implements the declare command for listing functions
func cmdDeclare(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		// List all functions
		if len(s.functions) == 0 {
			fmt.Println("No functions defined")
			return nil
		}
		
		fmt.Println("Defined functions:")
		for name, function := range s.functions {
			fmt.Printf("%s() {\n", name)
			for _, line := range function.Body {
				fmt.Printf("  %s\n", line)
			}
			fmt.Println("}")
			fmt.Println()
		}
		return nil
	}
	
	// Check for -f flag to list only function names
	if len(cmd.Args) == 1 && cmd.Args[0] == "-f" {
		for name := range s.functions {
			fmt.Println(name)
		}
		return nil
	}
	
	return fmt.Errorf("declare: unsupported option")
}

// TestFailureError represents a test command failure (exit code 1)
type TestFailureError struct{}

func (e TestFailureError) Error() string {
	return "test failed"
}

// cmdTest implements the test command (and [ command)
func cmdTest(s *Shell, cmd *Command) error {
	args := cmd.Args
	
	// Handle [ command - remove trailing ] if present
	if cmd.Name == "[" {
		if len(args) == 0 || args[len(args)-1] != "]" {
			return fmt.Errorf("[: missing closing ']'")
		}
		args = args[:len(args)-1]
	}
	
	// No arguments means false
	if len(args) == 0 {
		return TestFailureError{}
	}
	
	// Single argument means test if non-empty
	if len(args) == 1 {
		if args[0] != "" {
			return nil // success
		} else {
			return TestFailureError{}
		}
	}
	
	// Two arguments with unary operator
	if len(args) == 2 {
		operator := args[0]
		operand := args[1]
		
		switch operator {
		case "-z":
			// String is empty
			if operand == "" {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-n":
			// String is non-empty
			if operand != "" {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-f":
			// File exists and is regular file
			if info, err := os.Stat(operand); err == nil && info.Mode().IsRegular() {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-d":
			// Directory exists
			if info, err := os.Stat(operand); err == nil && info.IsDir() {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-e":
			// File or directory exists
			if _, err := os.Stat(operand); err == nil {
				return nil
			} else {
				return TestFailureError{}
			}
		default:
			return fmt.Errorf("test: unknown unary operator: %s", operator)
		}
	}
	
	// Three arguments with binary operator
	if len(args) == 3 {
		left := args[0]
		operator := args[1]
		right := args[2]
		
		switch operator {
		case "=", "==":
			// String equality
			if left == right {
				return nil
			} else {
				return TestFailureError{}
			}
		case "!=":
			// String inequality
			if left != right {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-eq":
			// Numeric equality
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum == rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-ne":
			// Numeric inequality
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum != rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-lt":
			// Numeric less than
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum < rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-le":
			// Numeric less than or equal
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum <= rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-gt":
			// Numeric greater than
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum > rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		case "-ge":
			// Numeric greater than or equal
			leftNum, err1 := strconv.Atoi(left)
			rightNum, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("test: non-numeric argument")
			}
			if leftNum >= rightNum {
				return nil
			} else {
				return TestFailureError{}
			}
		default:
			return fmt.Errorf("test: unknown binary operator: %s", operator)
		}
	}
	
	return fmt.Errorf("test: too many arguments")
}

// ReturnError represents a function return with exit code
type ReturnError struct {
	Code int
}

func (e ReturnError) Error() string {
	return fmt.Sprintf("return %d", e.Code)
}

// cmdReturn implements the return command for functions
func cmdReturn(s *Shell, cmd *Command) error {
	code := 0
	if len(cmd.Args) > 0 {
		if c, err := strconv.Atoi(cmd.Args[0]); err == nil {
			code = c
		}
	}
	return ReturnError{Code: code}
}

// cmdIf implements the if command (for interactive use)
func cmdIf(s *Shell, cmd *Command) error {
	return fmt.Errorf("if statements are only supported in scripts, not interactive mode")
}

// cmdThen implements the then command (for interactive use)
func cmdThen(s *Shell, cmd *Command) error {
	return fmt.Errorf("then statements are only supported in scripts, not interactive mode")
}

// cmdElse implements the else command (for interactive use)
func cmdElse(s *Shell, cmd *Command) error {
	return fmt.Errorf("else statements are only supported in scripts, not interactive mode")
}

// cmdElif implements the elif command (for interactive use)
func cmdElif(s *Shell, cmd *Command) error {
	return fmt.Errorf("elif statements are only supported in scripts, not interactive mode")
}

// cmdFi implements the fi command (for interactive use)
func cmdFi(s *Shell, cmd *Command) error {
	return fmt.Errorf("fi statements are only supported in scripts, not interactive mode")
}