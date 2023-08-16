// Package frontend holds the interface for reading different kind of languages/specs that we might
// support for reading statecharts.
package frontend

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

// FrontendKind is an enum holding all the possible frontends available.
type FrontendKind int

const (
	// FrontendKind_Gochart represents the main language with a custom parser.
	FrontendKind_Gochart FrontendKind = iota

	// FrontendKind_Yaml is a simple yaml parser, mostly used for development purposes.
	FrontendKind_Yaml
)

func (fk *FrontendKind) Set(value string) error {
	switch strings.ToLower(value) {
	case "gochart":
		*fk = FrontendKind_Gochart
	case "yaml":
		*fk = FrontendKind_Yaml
	default:
		return fmt.Errorf("invalid frontend value: want gochart or yaml, got %v", value)
	}

	return nil
}

type FrontendOption func(opt interface{})

// GochartFrontend is the abstract interface for all frontends, regardless of the type of data they
// consume. This is used to decouple language specifications from the rest of the program.
type GochartFrontend interface {
	// Process takes the content describing input for this frontend and is meant to return frontend
	// data that can then be consumed by others parts of the program, mostly the |ir| package.
	Process(r io.Reader) (*output.StatechartData, error)
}

// ProcessFromFile is a convenience function that reads a whole file and pass it to the particular
// GochartFrontend implementation. Mostly meant for each interface implementation to use it for they
// explosing their own |ProcessFromFile|.
func ProcessFromFile(gf GochartFrontend, path string) (*output.StatechartData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}

	return gf.Process(bytes.NewReader(data))
}
