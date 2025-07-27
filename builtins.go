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
		"exit":    cmdExit,
		"cd":      cmdCd,
		"pwd":     cmdPwd,
		"echo":    cmdEcho,
		"env":     cmdEnv,
		"export":  cmdExport,
		"unset":   cmdUnset,
		"alias":   cmdAlias,
		"unalias": cmdUnalias,
		"history": cmdHistory,
		"help":    cmdHelp,
		"which":   cmdWhich,
		"case":    cmdCase,
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

// cmdUnset implements the unset command
func cmdUnset(s *Shell, cmd *Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("unset: missing variable name")
	}

	for _, arg := range cmd.Args {
		delete(s.env, arg)
		os.Unsetenv(arg)
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
	fmt.Println("  unset VAR       - Unset environment variable")
	fmt.Println("  alias name=cmd  - Create command alias")
	fmt.Println("  unalias name    - Remove alias")
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