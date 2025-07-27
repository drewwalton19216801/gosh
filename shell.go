package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

// Shell represents the main shell instance
type Shell struct {
	env     map[string]string
	aliases map[string]string
	history []string
	running bool
	rl      *readline.Instance
}

// NewShell creates a new shell instance
func NewShell() *Shell {
	s := &Shell{
		env:     make(map[string]string),
		aliases: make(map[string]string),
		history: make([]string, 0),
		running: true,
	}

	// Initialize readline with history support and tab completion
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "gosh> ",
		HistoryFile:     "/tmp/.gosh_history",
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
		} else {
			cwd = filepath.Base(cwd)
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

// ExecuteLine processes a single command line
func (s *Shell) ExecuteLine(line string) error {
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

// ExecuteScript runs a script file
func (s *Shell) ExecuteScript(filename string) error {
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

	return s.executeScriptLines(allLines)
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

		// Check for case statement
		if fullLine.Len() == 0 && strings.HasPrefix(trimmed, "case ") {
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
	// Expand command name
	expandedNames, err := s.expandToken(cmd.Name)
	if err != nil {
		return err
	}
	if len(expandedNames) != 1 {
		return fmt.Errorf("command name expansion resulted in %d tokens, expected 1", len(expandedNames))
	}
	cmd.Name = expandedNames[0]

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
func (s *Shell) readLineWithContinuation(cwd string) (string, error) {
	var fullLine strings.Builder
	isFirstLine := true
	
	for {
		// Set appropriate prompt
		if isFirstLine {
			s.rl.SetPrompt(fmt.Sprintf("gosh:%s> ", cwd))
		} else {
			s.rl.SetPrompt("> ")
		}
		
		line, err := s.rl.Readline()
		if err != nil {
			return "", err
		}
		
		// Check if line ends with backslash (line continuation)
		trimmed := strings.TrimRightFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t'
		})
		
		if strings.HasSuffix(trimmed, "\\") {
			// Remove the backslash and continue reading
			continuedLine := strings.TrimSuffix(trimmed, "\\")
			fullLine.WriteString(continuedLine)
			fullLine.WriteString(" ") // Add space to separate continued lines
			isFirstLine = false
			continue
		} else {
			// No continuation, add this line and finish
			fullLine.WriteString(line)
			break
		}
	}
	
	return fullLine.String(), nil
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

	if len(fields) == 0 {
		// Completing command name
		completions = append(completions, s.getCommandCompletions(currentWord)...)
	} else {
		// Completing arguments - provide file/directory completions
		completions = append(completions, s.getFileCompletions(currentWord)...)
	}

	return completions
}

// getCommandCompletions returns command name completions
func (s *Shell) getCommandCompletions(prefix string) []string {
	var completions []string

	// Add built-in commands
	for cmd := range builtins {
		if strings.HasPrefix(cmd, prefix) {
			completions = append(completions, cmd)
		}
	}

	// Add aliases
	for alias := range s.aliases {
		if strings.HasPrefix(alias, prefix) {
			completions = append(completions, alias)
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
			if !strings.HasPrefix(name, prefix) {
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
						completions = append(completions, name)
						seenCommands[name] = true
					}
				}
			}
		}
	}

	return completions
}

// getFileCompletions returns file and directory completions
func (s *Shell) getFileCompletions(prefix string) []string {
	var completions []string

	// Check if the prefix contains glob patterns
	if strings.ContainsAny(prefix, "*?[]") {
		return s.getGlobCompletions(prefix)
	}

	// Handle different path types
	var searchDir, filePrefix string

	if strings.Contains(prefix, "/") {
		// Path contains directory separator
		searchDir = filepath.Dir(prefix)
		filePrefix = filepath.Base(prefix)

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
		filePrefix = prefix
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

		if strings.HasPrefix(name, filePrefix) {
			var completion string
			if strings.Contains(prefix, "/") {
				// Reconstruct the full path
				completion = filepath.Join(filepath.Dir(prefix), name)
			} else {
				completion = name
			}

			// Add trailing slash for directories
			if entry.IsDir() {
				completion += "/"
			}

			completions = append(completions, completion)
		}
	}

	return completions
}

// getGlobCompletions handles glob pattern completions
func (s *Shell) getGlobCompletions(pattern string) []string {
	var completions []string

	// Use filepath.Glob to expand the pattern
	var globPattern string
	if filepath.IsAbs(pattern) {
		globPattern = pattern
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return completions
		}
		globPattern = filepath.Join(cwd, pattern)
	}

	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return completions
	}

	// Convert absolute paths back to relative if needed
	for _, match := range matches {
		var completion string
		if strings.Contains(pattern, "/") {
			// Keep the path structure from the original pattern
			if filepath.IsAbs(pattern) {
				completion = match
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					continue
				}
				relPath, err := filepath.Rel(cwd, match)
				if err != nil {
					completion = match
				} else {
					completion = relPath
				}
			}
		} else {
			// Just the filename
			completion = filepath.Base(match)
		}

		// Add trailing slash for directories
		if info, err := os.Stat(match); err == nil && info.IsDir() {
			completion += "/"
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
