// Package gochart is the main entrypoint to the compiler as a library.
// For using it, see cmd/gochart
package gochart

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

type gochartCompiler struct {
	gf frontend.GochartFrontend
	gb backend.GochartBackend
}

func NewGochartCompiler(gf frontend.GochartFrontend, gb backend.GochartBackend) (*gochartCompiler, error) {
	return &gochartCompiler{
		gf: gf,
		gb: gb,
	}, nil
}

func (gc *gochartCompiler) PrintFrontend() (io.Reader, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")

}

func (gc *gochartCompiler) PrintIR(w io.Writer) (io.Reader, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")
}

type BackendOutput struct {
	Name   string
	Output io.Reader
}

func (gc *gochartCompiler) PrintBackend(w io.Writer) ([]*BackendOutput, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")
}
