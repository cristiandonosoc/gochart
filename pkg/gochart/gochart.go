// Package gochart is the main entrypoint to the compiler as a library.
// For using it, see cmd/gochart
package gochart

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/ir"
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

	frontendData, err := gc.gf.Process(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("processing frontend: %w", err)
	}

	printer := frontend.NewPrinter(w)
	printer.PrintStatechart(frontendData)

	return nil
}

func (gc *gochartCompiler) PrintIR(path string, w io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", path, err)
	}

	frontendData, err := gc.gf.Process(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("processing statechart: %w", err)
	}

	irData, err := ir.ProcessStatechartData(frontendData)
	if err != nil {
		return fmt.Errorf("processing frontend: %w", err)
	}

	printer := ir.NewPrinter(w)
	printer.PrintStatechart(irData)

	return nil
}

type BackendOutput struct {
	Name   string
	Output io.Reader
}

func (gc *gochartCompiler) PrintBackend(w io.Writer) ([]*BackendOutput, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")
}
