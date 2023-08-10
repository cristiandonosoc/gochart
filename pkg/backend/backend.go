// Package backend holds the interface for reading the IR of a statechart and generate the code
// needed to implement in different languages.
package backend

import (
	"fmt"
	"io"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// BackendKind is an enum holding all the possible backend available.
type BackendKind int

const (
	// BackendKind_Cpp will generate code for C++.
	// Generates a header (.h) and a body (.cpp) files.
	BackendKind_Cpp BackendKind = iota
)

func (bk *BackendKind) Set(value string) error {
	switch strings.ToLower(value) {
	case "cpp":
		*bk = BackendKind_Cpp
	default:
		return fmt.Errorf("invalid backend value: want cpp, got %v", value)
	}

	return nil
}

// GochartBackend is the abstract interface for all backends, regardless of the type of language
// they are meant to generate for. This is to decouple generated languages from the input (frontend)
// language defined to specify them.
type GochartBackend interface {
	Generate(sc *ir.Statechart) (header, body io.Reader, err error)
}
