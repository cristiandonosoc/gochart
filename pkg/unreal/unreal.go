// Package unreal is an utlity package to integrate gochart with Unreal projects.
package unreal

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
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
	ProjectDir string

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

func (ui *unrealIntegrator) Index(ctx context.Context, projectDir string) error {
	projectDir = filepath.Clean(projectDir)

	if err := ui.indexUnrealProject(ctx, projectDir, GochartExtension); err != nil {
		return fmt.Errorf("indexing unreal project at %q: %w", projectDir, err)
	}

	ui.ProjectDir = projectDir

	return nil
}

func (ui *unrealIntegrator) Generate(platformId string) []error {
	if len(ui.Gocharts) == 0 {
		return []error{fmt.Errorf("No gocharts listed. Did you call Index?")}
	}

	platform, err := NewUnrealPlatform(platformId)
	if err != nil {
		return []error{fmt.Errorf("could not find platform for %q: %w", platformId, err)}
	}

	var errors []error
	for _, gochart := range ui.Gocharts {
		if err := ui.generateGochart(gochart, platform); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (ui *unrealIntegrator) generateGochart(gochart *UnrealFile, platform UnrealPlatform) error {
	gochartCtx, err := ui.NewGochartFileContext(gochart, platform)
	if err != nil {
		return fmt.Errorf("generating gochart context for %q: %w", gochart.Name(), err)
	}

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

	fmt.Println("GENERATE:", gochartCtx)
	return nil
}

type GochartFileContext struct {
	GochartFile *UnrealFile
	Platform    UnrealPlatform

	Header *UnrealFile
	Cpp    *UnrealFile
}

func (ui *unrealIntegrator) NewGochartFileContext(gochart *UnrealFile, platform UnrealPlatform) (*GochartFileContext, error) {
	headerPath, cppPath, err := ui.gochartOutputPaths(gochart, platform)
	if err != nil {
		return nil, fmt.Errorf("getting output paths: %w", err)
	}

	header, err := ui.NewUnrealFile(headerPath)
	if err != nil {
		return nil, fmt.Errorf("creating unreal file for header %q: %w", headerPath, err)
	}

	cpp, err := ui.NewUnrealFile(cppPath)
	if err != nil {
		return nil, fmt.Errorf("creating unreal file for cpp %q: %w", cppPath, err)
	}

	return &GochartFileContext{
		GochartFile: gochart,
		Platform:    platform,
		Header:      header,
		Cpp:         cpp,
	}, nil
}

func (gfc *GochartFileContext) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Gochart: %s\n", gfc.GochartFile))
	sb.WriteString(fmt.Sprintf("Platform: %s\n", gfc.Platform))
	sb.WriteString(fmt.Sprintf("Header: %s\n", gfc.Header))
	sb.WriteString(fmt.Sprintf("Cpp: %s\n", gfc.Cpp))

	return sb.String()
}

// gochartOutputPaths gets the output paths belonging to this particular gochart.
// Intermediate/Build/<PLATFORM>/UnrealEditor/Inc/Gochart/<MODULE>/<INTERNAL_PATH>/<NAME>.gochart.(h|cpp)
func (ui *unrealIntegrator) gochartOutputPaths(gochart *UnrealFile, platform UnrealPlatform) (_header string, _cpp string, _err error) {
	if ui.ProjectDir == "" {
		return "", "", fmt.Errorf("project dir is empty. Did you call Index?")
	}

	// First it's built by the intermediate platform location.
	base := filepath.Join(ui.ProjectDir, "Intermediate", "Build", string(platform), "UnrealEditor", "Inc", "Gochart")

	// Then we get to the per module place and internal path.
	base = filepath.Join(base, gochart.Module.Name, filepath.Dir(gochart.InternalPath()))

	// Now we return the output paths for the C++ backend.
	gochartName := gochart.Name()
	ext := filepath.Ext(gochartName)
	name := gochartName[0 : len(gochartName)-len(ext)]

	header := filepath.Join(base, fmt.Sprintf("%s.gochart.h", name))
	cpp := filepath.Join(base, fmt.Sprintf("%s.gochart.cpp", name))

	return header, cpp, nil
}
