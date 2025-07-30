package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
)

// Function represents a user-defined shell function
type Function struct {
	Name   string
	Params []string
	Body   []string
}

// FunctionContext holds the execution context for a function
type FunctionContext struct {
	Name string
	Args []string
}

// Shell represents the main shell instance
type Shell struct {
	env           map[string]string
	aliases       map[string]string
	functions     map[string]*Function
	functionStack []*FunctionContext
	history       []string
	running       bool
	rl            *readline.Instance
	currentCmd    *os.Process
	sigChan       chan os.Signal
}

// NewShell creates a new shell instance
func NewShell() *Shell {
	s := &Shell{
		env:           make(map[string]string),
		aliases:       make(map[string]string),
		functions:     make(map[string]*Function),
		functionStack: make([]*FunctionContext, 0),
		history:       make([]string, 0),
		running:       true,
		sigChan:       make(chan os.Signal, 1),
	}

	// Set up signal handling for SIGINT (Ctrl-C)
	signal.Notify(s.sigChan, syscall.SIGINT)
	go s.handleSignals()

	// Get cross-platform history file path
	historyFile := getHistoryFilePath()

	// Initialize readline with history support and tab completion
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "gosh> ",
		HistoryFile:     historyFile,
		AutoComplete:    s.createCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	s.rl = rl

	return s
}

// handleSignals processes incoming signals
func (s *Shell) handleSignals() {
	for sig := range s.sigChan {
		switch sig {
		case syscall.SIGINT:
			// If there's a running command, terminate it
			if s.currentCmd != nil {
				// Send SIGINT to the running process
				s.currentCmd.Signal(syscall.SIGINT)
				s.currentCmd = nil
			}
			// Don't exit the shell, just interrupt the current command
		}
	}
}

// Run starts the interactive shell loop
func (s *Shell) Run() {
	fmt.Println("gosh - A simple shell")
	fmt.Println("Type 'help' for available commands or 'exit' to quit.")
	fmt.Println("Use up/down arrow keys to navigate command history.")
	fmt.Println("Press TAB for command and file completion.")

	defer s.rl.Close()

	for s.running {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "unknown"
		}

		// Read command line with line continuation support
		line, err := s.readLineWithContinuation(cwd)
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip comments in interactive mode
		if strings.HasPrefix(line, "#") {
			continue
		}

		s.history = append(s.history, line)
		if err := s.ExecuteLine(line); err != nil {
			fmt.Fprintf(os.Stderr, "gosh: %v\n", err)
		}
	}
}

// ExecuteLine processes a single command line or multi-line construct
func (s *Shell) ExecuteLine(line string) error {
	// Check for variable assignment (VAR=value)
	if s.isVariableAssignment(line) {
		return s.handleVariableAssignment(line)
	}

	// Check if this is a multi-line construct
	lines := strings.Split(line, "\n")
	if len(lines) > 1 {
		return s.executeMultiLineConstruct(lines)
	}

	commandChain, err := ParseLine(line)
	if err != nil {
		return err
	}

	if commandChain == nil {
		return nil
	}

	// Apply expansions to all commands in the chain
	err = s.expandCommandChain(commandChain)
	if err != nil {
		return err
	}

	// Execute each pipeline in the chain sequentially
	for _, commands := range commandChain.Pipelines {
		if len(commands) == 1 {
			// Single command - execute normally
			if err := s.ExecuteCommand(commands[0]); err != nil {
				return err
			}
		} else {
			// Pipeline - execute as connected commands
			if err := s.ExecutePipeline(commands); err != nil {
				return err
			}
		}
	}

	return nil
}

// ExecuteScript runs a script file with arguments
func (s *Shell) ExecuteScript(filename string, args []string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open script file: %v", err)
	}
	defer file.Close()

	// Read all lines first to handle multi-line constructs like case statements
	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Set up script context for positional parameters
	scriptCtx := &FunctionContext{
		Name: filename,
		Args: args,
	}

	// Push script context onto function stack
	s.functionStack = append(s.functionStack, scriptCtx)
	defer func() {
		// Pop script context when script exits
		if len(s.functionStack) > 0 {
			s.functionStack = s.functionStack[:len(s.functionStack)-1]
		}
	}()

	return s.executeScriptLines(allLines)
}

// executeMultiLineConstruct handles multi-line constructs in interactive mode
func (s *Shell) executeMultiLineConstruct(lines []string) error {
	// Check for case statement
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "case ") {
		return s.executeCaseStatement(lines)
	}

	// Check for if statement
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "if ") {
		return s.executeIfStatement(lines)
	}

	// If not a recognized multi-line construct, execute each line separately
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if err := s.ExecuteLine(line); err != nil {
			return fmt.Errorf("line %d: %v", i+1, err)
		}
	}

	return nil
}

