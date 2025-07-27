package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Command represents a parsed command with arguments
type Command struct {
	Name string
	Args []string
	Input string
	Output string
	Append bool
	Background bool
}

// CasePattern represents a pattern in a case statement
type CasePattern struct {
	Patterns []string // Multiple patterns separated by |
	Commands []*Command
}

// CaseStatement represents a case control structure
type CaseStatement struct {
	Variable string
	Patterns []*CasePattern
}

// IfStatement represents an if control structure
type IfStatement struct {
	Condition []*Command
	ThenCommands []*Command
	ElseCommands []*Command
	ElifBranches []*ElifBranch
}

// ElifBranch represents an elif branch in an if statement
type ElifBranch struct {
	Condition []*Command
	Commands []*Command
}

// ControlStructure represents different control flow structures
type ControlStructure struct {
	Type string // "case" or "if"
	Case *CaseStatement
	If *IfStatement
}

// CommandChain represents a sequence of command pipelines and control structures
type CommandChain struct {
	Pipelines [][]*Command
	Controls []*ControlStructure
}

// ParseLine parses a command line into command chains (semicolon-separated pipelines)
func ParseLine(line string) (*CommandChain, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	// Check if this is a case statement
	if strings.HasPrefix(line, "case ") {
		return parseCaseStatement(line)
	}

	// Check if this is an if statement
	if strings.HasPrefix(line, "if ") {
		return parseIfStatement(line)
	}

	// Split by semicolons first to handle command chaining
	chainSegments, err := splitBySemicolons(line)
	if err != nil {
		return nil, err
	}

	var pipelines [][]*Command
	for _, segment := range chainSegments {
		// For each segment, split by pipes
		pipeSegments, err := splitByPipes(segment)
		if err != nil {
			return nil, err
		}

		var commands []*Command
		for _, pipeSegment := range pipeSegments {
			cmd, err := parseCommand(pipeSegment)
			if err != nil {
				return nil, err
			}
			commands = append(commands, cmd)
		}
		pipelines = append(pipelines, commands)
	}

	return &CommandChain{Pipelines: pipelines}, nil
}

// parseCommand parses a single command string
func parseCommand(cmdStr string) (*Command, error) {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return nil, errors.New("empty command")
	}

	cmd := &Command{}
	tokens, err := tokenize(cmdStr)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, errors.New("no command specified")
	}

	// Check for background execution
	if len(tokens) > 0 && tokens[len(tokens)-1] == "&" {
		cmd.Background = true
		tokens = tokens[:len(tokens)-1]
	}

	// Parse redirections and command name/args
	i := 0
	for i < len(tokens) {
		token := tokens[i]
		
		switch token {
		case ">":
			if i+1 >= len(tokens) {
				return nil, errors.New("missing output file")
			}
			cmd.Output = tokens[i+1]
			cmd.Append = false
			i += 2
		case ">>":
			if i+1 >= len(tokens) {
				return nil, errors.New("missing output file")
			}
			cmd.Output = tokens[i+1]
			cmd.Append = true
			i += 2
		case "<":
			if i+1 >= len(tokens) {
				return nil, errors.New("missing input file")
			}
			cmd.Input = tokens[i+1]
			i += 2
		default:
			if cmd.Name == "" {
				cmd.Name = token
			} else {
				cmd.Args = append(cmd.Args, token)
			}
			i++
		}
	}

	if cmd.Name == "" {
		return nil, errors.New("no command specified")
	}

	return cmd, nil
}

// splitBySemicolons splits a command line by semicolon characters, respecting quotes
func splitBySemicolons(line string) ([]string, error) {
	var segments []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	escaped := false

	for i := 0; i < len(line); i++ {
		c := line[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			current.WriteByte(c)
			continue
		}

		if inQuotes {
			if c == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(c)
		} else {
			if c == '"' || c == '\'' {
				inQuotes = true
				quoteChar = c
				current.WriteByte(c)
			} else if c == ';' {
				// Found a semicolon - split here
				segment := strings.TrimSpace(current.String())
				if segment == "" {
					return nil, errors.New("empty command before semicolon")
				}
				segments = append(segments, segment)
				current.Reset()
			} else {
				current.WriteByte(c)
			}
		}
	}

	if inQuotes {
		return nil, errors.New("unclosed quote in command chain")
	}

	// Add the last segment
	lastSegment := strings.TrimSpace(current.String())
	if lastSegment == "" {
		return nil, errors.New("empty command after semicolon")
	}
	segments = append(segments, lastSegment)

	return segments, nil
}

