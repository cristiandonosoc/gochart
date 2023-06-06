package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/cristiandonosoc/gocharts/pkg/frontend"
)

func internalMain() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("Usage: gocharts <relpath>")
	}

	// Read the input file.
	filepath := os.Args[1]
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("reading %q: %w", filepath, err)

	}

	scanner := frontend.NewScanner()

	_, err = scanner.Scan(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("scanning statechart at %q: %w", filepath, err)
	}

	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