// executeScriptLines processes script lines, handling multi-line constructs
func (s *Shell) executeScriptLines(allLines []string) error {
	lineNum := 0
	var fullLine strings.Builder
	startLineNum := 0

	for lineNum < len(allLines) {
		line := allLines[lineNum]
		lineNum++

		// Skip empty lines and comments (only if not in continuation)
		trimmed := strings.TrimSpace(line)
		if fullLine.Len() == 0 && (trimmed == "" || strings.HasPrefix(trimmed, "#")) {
			continue
		}

		// Check for function definition
		if fullLine.Len() == 0 && s.isFunctionDefinition(trimmed) {
			// Parse multi-line function definition
			funcLines, endLine, err := s.extractFunctionDefinition(allLines, lineNum-1)
			if err != nil {
				return fmt.Errorf("line %d: %v", lineNum, err)
			}

			// Define the function
			if err := s.defineFunction(funcLines); err != nil {
				return fmt.Errorf("line %d: %v", lineNum, err)
			}

			lineNum = endLine + 1
			continue
		}

		// Check for case statement
		if strings.HasPrefix(trimmed, "case ") {
			// If we have accumulated content, execute it first
			if fullLine.Len() > 0 {
				completeCommand := strings.TrimSpace(fullLine.String())
				if completeCommand != "" {
					if err := s.ExecuteLine(completeCommand); err != nil {
						return fmt.Errorf("line %d: %v", startLineNum, err)
					}
				}
				fullLine.Reset()
			}

			// Parse multi-line case statement
			caseLines, endLine, err := s.extractCaseStatement(allLines, lineNum-1)
			if err != nil {
				return fmt.Errorf("line %d: %v", lineNum, err)
			}

			// Execute case statement
			if err := s.executeCaseStatement(caseLines); err != nil {
				return fmt.Errorf("line %d: %v", lineNum, err)
			}

			lineNum = endLine + 1
			continue
		}

		// Check for if statement
		if strings.HasPrefix(trimmed, "if ") {
			// If we have accumulated content, execute it first
			if fullLine.Len() > 0 {
				completeCommand := strings.TrimSpace(fullLine.String())
				if completeCommand != "" {
					if err := s.ExecuteLine(completeCommand); err != nil {
						return fmt.Errorf("line %d: %v", startLineNum, err)
					}
				}
				fullLine.Reset()
			}

			// Parse multi-line if statement
			ifLines, endLine, err := s.extractIfStatement(allLines, lineNum-1)
			if err != nil {
				return fmt.Errorf("line %d: %v", lineNum, err)
			}

			// Execute if statement
			if err := s.executeIfStatement(ifLines); err != nil {
				return fmt.Errorf("line %d: %v", lineNum, err)
			}

			lineNum = endLine + 1
			continue
		}

		// If this is the start of a new command, record the line number
		if fullLine.Len() == 0 {
			startLineNum = lineNum
		}

		// Check if line ends with backslash (line continuation)
		trimmedRight := strings.TrimRightFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t'
		})

		if strings.HasSuffix(trimmedRight, "\\") {
			// Remove the backslash and continue reading
			continuedLine := strings.TrimSuffix(trimmedRight, "\\")
			fullLine.WriteString(continuedLine)
			fullLine.WriteString(" ") // Add space to separate continued lines
			continue
		} else {
			// No continuation, add this line and execute
			fullLine.WriteString(line)
			completeCommand := strings.TrimSpace(fullLine.String())

			if completeCommand != "" {
				if err := s.ExecuteLine(completeCommand); err != nil {
					return fmt.Errorf("line %d: %v", startLineNum, err)
				}
			}

			// Reset for next command
			fullLine.Reset()
		}
	}

	// Handle case where file ends with a continuation line
	if fullLine.Len() > 0 {
		completeCommand := strings.TrimSpace(fullLine.String())
		if completeCommand != "" {
			if err := s.ExecuteLine(completeCommand); err != nil {
				return fmt.Errorf("line %d: %v", startLineNum, err)
			}
		}
	}

	return nil
}

// extractCaseStatement extracts a complete case statement from script lines
func (s *Shell) extractCaseStatement(allLines []string, startLine int) ([]string, int, error) {
	var caseLines []string
	i := startLine

	for i < len(allLines) {
		line := strings.TrimSpace(allLines[i])
		caseLines = append(caseLines, allLines[i])

		if line == "esac" {
			return caseLines, i, nil
		}
		i++
	}

	return nil, i, fmt.Errorf("case statement not properly closed with 'esac'")
}

// extractIfStatement extracts a complete if statement from script lines
func (s *Shell) extractIfStatement(allLines []string, startLine int) ([]string, int, error) {
	var ifLines []string
	i := startLine
	ifCount := 0
	fiCount := 0

	for i < len(allLines) {
		line := strings.TrimSpace(allLines[i])
		ifLines = append(ifLines, allLines[i])

		if strings.HasPrefix(line, "if ") {
			ifCount++
		} else if line == "fi" {
			fiCount++
			if ifCount == fiCount {
				// Found matching fi
				return ifLines, i, nil
			}
		}
		i++
	}

	return nil, i, fmt.Errorf("if statement not properly closed with 'fi'")
}

// executeCaseStatement executes a case statement
func (s *Shell) executeCaseStatement(caseLines []string) error {
	commandChain, err := ParseCaseFromLines(caseLines)
	if err != nil {
		return err
	}

	if len(commandChain.Controls) == 0 {
		return fmt.Errorf("no case statement found")
	}

	caseStmt := commandChain.Controls[0].Case
	if caseStmt == nil {
		return fmt.Errorf("invalid case statement")
	}

	// Expand the variable
	expandedVar, err := s.expandToken(caseStmt.Variable)
	if err != nil {
		return fmt.Errorf("error expanding case variable: %v", err)
	}

	if len(expandedVar) != 1 {
		return fmt.Errorf("case variable expansion resulted in %d tokens, expected 1", len(expandedVar))
	}

	varValue := expandedVar[0]
	// Remove quotes if present in the expanded value
	if len(varValue) >= 2 && ((varValue[0] == '"' && varValue[len(varValue)-1] == '"') || (varValue[0] == '\'' && varValue[len(varValue)-1] == '\'')) {
		varValue = varValue[1 : len(varValue)-1]
	}

	// Match patterns and execute commands
	for _, pattern := range caseStmt.Patterns {
		for _, patternStr := range pattern.Patterns {
			if s.matchPattern(varValue, patternStr) {
				// Execute commands for this pattern
				for _, cmd := range pattern.Commands {
					// Expand the command before executing
					if err := s.expandCommand(cmd); err != nil {
						return err
					}
					if err := s.ExecuteCommand(cmd); err != nil {
						return err
					}
				}
				return nil // Exit after first match
			}
		}
	}

	return nil // No pattern matched
}

