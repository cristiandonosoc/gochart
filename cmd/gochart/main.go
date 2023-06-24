package main

import (
	"fmt"
	"os"
)

func internalMain() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("Usage: gochart <relpath>")
	}

	// Read the input file.
	filepath := os.Args[1]
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("reading %q: %w", filepath, err)

	}

	fmt.Println(string(data))

	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
