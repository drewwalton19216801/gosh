package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var command string
	var useBubbleTea bool
	flag.StringVar(&command, "c", "", "execute command and exit")
	flag.BoolVar(&useBubbleTea, "bubbletea", false, "use Bubble Tea interface for better tab completion")
	flag.Parse()

	shell := NewShell()

	if command != "" {
		// Command mode (-c flag)
		if err := shell.ExecuteLine(command); err != nil {
			fmt.Fprintf(os.Stderr, "gosh: %v\n", err)
			os.Exit(1)
		}
	} else if flag.NArg() > 0 {
		// Script mode
		scriptFile := flag.Arg(0)
		scriptArgs := flag.Args()[1:] // Get all arguments after the script filename
		if err := shell.ExecuteScript(scriptFile, scriptArgs); err != nil {
			fmt.Fprintf(os.Stderr, "gosh: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Interactive mode
		if useBubbleTea {
			if err := shell.RunBubbleTeaShell(); err != nil {
				fmt.Fprintf(os.Stderr, "gosh: %v\n", err)
				os.Exit(1)
			}
		} else {
			shell.Run()
		}
	}
}
