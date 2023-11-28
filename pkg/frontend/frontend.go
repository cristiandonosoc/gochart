// Package frontend holds the interface for reading different kind of languages/specs that we might
// support for reading statecharts.
package frontend

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"

	"github.com/urfave/cli/v2"
)

// GochartFrontend is the abstract interface for all frontends, regardless of the type of data they
// consume. This is used to decouple language specifications from the rest of the program.
type GochartFrontend interface {
	// Process takes the content describing input for this frontend and is meant to return frontend
	// data that can then be consumed by others parts of the program, mostly the |ir| package.
	Process(lf *files.LoadedFile) (*output.StatechartData, error)
}

type frontendFactory func(*cli.Context) (GochartFrontend, error)

var factories = map[string]frontendFactory{
	"gochart": newGochartFrontend,
	"yaml":    newYamlFrontend,
}

type frontendFlags struct {
	kind string
}

// SetFlags adds all the flags available to configure the frontend.
func SetFlags() []cli.Flag {
	keys := make([]string, 0, len(factories))
	for key := range factories {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return []cli.Flag{
		&cli.StringFlag{
			Name:  "frontend",
			Value: "gochart",
			Usage: fmt.Sprintf("The frontend `LANGUAGE` to read the statechart from. Valid values: %v", keys),
		},
	}
}

// Parse gets the result of the fs.Parse and creates a frontend from them.
func Parse(cliCtx *cli.Context) (GochartFrontend, error) {
	keys := make([]string, 0, len(factories))
	for key := range factories {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	kind := strings.ToLower(cliCtx.String("frontend"))
	factory, ok := factories[kind]
	if !ok {
		return nil, fmt.Errorf("unknown frontend %q. Valid values: %v", kind, keys)
	}

	return factory(cliCtx)
}

func newGochartFrontend(cliCtx *cli.Context) (GochartFrontend, error) {
	options := []option.Option{
		gochart_lang.WithVerbose(cliCtx.Bool("verbose")),
	}

	return gochart_lang.NewGochartLangFrontend(options...)
}

func newYamlFrontend(cliCtx *cli.Context) (GochartFrontend, error) {
	return yaml.NewYamlFrontend()
}