// splitByPipes splits a command line by pipe characters, respecting quotes
func splitByPipes(line string) ([]string, error) {
	var segments []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	escaped := false

	for i := 0; i < len(line); i++ {
		c := line[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			current.WriteByte(c)
			continue
		}

		if inQuotes {
			if c == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(c)
		} else {
			if c == '"' || c == '\'' {
				inQuotes = true
				quoteChar = c
				current.WriteByte(c)
			} else if c == '|' {
				// Found a pipe - split here
				segment := strings.TrimSpace(current.String())
				if segment == "" {
					return nil, errors.New("empty command before pipe")
				}
				segments = append(segments, segment)
				current.Reset()
			} else {
				current.WriteByte(c)
			}
		}
	}

	if inQuotes {
		return nil, errors.New("unclosed quote in pipeline")
	}

	// Add the last segment
	lastSegment := strings.TrimSpace(current.String())
	if lastSegment == "" {
		return nil, errors.New("empty command after pipe")
	}
	segments = append(segments, lastSegment)

	return segments, nil
}

// parseCaseStatement parses a case statement from a multi-line input
func parseCaseStatement(line string) (*CommandChain, error) {
	// This is a simplified parser for case statements
	// In a real implementation, this would need to handle multi-line parsing
	// For now, we'll return an error indicating case statements need multi-line support
	return nil, errors.New("case statements require multi-line parsing - use ParseCaseFromLines instead")
}

// parseIfStatement parses an if statement from a multi-line input
func parseIfStatement(line string) (*CommandChain, error) {
	// This is a simplified parser for if statements
	// In a real implementation, this would need to handle multi-line parsing
	// For now, we'll return an error indicating if statements need multi-line support
	return nil, errors.New("if statements require multi-line parsing - use ParseIfFromLines instead")
}

// ParseCaseFromLines parses a case statement from multiple lines
func ParseCaseFromLines(lines []string) (*CommandChain, error) {
	if len(lines) == 0 {
		return nil, errors.New("empty case statement")
	}

	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "case ") {
		return nil, errors.New("not a case statement")
	}

	// Extract the variable from "case $var in"
	parts := strings.Fields(firstLine)
	if len(parts) < 3 || parts[2] != "in" {
		return nil, errors.New("invalid case syntax: expected 'case $var in'")
	}

	variable := parts[1]
	caseStmt := &CaseStatement{
		Variable: variable,
		Patterns: []*CasePattern{},
	}

	// Parse patterns and commands
	i := 1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "esac" {
			break
		}

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}

		// Parse pattern line (e.g., "pattern1|pattern2)")
		if strings.HasSuffix(line, ")") {
			patternStr := strings.TrimSuffix(line, ")")
			patterns := strings.Split(patternStr, "|")
			for j := range patterns {
				patterns[j] = strings.TrimSpace(patterns[j])
				// Remove quotes from patterns
				patterns[j] = removeQuotes(patterns[j])
			}

			casePattern := &CasePattern{
				Patterns: patterns,
				Commands: []*Command{},
			}

			// Parse commands until we hit ";;" or next pattern
			i++
			for i < len(lines) {
				cmdLine := strings.TrimSpace(lines[i])
				if cmdLine == ";;" {
					i++
					break
				}
				if cmdLine == "esac" {
					break
				}
				if strings.HasSuffix(cmdLine, ")") {
					// Next pattern, don't increment i
					break
				}

				// Parse command
				if cmdLine != "" && !strings.HasPrefix(cmdLine, "#") {
					cmd, err := parseCommand(cmdLine)
					if err != nil {
						return nil, fmt.Errorf("error parsing command in case: %v", err)
					}
					casePattern.Commands = append(casePattern.Commands, cmd)
				}
				i++
			}

			caseStmt.Patterns = append(caseStmt.Patterns, casePattern)
		} else {
			i++
		}
	}

	control := &ControlStructure{
		Type: "case",
		Case: caseStmt,
	}

	return &CommandChain{
		Pipelines: [][]*Command{},
		Controls: []*ControlStructure{control},
	}, nil
}