// executeIfStatement executes an if statement
func (s *Shell) executeIfStatement(ifLines []string) error {
	commandChain, err := ParseIfFromLines(ifLines)
	if err != nil {
		return err
	}

	if len(commandChain.Controls) == 0 {
		return fmt.Errorf("no if statement found")
	}

	ifStmt := commandChain.Controls[0].If
	if ifStmt == nil {
		return fmt.Errorf("invalid if statement")
	}

	// Execute the condition
	conditionPassed := false
	for _, condCmd := range ifStmt.Condition {
		// Expand the condition command before executing
		if err := s.expandCommand(condCmd); err != nil {
			return err
		}
		if err := s.ExecuteCommand(condCmd); err != nil {
			// If condition command fails, condition is false
			conditionPassed = false
			break
		} else {
			// If condition command succeeds, condition is true
			conditionPassed = true
		}
	}

	if conditionPassed {
		// Execute then commands
		for _, cmd := range ifStmt.ThenCommands {
			if err := s.expandCommand(cmd); err != nil {
				return err
			}
			if err := s.ExecuteCommand(cmd); err != nil {
				return err
			}
		}
	} else {
		// Check elif branches
		elifExecuted := false
		for _, elifBranch := range ifStmt.ElifBranches {
			elifConditionPassed := false
			for _, condCmd := range elifBranch.Condition {
				if err := s.expandCommand(condCmd); err != nil {
					return err
				}
				if err := s.ExecuteCommand(condCmd); err != nil {
					elifConditionPassed = false
					break
				} else {
					elifConditionPassed = true
				}
			}

			if elifConditionPassed {
				// Execute elif commands
				for _, cmd := range elifBranch.Commands {
					if err := s.expandCommand(cmd); err != nil {
						return err
					}
					if err := s.ExecuteCommand(cmd); err != nil {
						return err
					}
				}
				elifExecuted = true
				break
			}
		}

		// If no elif was executed, execute else commands
		if !elifExecuted {
			for _, cmd := range ifStmt.ElseCommands {
				if err := s.expandCommand(cmd); err != nil {
					return err
				}
				if err := s.ExecuteCommand(cmd); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// matchPattern checks if a value matches a shell pattern
func (s *Shell) matchPattern(value, pattern string) bool {
	// Expand variables in the pattern first
	expandedPattern := s.expandVariables(pattern)

	// Handle wildcard patterns
	matched, err := filepath.Match(expandedPattern, value)
	if err != nil {
		// If pattern matching fails, fall back to exact match
		return value == expandedPattern
	}
	return matched
}

// expandCommandChain applies expansions to all commands in a command chain
func (s *Shell) expandCommandChain(chain *CommandChain) error {
	for _, pipeline := range chain.Pipelines {
		for _, cmd := range pipeline {
			err := s.expandCommand(cmd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// expandCommand applies expansions to a single command
func (s *Shell) expandCommand(cmd *Command) error {
	// For builtin commands, don't apply glob expansion to the command name
	// Only expand variables and command substitutions
	if _, isBuiltin := builtins[cmd.Name]; isBuiltin {
		// Only expand variables and command substitutions, not globs
		expandedName, err := s.expandVariablesAndCommands(cmd.Name)
		if err != nil {
			return err
		}
		cmd.Name = expandedName
	} else {
		// For external commands, apply full expansion including globs
		expandedNames, err := s.expandToken(cmd.Name)
		if err != nil {
			return err
		}
		if len(expandedNames) != 1 {
			return fmt.Errorf("command name expansion resulted in %d tokens, expected 1", len(expandedNames))
		}
		cmd.Name = expandedNames[0]
	}

	// Expand arguments
	expandedArgs, err := s.expandTokens(cmd.Args)
	if err != nil {
		return err
	}
	cmd.Args = expandedArgs

	// Expand input redirection
	if cmd.Input != "" {
		expandedInput, err := s.expandToken(cmd.Input)
		if err != nil {
			return err
		}
		if len(expandedInput) != 1 {
			return fmt.Errorf("input redirection expansion resulted in %d tokens, expected 1", len(expandedInput))
		}
		cmd.Input = expandedInput[0]
	}

	// Expand output redirection
	if cmd.Output != "" {
		expandedOutput, err := s.expandToken(cmd.Output)
		if err != nil {
			return err
		}
		if len(expandedOutput) != 1 {
			return fmt.Errorf("output redirection expansion resulted in %d tokens, expected 1", len(expandedOutput))
		}
		cmd.Output = expandedOutput[0]
	}

	return nil
}

// readLineWithContinuation reads a command line with support for backslash line continuation
// and multi-line constructs like if and case statements
// abbreviatePath converts a full path to abbreviated form
// Example: /Users/drewwalton/Projects/go/gosh -> /U/d/P/g/gosh
func (s *Shell) abbreviatePath(path string) string {
	if path == "/" {
		return "/"
	}

	// Split the path into components
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) <= 1 {
		return path
	}

	// Abbreviate all parts except the last one
	var abbreviated []string
	for i, part := range parts {
		if i == len(parts)-1 {
			// Keep the last directory name in full
			abbreviated = append(abbreviated, part)
		} else if len(part) > 0 {
			// Abbreviate to first character
			abbreviated = append(abbreviated, string(part[0]))
		}
	}

	return "/" + strings.Join(abbreviated, "/")
}

func (s *Shell) readLineWithContinuation(cwd string) (string, error) {
	var fullLine strings.Builder
	isFirstLine := true

	for {
		// Set appropriate prompt
		if isFirstLine {
			abbreviatedPath := s.abbreviatePath(cwd)
			s.rl.SetPrompt(fmt.Sprintf("gosh:%s> ", abbreviatedPath))
		} else {
			s.rl.SetPrompt("> ")
		}

		line, err := s.rl.Readline()
		if err != nil {
			// Handle Ctrl-C (interrupt) - don't exit, just return empty line to continue
			if err == io.EOF || err == readline.ErrInterrupt {
				if fullLine.Len() > 0 {
					// If we have partial input, clear it and start fresh
					fullLine.Reset()
					isFirstLine = true
					continue
				}
				// Return empty string to continue the shell loop
				return "", nil
			}
			return "", err
		}

		// Check if line ends with backslash (line continuation)
		trimmed := strings.TrimRightFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t'
		})

		if strings.HasSuffix(trimmed, "\\") && s.isLineContinuationBackslash(trimmed) {
			// Remove the backslash and continue reading
			continuedLine := strings.TrimSuffix(trimmed, "\\")
			fullLine.WriteString(continuedLine)
			fullLine.WriteString(" ") // Add space to separate continued lines
			isFirstLine = false
			continue
		} else {
			// Add this line
			fullLine.WriteString(line)

			// Check if we need to read more lines for multi-line constructs
			currentContent := fullLine.String()
			if s.needsMoreLines(currentContent) {
				fullLine.WriteString("\n")
				isFirstLine = false
				continue
			}

			break
		}
	}

	return fullLine.String(), nil
}

// needsMoreLines determines if the current input requires more lines to complete
// a multi-line construct like if or case statements
func (s *Shell) needsMoreLines(input string) bool {
	lines := strings.Split(input, "\n")

	// Check for incomplete if statement
	if s.isIncompleteIfStatement(lines) {
		return true
	}

	// Check for incomplete case statement
	if s.isIncompleteCaseStatement(lines) {
		return true
	}

	return false
}

// isIncompleteIfStatement checks if we have an incomplete if statement
func (s *Shell) isIncompleteIfStatement(lines []string) bool {
	ifCount := 0
	fiCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "if ") {
			ifCount++
		} else if trimmed == "fi" {
			fiCount++
		}
	}

	return ifCount > fiCount
}

// isIncompleteCaseStatement checks if we have an incomplete case statement
func (s *Shell) isIncompleteCaseStatement(lines []string) bool {
	caseCount := 0
	esacCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "case ") {
			caseCount++
		} else if trimmed == "esac" {
			esacCount++
		}
	}

	return caseCount > esacCount
}

// isLineContinuationBackslash determines if a trailing backslash is for line continuation
// or part of a Windows path. Returns true only if it's likely a line continuation.
func (s *Shell) isLineContinuationBackslash(line string) bool {
	if !strings.HasSuffix(line, "\\") {
		return false
	}
	
	// If the line is just a backslash, it's continuation
	if line == "\\" {
		return true
	}
	
	// Get the character before the backslash
	if len(line) < 2 {
		return true
	}
	
	prevChar := line[len(line)-2]
	
	// If preceded by a space or tab, it's likely line continuation
	if prevChar == ' ' || prevChar == '\t' {
		return true
	}
	
	// If preceded by alphanumeric, dot, tilde, colon, or another backslash,
	// it's likely a path separator (especially on Windows)
	if (prevChar >= 'a' && prevChar <= 'z') ||
		(prevChar >= 'A' && prevChar <= 'Z') ||
		(prevChar >= '0' && prevChar <= '9') ||
		prevChar == '.' || prevChar == '~' || prevChar == ':' || prevChar == '\\' {
		return false
	}
	
	// For other characters, assume it's line continuation
	return true
}

// Exit sets the shell to stop running
func (s *Shell) Exit(code int) {
	s.running = false
}

// createCompleter creates a tab completion function for readline
func (s *Shell) createCompleter() readline.AutoCompleter {
	return &TabCompleter{shell: s}
}

// TabCompleter implements readline.AutoCompleter interface
type TabCompleter struct {
	shell *Shell
}

// Do implements the AutoCompleter interface
func (tc *TabCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line)
	completions := tc.shell.getCompletions(lineStr, pos)

	if len(completions) == 0 {
		return nil, 0
	}

	// Find the current word being completed
	fields := strings.Fields(lineStr[:pos])
	currentWord := ""

	// Get the word being completed
	if pos > 0 && !strings.HasSuffix(lineStr[:pos], " ") {
		if len(fields) > 0 {
			currentWord = fields[len(fields)-1]
		}
	}

	// Convert completions to [][]rune format expected by readline
	// We need to return only the suffix that completes the current word
	var result [][]rune
	for _, completion := range completions {
		// Only return the part that extends beyond the current word
		if strings.HasPrefix(completion, currentWord) {
			suffix := completion[len(currentWord):]
			result = append(result, []rune(suffix))
		} else {
			// Fallback: return the full completion
			result = append(result, []rune(completion))
		}
	}

	return result, 0 // Return 0 for length since we're providing suffixes
}

