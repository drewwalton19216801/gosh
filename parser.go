package main

import (
	"errors"
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

// CommandChain represents a sequence of command pipelines
type CommandChain struct {
	Pipelines [][]*Command
}

// ParseLine parses a command line into command chains (semicolon-separated pipelines)
func ParseLine(line string) (*CommandChain, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
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

// tokenize splits a command line into tokens, handling quotes and command substitution
func tokenize(line string) ([]string, error) {
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
			} else {
				current.WriteByte(c)
			}
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