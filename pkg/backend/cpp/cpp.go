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
}

func NewCppGochartBackend() *cppGochartBackend {
	return &cppGochartBackend{}
}

func (cpp *cppGochartBackend) Generate(sc *ir.Statechart) (_header, _body io.Reader, _err error) {
	common := &commonContext{
		Version: "DEVELOPMENT",
		Time:    time.Now(),
	}

	tm, err := newTemplateManager(sc, common)
	if err != nil {
		return nil, nil, fmt.Errorf("building new template manager: %w", err)
	}

	header, err := tm.generateHeader()
	if err != nil {
		return nil, nil, fmt.Errorf("generating header: %w", err)
	}

	body, err := tm.generateBody()
	if err != nil {
		return nil, nil, fmt.Errorf("generating body: %w", err)
	}

	return header, body, nil
}
