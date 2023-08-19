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

type GochartCompiler struct {
	gf frontend.GochartFrontend
	gb backend.GochartBackend
}

func NewGochartCompiler(gf frontend.GochartFrontend, gb backend.GochartBackend) (*GochartCompiler, error) {
	return &GochartCompiler{
		gf: gf,
		gb: gb,
	}, nil
}

func (gc *GochartCompiler) PrintFrontend(path string, w io.Writer) error {
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

func (gc *GochartCompiler) PrintIR(path string, w io.Writer) error {
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
	Name    string
	Content io.Reader
}

func (gc *GochartCompiler) Generate(path string) ([]*BackendOutput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", path, err)
	}

	frontendData, err := gc.gf.Process(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("processing statechart: %w", err)
	}

	irData, err := ir.ProcessStatechartData(frontendData)
	if err != nil {
		return nil, fmt.Errorf("processing frontend: %w", err)
	}

	header, cpp, err := gc.gb.Generate(irData)
	if err != nil {
		return nil, fmt.Errorf("generating files: %w", err)
	}

	outputs := []*BackendOutput{
		{
			Name:    "Header",
			Content: header,
		},
		{
			Name:    "Cpp",
			Content: cpp,
		},
	}

	return outputs, nil
}
