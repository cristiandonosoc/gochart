// cpp is a Gochart backend meant to generate C++ code for statechart.
package cpp

import (
	"fmt"
	"io"
	"time"

	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

var _ backend.GochartBackend = (*cppGochartBackend)(nil)

type cppGochartBackend struct {
	options *BackendOptions
}

type BackendOptions struct {
	HeaderInclude string
	Time          time.Time
	Version       string
}

type Option func(*BackendOptions)

func NewCppGochartBackend(opts ...Option) *cppGochartBackend {
	options := &BackendOptions{
		Version: "DEVELOPMENT",
		Time:    time.Now(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return &cppGochartBackend{
		options: options,
	}
}

func (cpp *cppGochartBackend) Generate(sc *ir.Statechart) (_header, _body io.Reader, _err error) {
	tm, err := newTemplateManager(sc)
	if err != nil {
		return nil, nil, fmt.Errorf("building new template manager: %w", err)
	}

	header, err := tm.generateHeader(cpp.options)
	if err != nil {
		return nil, nil, fmt.Errorf("generating header: %w", err)
	}

	body, err := tm.generateBody(cpp.options)
	if err != nil {
		return nil, nil, fmt.Errorf("generating body: %w", err)
	}

	return header, body, nil
}
