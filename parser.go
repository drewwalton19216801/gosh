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

// ParseLine parses a command line into individual commands
func ParseLine(line string) ([]*Command, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	// Handle pipes and command chaining later
	// For now, just parse a single command
	cmd, err := parseCommand(line)
	if err != nil {
		return nil, err
	}

	return []*Command{cmd}, nil
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

// tokenize splits a command line into tokens, handling quotes
func tokenize(line string) ([]string, error) {
	var tokens []string
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
			continue
		}

		if inQuotes {
			if c == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteByte(c)
			}
		} else {
			if c == '"' || c == '\'' {
				inQuotes = true
				quoteChar = c
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

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}