// getCompletions returns completion suggestions for the given input
func (s *Shell) getCompletions(line string, pos int) []string {
	// Parse the current line to understand context
	fields := strings.Fields(line[:pos])
	currentWord := ""

	// Get the word being completed
	if pos > 0 && !strings.HasSuffix(line[:pos], " ") {
		if len(fields) > 0 {
			currentWord = fields[len(fields)-1]
			fields = fields[:len(fields)-1]
		}
	}

	var completions []string

	// Check if the current word looks like a file path (contains / or starts with .)
	isFilePath := strings.Contains(currentWord, "/") || strings.HasPrefix(currentWord, ".")

	if len(fields) == 0 && !isFilePath {
		// Completing command name (but not if it looks like a file path)
		completions = append(completions, s.getCommandCompletions(currentWord)...)
	} else {
		// Completing arguments or file paths - provide file/directory completions
		completions = append(completions, s.getFileCompletions(currentWord)...)
	}

	// If we're completing a command name and no command completions were found,
	// also try file completions (for executable files)
	if len(fields) == 0 && !isFilePath && len(completions) == 0 {
		fileCompletions := s.getFileCompletions(currentWord)
		completions = append(completions, fileCompletions...)
	}

	return completions
}

// getCommandCompletions returns command name completions
func (s *Shell) getCommandCompletions(prefix string) []string {
	var completions []string

	// Add built-in commands
	for cmd := range builtins {
		if matchesPrefix(cmd, prefix) {
			completion := constructCompletion(cmd, prefix)
			completions = append(completions, completion)
		}
	}

	// Add aliases
	for alias := range s.aliases {
		if matchesPrefix(alias, prefix) {
			completion := constructCompletion(alias, prefix)
			completions = append(completions, completion)
		}
	}

	// Add external commands from PATH
	pathCompletions := s.getPathCompletions(prefix)
	completions = append(completions, pathCompletions...)

	// Sort and remove duplicates
	sort.Strings(completions)
	return removeDuplicates(completions)
}

