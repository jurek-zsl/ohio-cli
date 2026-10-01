package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Locate ohio executable
	ohioPath, err := exec.LookPath("ohio")
	if err != nil {
		if self, selfErr := os.Executable(); selfErr == nil {
			candidate := filepath.Join(filepath.Dir(self), "ohio")
			if _, statErr := os.Stat(candidate); statErr == nil {
				ohioPath = candidate
				err = nil
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ohfs is an alias for ohio. Please install or ensure 'ohio' is in your PATH.\n")
		os.Exit(1)
	}

	cmd := exec.Command(ohioPath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