// ParseIfFromLines parses an if statement from multiple lines
func ParseIfFromLines(lines []string) (*CommandChain, error) {
	if len(lines) == 0 {
		return nil, errors.New("empty if statement")
	}

	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "if ") {
		return nil, errors.New("not an if statement")
	}

	// Extract the condition from "if condition; then"
	conditionStr := strings.TrimPrefix(firstLine, "if ")
	conditionStr = strings.TrimSuffix(conditionStr, "; then")
	conditionStr = strings.TrimSuffix(conditionStr, ";then")
	conditionStr = strings.TrimSpace(conditionStr)

	// Parse the condition as a command
	conditionCmd, err := parseCommand(conditionStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing if condition: %v", err)
	}

	ifStmt := &IfStatement{
		Condition: []*Command{conditionCmd},
		ThenCommands: []*Command{},
		ElseCommands: []*Command{},
		ElifBranches: []*ElifBranch{},
	}

	// Parse the body
	i := 1
	currentSection := "then"
	var currentElifBranch *ElifBranch

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		
		if line == "fi" {
			break
		}

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}

		if line == "then" {
			currentSection = "then"
			i++
			continue
		}

		if line == "else" {
			currentSection = "else"
			i++
			continue
		}

		if strings.HasPrefix(line, "elif ") {
			// Save previous elif branch if exists
			if currentElifBranch != nil {
				ifStmt.ElifBranches = append(ifStmt.ElifBranches, currentElifBranch)
			}

			// Parse elif condition
			elifConditionStr := strings.TrimPrefix(line, "elif ")
			elifConditionStr = strings.TrimSuffix(elifConditionStr, "; then")
			elifConditionStr = strings.TrimSuffix(elifConditionStr, ";then")
			elifConditionStr = strings.TrimSpace(elifConditionStr)

			elifConditionCmd, err := parseCommand(elifConditionStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing elif condition: %v", err)
			}

			currentElifBranch = &ElifBranch{
				Condition: []*Command{elifConditionCmd},
				Commands: []*Command{},
			}
			currentSection = "elif"
			i++
			continue
		}

		// Parse command
		if line != "" && !strings.HasPrefix(line, "#") {
			cmd, err := parseCommand(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing command in if: %v", err)
			}

			switch currentSection {
			case "then":
				ifStmt.ThenCommands = append(ifStmt.ThenCommands, cmd)
			case "else":
				ifStmt.ElseCommands = append(ifStmt.ElseCommands, cmd)
			case "elif":
				if currentElifBranch != nil {
					currentElifBranch.Commands = append(currentElifBranch.Commands, cmd)
				}
			}
		}
		i++
	}

	// Save final elif branch if exists
	if currentElifBranch != nil {
		ifStmt.ElifBranches = append(ifStmt.ElifBranches, currentElifBranch)
	}

	control := &ControlStructure{
		Type: "if",
		If: ifStmt,
	}

	return &CommandChain{
		Pipelines: [][]*Command{},
		Controls: []*ControlStructure{control},
	}, nil
}

// stripInlineComment removes inline comments from a line, respecting quotes
func stripInlineComment(line string) string {
	inQuotes := false
	quoteChar := byte(0)
	escaped := false
	inCommandSubst := false
	parenDepth := 0

	for i := 0; i < len(line); i++ {
		c := line[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if inQuotes {
			if c == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
		} else if inCommandSubst {
			if c == '(' {
				parenDepth++
			} else if c == ')' {
				parenDepth--
				if parenDepth == 0 {
					inCommandSubst = false
				}
			}
		} else {
			if c == '"' || c == '\'' {
				inQuotes = true
				quoteChar = c
			} else if c == '$' && i+1 < len(line) && line[i+1] == '(' {
				// Start of command substitution
				inCommandSubst = true
				parenDepth = 1
				i++ // Skip the opening parenthesis
			} else if c == '#' {
				// Found unquoted comment, return line up to this point
				return strings.TrimRightFunc(line[:i], func(r rune) bool {
					return r == ' ' || r == '\t'
				})
			}
		}
	}

	return line
}

// tokenize splits a command line into tokens, handling quotes and command substitution
func tokenize(line string) ([]string, error) {
	// Strip inline comments first
	line = stripInlineComment(line)
	
	var tokens []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	escaped := false
	inCommandSubst := false
	parenDepth := 0

	for i := 0; i < len(line); i++ {
		c := line[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			current.WriteByte(c)
			continue
		}

		if inQuotes {
			if c == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(c)
		} else if inCommandSubst {
			current.WriteByte(c)
			if c == '(' {
				parenDepth++
			} else if c == ')' {
				parenDepth--
				if parenDepth == 0 {
					inCommandSubst = false
				}
			}
		} else {
			if c == '"' || c == '\'' {
				inQuotes = true
				quoteChar = c
				current.WriteByte(c)
			} else if c == '$' && i+1 < len(line) && line[i+1] == '(' {
				// Start of command substitution
				inCommandSubst = true
				parenDepth = 1
				current.WriteByte(c)
				// Also add the opening parenthesis
				i++
				current.WriteByte('(')
			} else if unicode.IsSpace(rune(c)) {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(c)
			}
		}
	}

	if inQuotes {
		return nil, errors.New("unclosed quote")
	}

	if inCommandSubst {
		return nil, errors.New("unclosed command substitution")
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}

// removeQuotes removes surrounding quotes from a string if present
func removeQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}