// getPathCompletions returns external command completions from PATH
func (s *Shell) getPathCompletions(prefix string) []string {
	var completions []string
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return completions
	}

	paths := strings.Split(pathEnv, ":")
	seenCommands := make(map[string]bool)

	for _, dir := range paths {
		if dir == "" {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()
			if !matchesPrefix(name, prefix) {
				continue
			}

			// Check if it's executable
			if entry.IsDir() {
				continue
			}

			filePath := filepath.Join(dir, name)
			if info, err := os.Stat(filePath); err == nil {
				if info.Mode()&0111 != 0 { // Check if executable
					if !seenCommands[name] {
						completion := constructCompletion(name, prefix)
						completions = append(completions, completion)
						seenCommands[name] = true
					}
				}
			}
		}
	}

	return completions
}

// containsPathSeparator checks if a path contains any path separator (/ or \)
func containsPathSeparator(path string) bool {
	return strings.ContainsAny(path, "/\\")
}

// getPathSeparator returns the path separator used in the given path
// Returns the first separator found, defaulting to forward slash
func getPathSeparator(path string) string {
	if strings.Contains(path, "\\") {
		return "\\"
	}
	return "/"
}

// joinPathWithSeparator joins path components using the specified separator
func joinPathWithSeparator(dir, file, separator string) string {
	if separator == "\\" {
		// Use filepath.Join for Windows-style paths to handle edge cases
		return filepath.Join(dir, file)
	}
	// Use forward slash for Unix-style paths
	if dir == "" || dir == "." {
		return file
	}
	return dir + "/" + file
}

// shouldUseCaseInsensitiveCompletion returns true if case-insensitive completion should be used
// Case-insensitive completion is only enabled on Windows
func shouldUseCaseInsensitiveCompletion() bool {
	return runtime.GOOS == "windows"
}

// matchesPrefix checks if name matches prefix, using case-sensitive or case-insensitive matching based on platform
func matchesPrefix(name, prefix string) bool {
	if shouldUseCaseInsensitiveCompletion() {
		return strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix))
	}
	return strings.HasPrefix(name, prefix)
}

// constructCompletion constructs a completion string, preserving user's case on Windows
func constructCompletion(name, prefix string) string {
	if shouldUseCaseInsensitiveCompletion() && matchesPrefix(name, prefix) {
		// Case-insensitive match - construct completion that starts with original prefix
		return prefix + name[len(prefix):]
	}
	// Exact match or case-sensitive platform
	return name
}

// constructPathCompletion constructs a path completion string, handling case-insensitive matching
func constructPathCompletion(name, prefixBase, prefixDir, userSeparator string, filePrefix string) string {
	if filePrefix != "" && matchesPrefix(name, prefixBase) {
		if shouldUseCaseInsensitiveCompletion() {
			// Case-insensitive match - construct completion that starts with original prefix
			return prefixDir + userSeparator + prefixBase + name[len(prefixBase):]
		} else {
			// Case-sensitive match
			return prefixDir + userSeparator + name
		}
	} else {
		// Exact match or directory listing
		return prefixDir + userSeparator + name
	}
}

