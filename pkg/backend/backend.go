// Package backend holds the interface for reading the IR of a statechart and generate the code
// needed to implement in different languages.
package backend

import (
	"io"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// GochartBackend is the abstract interface for all backends, regardless of the type of language
// they are meant to generate for. This is to decouple generated languages from the input (frontend)
// language defined to specify them.
type GochartBackend interface {
	Generate(sc *ir.Statechart) (header, body io.Reader, err error)
}
