// Package frontend holds the interface for reading different kind of languages/specs that we might
// support for reading statecharts.
package frontend

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"
)

// GochartFrontend is the abstract interface for all frontends, regardless of the type of data they
// consume. This is used to decouple language specifications from the rest of the program.
type GochartFrontend interface {
	// Process takes the content describing input for this frontend and is meant to return frontend
	// data that can then be consumed by others parts of the program, mostly the |ir| package.
	Process(r io.Reader) (*output.StatechartData, error)
}

type frontendFactory func(*frontendFlags) (GochartFrontend, error)

var factories = map[string]frontendFactory{
	"gochart": newGochartFrontend,
	"yaml":    newYamlFrontend,
}

type frontendFlags struct {
	kind string
}

// SetFlags adds all the flags available to configure the frontend.
func SetFlags(fs *flag.FlagSet) *frontendFlags {
	ff := &frontendFlags{}

	keys := make([]string, 0, len(factories))
	for key := range factories {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fs.StringVar(&ff.kind, "frontend", "gochart",
		fmt.Sprintf("The frontend language to read statechart from. Valid values: %v", keys))

	return ff
}

// Parse gets the result of the fs.Parse and creates a frontend from them.
func Parse(ff *frontendFlags) (GochartFrontend, error) {
	factory, ok := factories[strings.ToLower(ff.kind)]
	if !ok {
		return nil, fmt.Errorf("unknown frontend %q", ff.kind)
	}

	return factory(ff)
}

func newGochartFrontend(ff *frontendFlags) (GochartFrontend, error) {
	// Parse the options.
	return gochart_lang.NewGochartLangFrontend()
}

func newYamlFrontend(ff *frontendFlags) (GochartFrontend, error) {
	return yaml.NewYamlFrontend()
}