// constructSimplePathCompletion constructs a simple path completion for cases like "./" or simple filenames
func constructSimplePathCompletion(name, prefix, pathPrefix string, filePrefix string) string {
	if filePrefix != "" && matchesPrefix(name, prefix) {
		if shouldUseCaseInsensitiveCompletion() {
			// Case-insensitive match - construct completion that starts with original prefix
			return pathPrefix + prefix + name[len(prefix):]
		} else {
			// Case-sensitive match
			return pathPrefix + name
		}
	} else {
		// Exact match
		return pathPrefix + name
	}
}

// getFileCompletions returns file and directory completions
func (s *Shell) getFileCompletions(prefix string) []string {
	var completions []string

	// Check if the prefix contains glob patterns
	if strings.ContainsAny(prefix, "*?[]") {
		return s.getGlobCompletions(prefix)
	}

	// Expand tilde in the prefix first
	expandedPrefix := prefix
	if strings.HasPrefix(prefix, "~") {
		expanded, err := s.expandTilde(prefix)
		if err == nil {
			expandedPrefix = expanded
		}
	}

	// Handle different path types
	var searchDir, filePrefix string

	// Special case: if the original prefix is just "~", we want to complete to "~/"
	if prefix == "~" {
		// Return just "~/" as the completion
		return []string{"~/"}
	}

	// Special case: if the original prefix is "~/", don't expand further
	if prefix == "~/" {
		// Return empty completions to avoid expanding to ~/username/
		return []string{}
	}

	// Special case: if the original prefix is "~user" (without trailing separator),
	// complete it to "~user/" instead of listing files from the user's directory
	// But don't apply this to patterns like "~user/path" that already have a path component
	if strings.HasPrefix(prefix, "~") && !strings.HasSuffix(prefix, "/") && !strings.HasSuffix(prefix, "\\") && prefix != expandedPrefix {
		// Check if this is just ~user without any path component
		// Count separators after the ~
		separatorCount := 0
		for _, char := range prefix[1:] {
			if char == '/' || char == '\\' {
				separatorCount++
			}
		}
		
		// Only apply this special case if there are no separators (just ~user)
		if separatorCount == 0 {
			// This is a ~user pattern that was successfully expanded
			// Complete it to ~user/ instead of listing directory contents
			userSeparator := getPathSeparator(prefix)
			if userSeparator == "\\" {
				return []string{prefix + "\\"}
			} else {
				return []string{prefix + "/"}
			}
		}
	}

	if containsPathSeparator(expandedPrefix) {
		// Check if the prefix ends with a separator - this means we want to list contents of that directory
		if strings.HasSuffix(expandedPrefix, "/") || strings.HasSuffix(expandedPrefix, "\\") {
			// User wants to see contents of the directory, not complete the directory name
			searchDir = expandedPrefix
			filePrefix = ""
		} else {
			// Normal path completion - extract directory and file prefix
			searchDir = filepath.Dir(expandedPrefix)
			filePrefix = filepath.Base(expandedPrefix)
		}

		// Handle absolute vs relative paths
		if !filepath.IsAbs(searchDir) {
			cwd, err := os.Getwd()
			if err != nil {
				return completions
			}
			searchDir = filepath.Join(cwd, searchDir)
		}
	} else {
		// No directory separator - search current directory
		cwd, err := os.Getwd()
		if err != nil {
			return completions
		}
		searchDir = cwd
		filePrefix = expandedPrefix
	}

	// Read directory entries
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return completions
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless prefix starts with dot
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(filePrefix, ".") {
			continue
		}

		// If filePrefix is empty, we're listing directory contents, so include all files
		// Otherwise, check if the name matches the prefix
		if filePrefix == "" || matchesPrefix(name, filePrefix) {
			var completion string

			// Detect the path separator style used by the user
			userSeparator := getPathSeparator(prefix)

			// Handle tilde expansion case first
			if strings.HasPrefix(prefix, "~") && prefix != expandedPrefix {
				// We expanded a tilde, so we need to reconstruct with the original tilde prefix
				if prefix == "~" {
					completion = "~/" + name
				} else if strings.HasPrefix(prefix, "~/") {
					// Handle ~/ case
					if strings.HasSuffix(prefix, "/") {
						// Directory listing case - we're already in the directory, so just use the tilde path up to the last separator + filename
						// For ~/Projects/, we want ~/Projects/filename, not ~/Projects/Projects
						completion = prefix + name
					} else {
						// For cases like ~/PR completing to ~/Projects, use the correct filesystem case
						prefixDir := filepath.Dir(prefix)
						if prefixDir == "." {
							// prefix is just ~/something without slashes
							prefixBase := filepath.Base(prefix)
							if filePrefix != "" && matchesPrefix(name, prefixBase) {
								if shouldUseCaseInsensitiveCompletion() {
									// Case-insensitive match - construct completion that starts with original prefix
									completion = "~/" + prefixBase + name[len(prefixBase):]
								} else {
									// Case-sensitive match
									completion = "~/" + name
								}
							} else {
								// Exact match or directory listing
								completion = "~/" + name
							}
						} else {
							// Handle cases like ~/Documents/PR completing to ~/Documents/Projects
							prefixBase := filepath.Base(prefix)
							if filePrefix != "" && matchesPrefix(name, prefixBase) {
								if shouldUseCaseInsensitiveCompletion() {
									// Case-insensitive match - construct completion that starts with original prefix
									completion = prefixDir + userSeparator + prefixBase + name[len(prefixBase):]
								} else {
									// Case-sensitive match
									completion = prefixDir + userSeparator + name
								}
							} else {
								// Exact match or directory listing
								completion = prefixDir + userSeparator + name
							}
						}
					}
				} else {
					// Handle ~user case
					if strings.HasSuffix(prefix, "/") || strings.HasSuffix(prefix, "\\") {
						// Directory listing case
						completion = prefix + name
					} else {
						// Check for case-insensitive prefix matching
						prefixBase := filepath.Base(prefix)
						if filePrefix != "" && matchesPrefix(name, prefixBase) {
							if shouldUseCaseInsensitiveCompletion() {
								// Case-insensitive match - complete the full name
								prefixDir := filepath.Dir(prefix)
								if prefixDir == "." {
									// Simple ~user/filename case
									completion = strings.TrimSuffix(prefix, prefixBase) + prefixBase + name[len(prefixBase):]
								} else {
									// ~user/path/filename case
									completion = prefixDir + userSeparator + prefixBase + name[len(prefixBase):]
								}
							} else {
								// Case-sensitive match
								completion = prefix + userSeparator + name
							}
						} else {
							// Exact match or no prefix match
							completion = prefix + userSeparator + name
						}
					}
				}
			} else if containsPathSeparator(prefix) {
				// Check if we're doing directory listing (prefix ends with separator)
				if strings.HasSuffix(prefix, "/") || strings.HasSuffix(prefix, "\\") {
					// Directory listing - just append the filename to the prefix
					completion = prefix + name
				} else {
					// Reconstruct the full path, preserving the original prefix format
					prefixDir := filepath.Dir(prefix)
					prefixBase := filepath.Base(prefix)
					if prefixDir == "." && strings.HasPrefix(prefix, "./") {
						// Preserve './' prefix format
						if filePrefix != "" && matchesPrefix(name, prefixBase) {
							if shouldUseCaseInsensitiveCompletion() {
								// Case-insensitive match
								completion = "./" + prefixBase + name[len(prefixBase):]
							} else {
								// Case-sensitive match
								completion = "./" + name
							}
						} else {
							// Exact match
							completion = "./" + name
						}
					} else {
						// Use the correct filesystem case but preserve prefix structure and separator style
						if filePrefix != "" && matchesPrefix(name, prefixBase) {
							if shouldUseCaseInsensitiveCompletion() {
								// Case-insensitive match - preserve user's separator style
								completion = joinPathWithSeparator(prefixDir, prefixBase+name[len(prefixBase):], userSeparator)
							} else {
								// Case-sensitive match - preserve user's separator style
								completion = joinPathWithSeparator(prefixDir, name, userSeparator)
							}
						} else {
							// Exact match - preserve user's separator style
							completion = joinPathWithSeparator(prefixDir, name, userSeparator)
						}
					}
				}
			} else {
				// Simple filename completion
				if filePrefix != "" && matchesPrefix(name, prefix) {
					if shouldUseCaseInsensitiveCompletion() {
						// Case-insensitive match - construct completion that starts with original prefix
						completion = prefix + name[len(prefix):]
					} else {
						// Case-sensitive match
						completion = name
					}
				} else {
					// Exact match
					completion = name
				}
			}

			// Add trailing slash for directories
			if entry.IsDir() {
				completion += userSeparator
			}

			completions = append(completions, completion)
		}
	}

	return completions
}

