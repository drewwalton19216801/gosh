package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

// Shell represents the main shell instance
type Shell struct {
	env       map[string]string
	aliases   map[string]string
	history   []string
	exitCode  int
	running   bool
	rl        *readline.Instance
}

// NewShell creates a new shell instance
func NewShell() *Shell {
	s := &Shell{
		env:     make(map[string]string),
		aliases: make(map[string]string),
		history: make([]string, 0),
		running: true,
	}
	
	// Initialize readline with history support
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "gosh> ",
		HistoryFile:     "/tmp/.gosh_history",
		AutoComplete:    nil,
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

	defer s.rl.Close()

	for s.running {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "unknown"
		} else {
			cwd = filepath.Base(cwd)
		}
		
		// Update prompt with current directory
		s.rl.SetPrompt(fmt.Sprintf("gosh:%s> ", cwd))
		
		line, err := s.rl.Readline()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
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

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if err := s.ExecuteLine(line); err != nil {
			return fmt.Errorf("line %d: %v", lineNum, err)
		}
	}

	return scanner.Err()
}

// Exit sets the shell to stop running
func (s *Shell) Exit(code int) {
	s.exitCode = code
	s.running = false
}