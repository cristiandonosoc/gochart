// Package backend holds the interface for reading the IR of a statechart and generate the code
// needed to implement in different languages.
package backend

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// GochartBackend is the abstract interface for all backends, regardless of the type of language
// they are meant to generate for. This is to decouple generated languages from the input (frontend)
// language defined to specify them.
type GochartBackend interface {
	Generate(sc *ir.Statechart) (header, body io.Reader, err error)
}

type backendFactory interface {
	parse() (GochartBackend, error)
}

// SetFlags adds all the flags available to configure the given backend.
func SetFlags(backendKind string, fs *flag.FlagSet) (backendFactory, error) {

	switch strings.ToLower(backendKind) {
	case "cpp":
		return newCppBackendFactory(fs)
	default:
		return nil, fmt.Errorf("unknown backend %q", backendKind)
	}
}

// Parse gets the result of the fs.Parse and creates a backend from them.
func Parse(bf backendFactory) (GochartBackend, error) {
	return bf.parse()
}

// CPP Backend -------------------------------------------------------------------------------------

type cppBackendFactory struct {
	headerIncludePath string
}

var _ backendFactory = (*cppBackendFactory)(nil)

func newCppBackendFactory(fs *flag.FlagSet) (backendFactory, error) {
	cbf := &cppBackendFactory{}

	fs.StringVar(&cbf.headerIncludePath, "header-include-path", "", "The include path to put in the generated cpp file")

	return cbf, nil
}

func (cbf *cppBackendFactory) parse() (GochartBackend, error) {
	var options []option.Option

	if cbf.headerIncludePath != "" {
		options = append(options, cpp.WithHeaderIncludePath(cbf.headerIncludePath))
	}

	return cpp.NewCppGochartBackend(options...)
}
