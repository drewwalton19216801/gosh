package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExecuteCommand executes a parsed command
func (s *Shell) ExecuteCommand(cmd *Command) error {
	if cmd == nil || cmd.Name == "" {
		return fmt.Errorf("invalid command")
	}

	// Handle special nested if command
	if cmd.Name == "__nested_if__" {
		return s.executeIfStatement(cmd.Args)
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

	// Expand aliases (case-insensitive)
	for {
		var alias string
		var found bool
		for aliasName, aliasValue := range s.aliases {
			if strings.EqualFold(cmd.Name, aliasName) {
				alias = aliasValue
				found = true
				break
			}
		}
		if !found {
			break
		}
		aliasArgs := strings.Fields(alias)
		cmd.Name = aliasArgs[0]
		cmd.Args = append(aliasArgs[1:], cmd.Args...)
	}

	// Check if it's a user-defined function (case-insensitive)
	for funcName := range s.functions {
		if strings.EqualFold(cmd.Name, funcName) {
			return s.executeFunction(funcName, cmd.Args)
		}
	}

	// Check if it's a built-in command (case-insensitive)
	for builtinName, builtin := range builtins {
		if strings.EqualFold(cmd.Name, builtinName) {
			return builtin(s, cmd)
		}
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

	// On Windows, check if this is a shell script and handle it specially
	if runtime.GOOS == "windows" && s.isShellScript(cmdPath) {
		return s.executeShellScriptOnWindows(cmdPath, cmd)
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

	// Start the command and track it for signal handling
	err = execCmd.Start()
	if err != nil {
		return err
	}

	// Set the current command for signal handling
	s.currentCmd = execCmd.Process

	// Wait for the command to complete
	err = execCmd.Wait()

	// Clear the current command
	s.currentCmd = nil

	return err
}

// ExecutePipeline executes a series of commands connected by pipes
func (s *Shell) ExecutePipeline(commands []*Command) error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands in pipeline")
	}

	if len(commands) == 1 {
		return s.ExecuteCommand(commands[0])
	}

	// Check for background execution - only last command can be background
	for i := 0; i < len(commands)-1; i++ {
		if commands[i].Background {
			return fmt.Errorf("only the last command in a pipeline can run in background")
		}
	}

	// Create pipes for connecting commands
	var pipes []io.ReadCloser
	var execCmds []*exec.Cmd
	var processes []*os.Process

	for i, cmd := range commands {
		// Resolve command path
		cmdPath, err := s.resolvePath(cmd.Name)
		if err != nil {
			// Check if it's a builtin (case-insensitive)
			builtinFound := false
			for builtinName := range builtins {
				if strings.EqualFold(cmd.Name, builtinName) {
					builtinFound = true
					break
				}
			}
			if !builtinFound {
				return err
			}
			// Builtins in pipelines are not supported for now
			return fmt.Errorf("builtin commands not supported in pipelines: %s", cmd.Name)
		}

		// Create the command
		execCmd := exec.Command(cmdPath, cmd.Args...)

		// Set environment
		execCmd.Env = os.Environ()
		for key, value := range s.env {
			execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", key, value))
		}

		// Handle input
		if i == 0 {
			// First command - use stdin or input redirection
			if cmd.Input != "" {
				inputFile, err := os.Open(cmd.Input)
				if err != nil {
					return fmt.Errorf("cannot open input file %s: %v", cmd.Input, err)
				}
				execCmd.Stdin = inputFile
				defer inputFile.Close()
			} else {
				execCmd.Stdin = os.Stdin
			}
		} else {
			// Middle/last commands - use pipe from previous command
			execCmd.Stdin = pipes[i-1]
		}

		// Handle output
		if i == len(commands)-1 {
			// Last command - use stdout or output redirection
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
				execCmd.Stdout = outputFile
				defer outputFile.Close()
			} else {
				execCmd.Stdout = os.Stdout
			}
			execCmd.Stderr = os.Stderr
		} else {
			// Middle commands - create pipe to next command
			stdoutPipe, err := execCmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("failed to create pipe: %v", err)
			}
			pipes = append(pipes, stdoutPipe)
			execCmd.Stderr = os.Stderr
		}

		execCmds = append(execCmds, execCmd)
	}

	// Start all commands
	for i, execCmd := range execCmds {
		if err := execCmd.Start(); err != nil {
			return fmt.Errorf("failed to start command %s: %v", commands[i].Name, err)
		}
		processes = append(processes, execCmd.Process)
	}

	// Set the first process as current for signal handling
	if len(processes) > 0 {
		s.currentCmd = processes[0]
	}

	// Close pipe read ends that we created
	for _, pipe := range pipes {
		defer pipe.Close()
	}

	// Wait for all commands to complete
	var lastErr error
	for i, execCmd := range execCmds {
		if commands[len(commands)-1].Background && i == len(execCmds)-1 {
			// Last command is background - don't wait
			continue
		}
		if err := execCmd.Wait(); err != nil {
			lastErr = fmt.Errorf("command %s failed: %v", commands[i].Name, err)
		}
	}

	// Clear the current command
	s.currentCmd = nil

	return lastErr
}

// resolvePath finds the full path to a command
func (s *Shell) resolvePath(cmdName string) (string, error) {
	// If it contains a slash or backslash, treat as relative/absolute path
	if strings.Contains(cmdName, "/") || strings.Contains(cmdName, "\\") {
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

// isShellScript checks if a file is a shell script by examining its extension and shebang
func (s *Shell) isShellScript(filePath string) bool {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".sh" {
		return true
	}

	// Check for shebang line
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := strings.TrimSpace(scanner.Text())
		// Check for common shell shebangs
		if strings.HasPrefix(firstLine, "#!/bin/sh") ||
			strings.HasPrefix(firstLine, "#!/bin/bash") ||
			strings.HasPrefix(firstLine, "#!/usr/bin/env sh") ||
			strings.HasPrefix(firstLine, "#!/usr/bin/env bash") ||
			strings.HasPrefix(firstLine, "#!/usr/bin/env gosh") ||
			strings.Contains(firstLine, "gosh") {
			return true
		}
	}

	return false
}

// executeShellScriptOnWindows executes a shell script on Windows by running it through gosh
func (s *Shell) executeShellScriptOnWindows(scriptPath string, cmd *Command) error {
	// Get the current executable path (gosh.exe)
	goshPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find gosh executable: %v", err)
	}

	// Prepare arguments: gosh.exe scriptPath [script args...]
	args := []string{scriptPath}
	args = append(args, cmd.Args...)

	// Create the command to run gosh with the script
	execCmd := exec.Command(goshPath, args...)

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

	// Start the command and track it for signal handling
	err = execCmd.Start()
	if err != nil {
		return err
	}

	// Set the current command for signal handling
	s.currentCmd = execCmd.Process

	// Wait for the command to complete
	err = execCmd.Wait()

	// Clear the current command
	s.currentCmd = nil

	return err
}
