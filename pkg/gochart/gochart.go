// Package gochart is the main entrypoint to the compiler as a library.
// For using it, see cmd/gochart
package gochart

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

type GochartCompiler struct {
	Frontend frontend.GochartFrontend
	Backend  backend.GochartBackend
}

func NewGochartCompiler(gf frontend.GochartFrontend, gb backend.GochartBackend) (*GochartCompiler, error) {
	return &GochartCompiler{
		Frontend: gf,
		Backend:  gb,
	}, nil
}

func (gc *GochartCompiler) PrintFrontend(path string, w io.Writer) error {
	lf, err := output.LoadFileFromPath(path)
	if err != nil {
		return fmt.Errorf("loading file: %w", err)
	}

	frontendData, err := gc.Frontend.Process(lf)
	if err != nil {
		return fmt.Errorf("processing frontend: %w", err)
	}

	printer := frontend.NewPrinter(w)
	printer.PrintStatechart(frontendData)

	return nil
}

func (gc *GochartCompiler) PrintIR(path string, w io.Writer) error {
	lf, err := output.LoadFileFromPath(path)
	if err != nil {
		return fmt.Errorf("loading file: %w", err)
	}

	frontendData, err := gc.Frontend.Process(lf)
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
	Tag     string
	Content io.Reader
}

func (gc *GochartCompiler) Generate(path string, opts ...option.Option) ([]*BackendOutput, error) {
	lf, err := output.LoadFileFromPath(path)
	if err != nil {
		return nil, fmt.Errorf("loading file: %w", err)
	}

	frontendData, err := gc.Frontend.Process(lf)
	if err != nil {
		return nil, fmt.Errorf("processing statechart: %w", err)
	}

	irData, err := ir.ProcessStatechartData(frontendData)
	if err != nil {
		return nil, fmt.Errorf("processing frontend: %w", err)
	}

	header, cpp, err := gc.Backend.Generate(irData, opts...)
	if err != nil {
		return nil, fmt.Errorf("generating files: %w", err)
	}

	outputs := []*BackendOutput{
		{
			Tag:     "header",
			Content: header,
		},
		{
			Tag:     "cpp",
			Content: cpp,
		},
	}

	return outputs, nil
}