// getGlobCompletions handles glob pattern completions
func (s *Shell) getGlobCompletions(pattern string) []string {
	var completions []string

	// Expand tilde in the pattern first
	expandedPattern := pattern
	if strings.HasPrefix(pattern, "~") {
		expanded, err := s.expandTilde(pattern)
		if err == nil {
			expandedPattern = expanded
		}
	}

	// Use filepath.Glob to expand the pattern
	var globPattern string
	if filepath.IsAbs(expandedPattern) {
		globPattern = expandedPattern
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return completions
		}
		globPattern = filepath.Join(cwd, expandedPattern)
	}

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return completions
	}

	// Convert absolute paths back to relative if needed
	for _, match := range matches {
		var completion string
		if containsPathSeparator(pattern) {
			// Keep the path structure from the original pattern
			if filepath.IsAbs(pattern) {
				completion = match
			} else if strings.HasPrefix(pattern, "~") && pattern != expandedPattern {
				// Handle tilde expansion case - convert back to tilde format
				if strings.HasPrefix(pattern, "~/") {
					home, err := os.UserHomeDir()
					if err == nil && strings.HasPrefix(match, home) {
						completion = "~" + match[len(home):]
					} else {
						completion = match
					}
				} else {
					// Handle ~user case - this is more complex, for now just return the match
					completion = match
				}
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					continue
				}
				relPath, err := filepath.Rel(cwd, match)
				if err != nil {
					completion = match
				} else {
					// Preserve the user's original path separator style
					userSeparator := getPathSeparator(pattern)
					if userSeparator == "\\" && strings.Contains(relPath, "/") {
						completion = strings.ReplaceAll(relPath, "/", "\\")
					} else {
						completion = relPath
					}
				}
			}
		} else {
			// Just the filename
			completion = filepath.Base(match)
		}

		// Add trailing slash for directories
		if info, err := os.Stat(match); err == nil && info.IsDir() {
			userSeparator := getPathSeparator(pattern)
			completion += userSeparator
		}

		completions = append(completions, completion)
	}

	return completions
}

