package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

// expandToken performs all shell expansions on a single token
func (s *Shell) expandToken(token string) ([]string, error) {
	// First expand variables and command substitutions
	expanded, err := s.expandVariablesAndCommands(token)
	if err != nil {
		return nil, err
	}

	// Then expand tilde
	expanded, err = s.expandTilde(expanded)
	if err != nil {
		return nil, err
	}

	// Remove quotes before glob expansion
	expanded = removeQuotes(expanded)

	// Finally expand globs (this can return multiple results)
	return s.expandGlob(expanded)
}

// expandVariablesAndCommands expands $VAR, ${VAR}, $(cmd), and `cmd`
func (s *Shell) expandVariablesAndCommands(input string) (string, error) {
	result := input
	var err error

	// Expand command substitutions first ($(cmd) and `cmd`)
	result, err = s.expandCommandSubstitution(result)
	if err != nil {
		return "", err
	}

	// Then expand variables
	result = s.expandVariables(result)

	return result, nil
}

// expandVariables expands $VAR and ${VAR} patterns
func (s *Shell) expandVariables(input string) string {
	// Handle ${VAR} first
	bracePattern := regexp.MustCompile(`\$\{([^}]+)\}`)
	result := bracePattern.ReplaceAllStringFunc(input, func(match string) string {
		varName := match[2 : len(match)-1] // Remove ${ and }
		return s.getVariable(varName)
	})

	// Handle $VAR and positional parameters like $1, $2, etc.
	// This pattern matches: $VAR, $1, $2, $#, $@, $*, etc.
	varPattern := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*|[0-9]+|[#@*])`)
	result = varPattern.ReplaceAllStringFunc(result, func(match string) string {
		varName := match[1:] // Remove $
		return s.getVariable(varName)
	})

	return result
}

// getVariable gets a variable value from shell env or system env
func (s *Shell) getVariable(name string) string {
	// Handle positional parameters if we're in a function context
	if len(s.functionStack) > 0 {
		ctx := s.functionStack[len(s.functionStack)-1]

		// Handle special parameters
		switch name {
		case "0":
			return ctx.Name
		case "#":
			return fmt.Sprintf("%d", len(ctx.Args))
		case "@":
			return strings.Join(ctx.Args, " ")
		case "*":
			return strings.Join(ctx.Args, " ")
		default:
			// Handle numbered parameters $1, $2, etc.
			if len(name) > 0 && name[0] >= '1' && name[0] <= '9' {
				if paramNum := int(name[0] - '0'); paramNum <= len(ctx.Args) {
					return ctx.Args[paramNum-1]
				}
				return ""
			}
		}
	}

	// Check shell-specific environment first
	if value, exists := s.env[name]; exists {
		return value
	}
	// Fall back to system environment
	return os.Getenv(name)
}

// expandCommandSubstitution expands $(cmd) and `cmd` patterns
func (s *Shell) expandCommandSubstitution(input string) (string, error) {
	// Handle $(cmd) first
	parenPattern := regexp.MustCompile(`\$\(([^)]+)\)`)
	result := input
	var err error

	result = parenPattern.ReplaceAllStringFunc(result, func(match string) string {
		cmdStr := match[2 : len(match)-1] // Remove $( and )
		output, cmdErr := s.executeCommandSubstitution(cmdStr)
		if cmdErr != nil {
			err = cmdErr
			return match // Return original on error
		}
		return strings.TrimSpace(output)
	})

	if err != nil {
		return "", err
	}

	// Handle `cmd` (backticks)
	backtickPattern := regexp.MustCompile("`([^`]+)`")
	result = backtickPattern.ReplaceAllStringFunc(result, func(match string) string {
		cmdStr := match[1 : len(match)-1] // Remove backticks
		output, cmdErr := s.executeCommandSubstitution(cmdStr)
		if cmdErr != nil {
			err = cmdErr
			return match // Return original on error
		}
		return strings.TrimSpace(output)
	})

	return result, err
}

// executeCommandSubstitution executes a command and returns its output
func (s *Shell) executeCommandSubstitution(cmdStr string) (string, error) {
	// Parse the command
	commandChain, err := ParseLine(cmdStr)
	if err != nil {
		return "", fmt.Errorf("command substitution parse error: %v", err)
	}

	if commandChain == nil || len(commandChain.Pipelines) == 0 {
		return "", nil
	}

	// For command substitution, we only support single pipelines
	if len(commandChain.Pipelines) > 1 {
		return "", fmt.Errorf("command chaining not supported in command substitution")
	}

	commands := commandChain.Pipelines[0]
	if len(commands) == 1 {
		// Single command
		return s.executeCommandForOutput(commands[0])
	} else {
		// Pipeline
		return s.executePipelineForOutput(commands)
	}
}

// executeCommandForOutput executes a single command and captures its output
func (s *Shell) executeCommandForOutput(cmd *Command) (string, error) {
	// Check if it's a built-in command
	if builtin, exists := builtins[cmd.Name]; exists {
		// For built-ins, we need to capture stdout
		return s.executeBuiltinForOutput(builtin, cmd)
	}

	// Handle external commands
	cmdPath, err := s.resolvePath(cmd.Name)
	if err != nil {
		return "", err
	}

	execCmd := exec.Command(cmdPath, cmd.Args...)
	execCmd.Env = os.Environ()
	for key, value := range s.env {
		execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	output, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("command substitution failed: %v", err)
	}

	return string(output), nil
}

// executeBuiltinForOutput executes a builtin and captures its output
func (s *Shell) executeBuiltinForOutput(builtin func(*Shell, *Command) error, cmd *Command) (string, error) {
	// This is a simplified implementation - in a full shell you'd need to
	// redirect stdout to capture the output properly
	// For now, we'll handle common cases
	switch cmd.Name {
	case "pwd":
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return wd, nil
	case "echo":
		return strings.Join(cmd.Args, " "), nil
	default:
		return "", fmt.Errorf("builtin '%s' not supported in command substitution", cmd.Name)
	}
}

// executePipelineForOutput executes a pipeline and captures its output
func (s *Shell) executePipelineForOutput(commands []*Command) (string, error) {
	// This is a simplified implementation
	// In a full implementation, you'd need to set up the full pipeline
	return "", fmt.Errorf("pipelines in command substitution not fully implemented")
}

// expandTilde expands ~ and ~user patterns
func (s *Shell) expandTilde(input string) (string, error) {
	if !strings.HasPrefix(input, "~") {
		return input, nil
	}

	if input == "~" {
		// Just ~, expand to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return input, err
		}
		return home, nil
	}

	if strings.HasPrefix(input, "~/") {
		// ~/path, expand ~ to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return input, err
		}
		return filepath.Join(home, input[2:]), nil
	}

	// ~user or ~user/path
	slashIndex := strings.Index(input, "/")
	var username, rest string
	if slashIndex == -1 {
		username = input[1:]
		rest = ""
	} else {
		username = input[1:slashIndex]
		rest = input[slashIndex+1:]
	}

	user, err := user.Lookup(username)
	if err != nil {
		return input, err // Return original if user not found
	}

	if rest == "" {
		return user.HomeDir, nil
	}
	return filepath.Join(user.HomeDir, rest), nil
}

// expandGlob expands glob patterns like *.txt, ?, etc.
func (s *Shell) expandGlob(input string) ([]string, error) {
	// Check if the input contains glob characters
	if !strings.ContainsAny(input, "*?[]") {
		return []string{input}, nil
	}

	// Use filepath.Glob to expand the pattern
	matches, err := filepath.Glob(input)
	if err != nil {
		return nil, fmt.Errorf("glob expansion error: %v", err)
	}

	// If no matches found, return the original pattern
	if len(matches) == 0 {
		return []string{input}, nil
	}

	return matches, nil
}

// expandTokens expands a slice of tokens, handling the fact that
// glob expansion can turn one token into multiple tokens
func (s *Shell) expandTokens(tokens []string) ([]string, error) {
	var result []string
	for _, token := range tokens {
		expanded, err := s.expandToken(token)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}
