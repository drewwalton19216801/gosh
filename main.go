package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var command string
	flag.StringVar(&command, "c", "", "execute command and exit")
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
		shell.Run()
	}
}
