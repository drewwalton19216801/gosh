package main

import (
	"fmt"
	"os"
)

func main() {
	shell := NewShell()
	if len(os.Args) > 1 {
		// Script mode
		if err := shell.ExecuteScript(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "gosh: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Interactive mode
		shell.Run()
	}
}