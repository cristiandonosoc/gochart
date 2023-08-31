// Package unreal is an utlity package to integrate gochart with Unreal projects.
package unreal

import (
	// "bytes"
	"context"
	"fmt"
	// "os"

	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	// "github.com/cristiandonosoc/gochart/pkg/ir"
)

const (
	GochartExtension = ".gochart"
)

type GochartError struct {
	Gochart string
	Msg     string
}

func gochartErrorf(gochart, format string, args ...interface{}) *GochartError {
	return &GochartError{
		Gochart: gochart,
		Msg:     fmt.Sprintf(format, args...),
	}
}

type unrealIntegrator struct {
	Gocharts []*UnrealFile
	Modules  []*UnrealModule

	gf frontend.GochartFrontend
	gb backend.GochartBackend

	workers int
}

func NewUnrealIntegrator(gf frontend.GochartFrontend, gb backend.GochartBackend) (*unrealIntegrator, error) {
	return &unrealIntegrator{
		gf: gf,
		gb: gb,
	}, nil
}

func (ui *unrealIntegrator) Index(ctx context.Context, sourceDir string) error {
	if err := ui.indexUnrealProject(ctx, sourceDir, GochartExtension); err != nil {
		return fmt.Errorf("indexing unreal project at %q: %w", sourceDir, err)
	}

	return nil
}

func (ui *unrealIntegrator) Generate(platform string) []error {
	if len(ui.Gocharts) == 0 {
		return []error{fmt.Errorf("No gocharts listed. Did you call Index?")}
	}

	var errors []error

	for _, gochart := range ui.Gocharts {
		if err := ui.generateGochart(gochart); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (ui *unrealIntegrator) generateGochart(gochart *UnrealFile) error {
	// data, err := os.ReadFile(gochart)
	// if err != nil {
	// 	return fmt.Errorf("reading %q: %w", gochart, err)
	// }

	// frontendData, err := ui.gf.Process(bytes.NewReader(data))
	// if err != nil {
	// 	return fmt.Errorf("processing statechart: %w", err)
	// }

	// irData, err := ir.ProcessStatechartData(frontendData)
	// if err != nil {
	// 	return fmt.Errorf("processing frontend: %w", err)
	// }

	// header, cpp, err := ui.gb.Generate(irData, opts...)
	// if err != nil {
	// 	return fmt.Errorf("generating files: %w", err)
	// }

	fmt.Println("GENERATE:", gochart)
	return nil
}
