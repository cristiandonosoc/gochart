// Package gochart is the main entrypoint to the compiler as a library.
// For using it, see cmd/gochart
package gochart

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/cristiandonosoc/gochart/internal/files"
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
	Tag     string
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

func (gc *GochartCompiler) GenerateToFiles(path, headerPath, cppPath string) error {
	outputs, err := gc.Generate(path)
	if err != nil {
		return fmt.Errorf("generating C++ output files: %w", err)
	}

	outputCount := 0
	for _, output := range outputs {
		if output.Tag == "header" {
			if err := files.WriteToFile(headerPath, output.Content); err != nil {
				return fmt.Errorf("outputing header to %q: %w", headerPath, err)
			}

			outputCount++
			fmt.Printf("Output header to %q\n", path)
		}

		if output.Tag == "cpp" {
			if err := files.WriteToFile(cppPath, output.Content); err != nil {
				return fmt.Errorf("outputing cpp to %q: %w", cppPath, err)
			}

			outputCount++
			fmt.Printf("Output cpp to %q\n", path)
		}
	}

	if outputCount != 2 {
		return fmt.Errorf("expected to output 2 files, outputted %v", outputCount)
	}

	return nil

}
