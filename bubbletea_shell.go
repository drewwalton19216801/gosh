package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BubbleTeaShell represents the Bubble Tea-based shell interface
type BubbleTeaShell struct {
	shell     *Shell
	textInput textinput.Model
	err       error
	quitting  bool
}

// NewBubbleTeaShell creates a new Bubble Tea shell instance
func NewBubbleTeaShell(shell *Shell) *BubbleTeaShell {
	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80

	return &BubbleTeaShell{
		shell:     shell,
		textInput: ti,
	}
}

// Init initializes the Bubble Tea model
func (m *BubbleTeaShell) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m *BubbleTeaShell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			// Execute the command
			line := strings.TrimSpace(m.textInput.Value())
			if line == "" {
				return m, nil
			}

			// Handle exit command
			if line == "exit" || line == "quit" {
				m.quitting = true
				return m, tea.Quit
			}

			// Execute the command
			if err := m.shell.ExecuteLine(line); err != nil {
				m.err = err
			}

			// Clear the input
			m.textInput.SetValue("")
			return m, nil

		case tea.KeyTab:
			// Handle tab completion
			return m.handleTabCompletion()
		}

	case tea.WindowSizeMsg:
		// Update text input width based on terminal size
		m.textInput.Width = msg.Width - 10 // Leave some margin
	}

	// Update the text input
	m.textInput, cmd = m.textInput.Update(msg)

	// Update suggestions based on current input
	m.updateSuggestions()

	return m, cmd
}

// View renders the shell interface
func (m *BubbleTeaShell) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Get current working directory for prompt
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	// Create prompt style
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	// Build the view
	var view strings.Builder

	// Show current directory
	view.WriteString(promptStyle.Render(fmt.Sprintf("gosh:%s> ", filepath.Base(cwd))))
	view.WriteString(m.textInput.View())
	view.WriteString("\n")

	// Show error if any
	if m.err != nil {
		view.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		view.WriteString("\n")
		m.err = nil // Clear error after showing
	}

	// Show suggestions if any
	suggestions := m.textInput.MatchedSuggestions()
	if len(suggestions) > 0 && m.textInput.Value() != "" {
		view.WriteString("\nSuggestions:\n")
		for i, suggestion := range suggestions {
			if i >= 10 { // Limit to 10 suggestions
				break
			}
			if i == m.textInput.CurrentSuggestionIndex() {
				view.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(fmt.Sprintf("  → %s", suggestion)))
			} else {
				view.WriteString(fmt.Sprintf("    %s", suggestion))
			}
			view.WriteString("\n")
		}
	}

	view.WriteString("\nPress Tab for completion, Ctrl+C to exit")

	return view.String()
}

// updateSuggestions updates the text input suggestions based on current input
func (m *BubbleTeaShell) updateSuggestions() {
	input := m.textInput.Value()
	if input == "" {
		m.textInput.SetSuggestions([]string{})
		return
	}

	// Get completions using the existing shell completion logic
	completions := m.getCompletions(input)
	m.textInput.SetSuggestions(completions)
}

// handleTabCompletion handles tab completion with proper character replacement
func (m *BubbleTeaShell) handleTabCompletion() (tea.Model, tea.Cmd) {
	input := m.textInput.Value()
	if input == "" {
		return m, nil
	}

	completions := m.getCompletions(input)
	if len(completions) == 0 {
		return m, nil
	}

	if len(completions) == 1 {
		// Single completion - replace the current word with the completion
		completion := completions[0]
		
		// Find the current word being completed
		words := strings.Fields(input)
		if len(words) == 0 {
			m.textInput.SetValue(completion)
			return m, nil
		}

		// Replace the last word with the completion
		if strings.HasSuffix(input, " ") {
			// If input ends with space, append the completion
			m.textInput.SetValue(input + completion)
		} else {
			// Replace the last word
			lastSpaceIndex := strings.LastIndex(input, " ")
			if lastSpaceIndex == -1 {
				// No spaces, replace entire input
				m.textInput.SetValue(completion)
			} else {
				// Replace from last space to end
				newValue := input[:lastSpaceIndex+1] + completion
				m.textInput.SetValue(newValue)
			}
		}
	}
	// For multiple completions, the suggestions are already shown in the view

	return m, nil
}

