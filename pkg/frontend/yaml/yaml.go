// Package yaml is a simple frontend that reads yaml. Mostly used to quickly test the whole pipeline
// instead of requiring a custom language/parser.
package yaml

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"gopkg.in/yaml.v2"
)

func NewYamlFrontend() *yamlFrontend {
	return &yamlFrontend{}
}

var _ frontend.GochartFrontend = (*yamlFrontend)(nil)

type yamlFrontend struct {
}

func (yf *yamlFrontend) Process(r io.Reader) (*frontend.StatechartData, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input reader: %w", err)
	}

	var scdata frontend.StatechartData
	if err := yaml.Unmarshal(data, &scdata); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}

	return &scdata, nil
}

func (yf *yamlFrontend) ProcessFromFile(path string) (*frontend.StatechartData, error) {
	return frontend.ProcessFromFile(yf, path)
}