// removeDuplicates removes duplicate strings from a sorted slice
func removeDuplicates(strs []string) []string {
	if len(strs) <= 1 {
		return strs
	}

	result := make([]string, 0, len(strs))
	result = append(result, strs[0])

	for i := 1; i < len(strs); i++ {
		if strs[i] != strs[i-1] {
			result = append(result, strs[i])
		}
	}

	return result
}

// isFunctionDefinition checks if a line starts a function definition
func (s *Shell) isFunctionDefinition(line string) bool {
	// Function definitions can be in two formats:
	// 1. function_name() {
	// 2. function function_name() {
	line = strings.TrimSpace(line)

	// Check for "function name() {" format
	if strings.HasPrefix(line, "function ") {
		rest := strings.TrimPrefix(line, "function ")
		return strings.Contains(rest, "()") && strings.HasSuffix(strings.TrimSpace(rest), "{")
	}

	// Check for "name() {" format
	return strings.Contains(line, "()") && strings.HasSuffix(strings.TrimSpace(line), "{")
}

// extractFunctionDefinition extracts a complete function definition from script lines
func (s *Shell) extractFunctionDefinition(allLines []string, startLine int) ([]string, int, error) {
	var funcLines []string
	i := startLine
	braceCount := 0

	for i < len(allLines) {
		line := allLines[i]
		funcLines = append(funcLines, line)

		// Count braces to find the end of the function
		for _, char := range line {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					return funcLines, i, nil
				}
			}
		}

		i++
	}

	return nil, i, fmt.Errorf("function definition not properly closed with '}'")
}

// defineFunction parses and stores a function definition
func (s *Shell) defineFunction(funcLines []string) error {
	if len(funcLines) == 0 {
		return fmt.Errorf("empty function definition")
	}

	firstLine := strings.TrimSpace(funcLines[0])
	var funcName string
	var params []string

	// Parse function name and parameters
	if strings.HasPrefix(firstLine, "function ") {
		// "function name() {" format
		rest := strings.TrimPrefix(firstLine, "function ")
		parenIndex := strings.Index(rest, "()")
		if parenIndex == -1 {
			return fmt.Errorf("invalid function syntax: missing ()")
		}
		funcName = strings.TrimSpace(rest[:parenIndex])
	} else {
		// "name() {" format
		parenIndex := strings.Index(firstLine, "()")
		if parenIndex == -1 {
			return fmt.Errorf("invalid function syntax: missing ()")
		}
		funcName = strings.TrimSpace(firstLine[:parenIndex])
	}

	if funcName == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	// Extract function body (everything except first and last lines)
	var body []string
	for i := 1; i < len(funcLines)-1; i++ {
		body = append(body, funcLines[i])
	}

	// Store the function
	s.functions[funcName] = &Function{
		Name:   funcName,
		Params: params,
		Body:   body,
	}

	return nil
}

// executeFunction executes a user-defined function
func (s *Shell) executeFunction(funcName string, args []string) error {
	func_, exists := s.functions[funcName]
	if !exists {
		return fmt.Errorf("function not found: %s", funcName)
	}

	// Create function context
	ctx := &FunctionContext{
		Name: funcName,
		Args: args,
	}

	// Push context onto stack
	s.functionStack = append(s.functionStack, ctx)
	defer func() {
		// Pop context when function exits
		if len(s.functionStack) > 0 {
			s.functionStack = s.functionStack[:len(s.functionStack)-1]
		}
	}()

	// Execute function body
	err := s.executeScriptLines(func_.Body)

	// Handle return statements
	if returnErr, ok := err.(ReturnError); ok {
		// Return statements are normal function exits, not errors
		// For now, we'll just return nil (success) regardless of return code
		// In a full shell implementation, you might want to set an exit status
		_ = returnErr.Code // Use the return code if needed
		return nil
	}

	return err
}

// isVariableAssignment checks if a line is a variable assignment (VAR=value)
func (s *Shell) isVariableAssignment(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Must contain an equals sign
	if !strings.Contains(trimmed, "=") {
		return false
	}

	// Find the first equals sign
	eqIndex := strings.Index(trimmed, "=")
	if eqIndex == 0 {
		return false // Can't start with =
	}

	// Extract variable name (everything before =)
	varName := trimmed[:eqIndex]

	// Variable name must be valid (alphanumeric + underscore, can't start with digit)
	if len(varName) == 0 {
		return false
	}

	// Check if it's a valid variable name
	for i, char := range varName {
		if i == 0 {
			// First character: must be letter or underscore
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_') {
				return false
			}
		} else {
			// Subsequent characters: letter, digit, or underscore
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
				return false
			}
		}
	}

	return true
}

// handleVariableAssignment processes a variable assignment
func (s *Shell) handleVariableAssignment(line string) error {
	trimmed := strings.TrimSpace(line)

	// Find the first equals sign
	eqIndex := strings.Index(trimmed, "=")
	varName := trimmed[:eqIndex]

	// Reject bare variable assignments - require 'local' or 'export'
	return fmt.Errorf("variable assignment without declaration: %s\nUse 'local %s' for local variables or 'export %s' for environment variables", varName, trimmed, trimmed)
}

// getHistoryFilePath returns a cross-platform path for the history file
func getHistoryFilePath() string {
	// Try to get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not available
		return ".gosh_history"
	}

	// Use home directory with hidden file
	return filepath.Join(homeDir, ".gosh_history")
}