// getCompletions gets completion suggestions for the given input
func (m *BubbleTeaShell) getCompletions(input string) []string {
	// Parse the input to find what we're completing
	words := strings.Fields(input)
	if len(words) == 0 {
		return []string{}
	}

	var currentWord string
	var isFirstWord bool

	if strings.HasSuffix(input, " ") {
		// Completing a new word
		currentWord = ""
		isFirstWord = len(words) == 1
	} else {
		// Completing the last word
		currentWord = words[len(words)-1]
		isFirstWord = len(words) == 1
	}

	var completions []string

	if isFirstWord {
		// Complete command names
		completions = append(completions, m.getCommandCompletions(currentWord)...)
	}

	// Always try file/directory completion
	fileCompletions := m.getFileCompletions(currentWord)
	completions = append(completions, fileCompletions...)

	// Remove duplicates and sort
	completions = removeDuplicateStrings(completions)
	sort.Strings(completions)

	return completions
}

// getCommandCompletions gets command name completions
func (m *BubbleTeaShell) getCommandCompletions(prefix string) []string {
	var completions []string

	// Built-in commands
	builtins := []string{"cd", "pwd", "echo", "exit", "help", "history", "alias", "unalias", "export", "unset", "source", "function", "case", "if", "for", "while"}
	for _, builtin := range builtins {
		if strings.HasPrefix(builtin, prefix) {
			completions = append(completions, builtin)
		}
	}

	// Aliases
	for alias := range m.shell.aliases {
		if strings.HasPrefix(alias, prefix) {
			completions = append(completions, alias)
		}
	}

	// Functions
	for funcName := range m.shell.functions {
		if strings.HasPrefix(funcName, prefix) {
			completions = append(completions, funcName)
		}
	}

	// Executables in PATH
	pathCompletions := getExecutablesInPath(prefix)
	completions = append(completions, pathCompletions...)

	return completions
}

// getFileCompletions gets file and directory completions
func (m *BubbleTeaShell) getFileCompletions(prefix string) []string {
	var completions []string

	// Determine the directory to search
	dir := "."
	pattern := prefix

	if strings.Contains(prefix, string(os.PathSeparator)) {
		dir = filepath.Dir(prefix)
		pattern = filepath.Base(prefix)
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return completions
	}

	// Check if filesystem is case-sensitive
	caseSensitive := isCaseSensitiveFilesystem()

	for _, entry := range entries {
		name := entry.Name()
		
		// Skip hidden files unless prefix starts with dot
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(pattern, ".") {
			continue
		}

		var matches bool
		if caseSensitive {
			matches = strings.HasPrefix(name, pattern)
		} else {
			matches = strings.HasPrefix(strings.ToLower(name), strings.ToLower(pattern))
		}

		if matches {
			completion := name
			if dir != "." {
				completion = filepath.Join(dir, name)
			}

			// Add trailing slash for directories
			if entry.IsDir() {
				completion += string(os.PathSeparator)
			}

			completions = append(completions, completion)
		}
	}

	return completions
}

// getExecutablesInPath finds executable files in PATH that match the prefix
func getExecutablesInPath(prefix string) []string {
	var completions []string

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return completions
	}

	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	paths := strings.Split(pathEnv, pathSeparator)
	seen := make(map[string]bool)

	for _, path := range paths {
		if path == "" {
			continue
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			
			// On Windows, remove .exe extension for completion
			if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(name), ".exe") {
				name = name[:len(name)-4]
			}

			if strings.HasPrefix(name, prefix) && !seen[name] {
				completions = append(completions, name)
				seen[name] = true
			}
		}
	}

	return completions
}

// removeDuplicateStrings removes duplicate strings from a slice
func removeDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// RunBubbleTeaShell starts the Bubble Tea-based shell interface
func (s *Shell) RunBubbleTeaShell() error {
	fmt.Println("gosh - A simple shell (Bubble Tea interface)")
	fmt.Println("Type commands and press Enter to execute.")
	fmt.Println("Use Tab for completion, Ctrl+C to exit.")
	fmt.Println()

	model := NewBubbleTeaShell(s)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err := p.Run()
	return err
}