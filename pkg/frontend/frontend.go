// Package frontend holds the interface for reading different kind of languages/specs that we might
// support for reading statecharts.
package frontend

import (
	"bytes"
	"io"
	"os"
	"fmt"
)

// GochartFrontend is the abstract interface for all frontends, regardless of the type of data they
// consume. This is used to decouple language specifications from the rest of the program.
type GochartFrontend interface {
	// Process takes the content describing input for this frontend and is meant to return frontend
	// data that can then be consumed by others parts of the program, mostly the |ir| package.
	Process(r io.Reader) (*StatechartData, error)

	// ProcessFromFile is meant to read the contents of a file pointed by |path| and return the
	// statechart definition from it.
	ProcessFromFile(path string) (*StatechartData, error)
}

// ProcessFromFile is a convenience function that reads a whole file and pass it to the particular
// GochartFrontend implementation. Mostly meant for each interface implementation to use it for they
// explosing their own |ProcessFromFile|.
func ProcessFromFile(gf GochartFrontend, path string) (*StatechartData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}

	return gf.Process(bytes.NewReader(data))
}
