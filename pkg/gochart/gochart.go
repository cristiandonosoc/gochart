// Package gochart is the main entrypoint to the compiler as a library.
// For using it, see cmd/gochart
package gochart

import (
	"bytes"
	"fmt"
	"os"
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

func (gc *gochartCompiler) PrintFrontend(path string, w io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", path, err)
	}

	scdata, err := gc.gf.Process(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("processing statechart: %w", err)
	}

	printer := frontend.NewPrinter(w)
	printer.PrintStatechart(scdata)

	return nil
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
