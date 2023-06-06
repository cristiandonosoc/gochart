// Package frontend holds the logic to partse a statechart from a file/text into our intermediate
// representation, which can then be used by a backend.
package frontend

import (
	"bufio"
	"fmt"
	"io"

	"github.com/cristiandonosoc/gocharts/pkg/ir"
)

// Scanner is an object capable of taking in a reader and scanning over to create a valid
// representation of our statechart grammar.
type Scanner struct {
	start   int
	current int
	line    int
}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Scan(r io.Reader) (*ir.Statechart, error) {
	var lines []string

	lineScanner := bufio.NewScanner(r)
	lineScanner.Split(bufio.ScanLines)
	for lineScanner.Scan() {
		lines = append(lines, lineScanner.Text())
	}

	for i, line := range lines {
		fmt.Printf("%d: %s\n", i, line)
	}

	return &ir.Statechart{}, nil
}
