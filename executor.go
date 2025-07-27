package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecuteCommand executes a parsed command
func (s *Shell) ExecuteCommand(cmd *Command) error {
	if cmd == nil || cmd.Name == "" {
		return fmt.Errorf("invalid command")
	}

	// Handle redirection for both built-ins and external commands
	return s.executeWithRedirection(cmd)
}

// executeWithRedirection handles command execution with I/O redirection
func (s *Shell) executeWithRedirection(cmd *Command) error {
	// Save original stdin/stdout/stderr
	origStdin := os.Stdin
	origStdout := os.Stdout
	origStderr := os.Stderr

	// Set up input redirection
	if cmd.Input != "" {
		inputFile, err := os.Open(cmd.Input)
		if err != nil {
			return fmt.Errorf("cannot open input file %s: %v", cmd.Input, err)
		}
		defer inputFile.Close()
		os.Stdin = inputFile
	}

	// Set up output redirection
	if cmd.Output != "" {
		var outputFile *os.File
		var err error
		if cmd.Append {
			outputFile, err = os.OpenFile(cmd.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			outputFile, err = os.Create(cmd.Output)
		}
		if err != nil {
			return fmt.Errorf("cannot create output file %s: %v", cmd.Output, err)
		}
		defer outputFile.Close()
		os.Stdout = outputFile
		os.Stderr = outputFile
	}

	// Restore original stdin/stdout/stderr when done
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Check if it's a built-in command
	if builtin, exists := builtins[cmd.Name]; exists {
		return builtin(s, cmd)
	}

	// Handle external commands
	return s.executeExternal(cmd)
}

// executeExternal runs external commands
func (s *Shell) executeExternal(cmd *Command) error {
	// Resolve command path
	cmdPath, err := s.resolvePath(cmd.Name)
	if err != nil {
		return err
	}

	// Create the command
	execCmd := exec.Command(cmdPath, cmd.Args...)

	// Use current stdin/stdout/stderr (which may be redirected)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Set environment variables
	execCmd.Env = os.Environ()
	for key, value := range s.env {
		execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute the command
	if cmd.Background {
		return execCmd.Start()
	}

	return execCmd.Run()
}

// resolvePath finds the full path to a command
func (s *Shell) resolvePath(cmdName string) (string, error) {
	// If it contains a slash, treat as relative/absolute path
	if strings.Contains(cmdName, "/") {
		if filepath.IsAbs(cmdName) {
			return cmdName, nil
		}
		// Relative path
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, cmdName), nil
	}

	// Search in PATH
	path, err := exec.LookPath(cmdName)
	if err != nil {
		return "", fmt.Errorf("command not found: %s", cmdName)
	}

	return path, nil
}