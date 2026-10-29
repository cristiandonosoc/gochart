// Package backend holds the interface for reading the IR of a gochart and generate the code
// needed to implement in different languages.
package backend

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/backend/rust"
	"github.com/cristiandonosoc/gochart/pkg/ir"

	"github.com/urfave/cli/v2"
)

// GochartBackend is the abstract interface for all backends, regardless of the type of language
// they are meant to generate for. This is to decouple generated languages from the input (frontend)
// language defined to specify them.
type GochartBackend interface {
	Generate(ctx context.Context, gc *ir.Gochart, opts ...option.Option) ([]*common.BackendOutput, error)
}

type backendFactory func(*cli.Context) (GochartBackend, error)

var factories = map[string]backendFactory{
	"cpp":  newCppBackend,
	"rust": newRustBackend,
}

// SetFlags adds all the flags available to configure the given backend.
func SetFlags(backendKind string) ([]cli.Flag, error) {
	keys := make([]string, 0, len(factories))
	for key := range factories {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	switch strings.ToLower(backendKind) {
	case "cpp":
		return setCppFlags(), nil
	case "rust":
		return []cli.Flag{}, nil
	default:
		return nil, fmt.Errorf("unknown backend %q. Valid values: %v", backendKind, keys)
	}
}

func Parse(backendKind string, cliCtx *cli.Context) (GochartBackend, error) {
	keys := make([]string, 0, len(factories))
	for key := range factories {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	kind := strings.ToLower(strings.ToLower(backendKind))
	factory, ok := factories[kind]
	if !ok {
		return nil, fmt.Errorf("unknown frontend %q. Valid values: %v", kind, keys)
	}

	return factory(cliCtx)
}

// CPP Backend -------------------------------------------------------------------------------------

func setCppFlags() []cli.Flag {
	return []cli.Flag{}
}

func newCppBackend(cliCtx *cli.Context) (GochartBackend, error) {
	var options []option.Option
	return cpp.NewCppGochartBackend(options...)
}

// Rust Backend ------------------------------------------------------------------------------------

func newRustBackend(cliCtx *cli.Context) (GochartBackend, error) {
	return rust.NewRustGochartBackend()
}
