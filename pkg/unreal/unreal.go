// Package unreal is an utlity package to integrate gochart with Unreal projects.
package unreal

/*

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	//"github.com/cristiandonosoc/gochart/internal/files"
	//"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend"
	//cpp_backend "github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	//"github.com/cristiandonosoc/gochart/pkg/ir"
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

//go:generate stringer -type=GenerationStrategy
type GenerationStrategy int

const (
	// Intermediate means generating files into the project's intermediate files.
	GenerationStrategy_Intermediate GenerationStrategy = iota

	// Inline means generating the files right next to the gochart file.
	GenerationStrategy_Inline

	// ModuleGeneratedDir creates a Generated directory in the module root and append the internal
	// route to it.
	GenerationStrategy_ModuleGeneratedDir
)

func (ui *unrealIntegrator) Generate(platformId string, strategy GenerationStrategy) []error {
	if len(ui.Gocharts) == 0 {
		return []error{fmt.Errorf("No gocharts listed. Did you call Index?")}
	}

	platform, err := NewUnrealPlatform(platformId)
	if err != nil {
		return []error{fmt.Errorf("could not find platform for %q: %w", platformId, err)}
	}

	fmt.Printf("Running generation for project %s, Platform: %s, Generation Strategy: %s\n", ui.ProjectDir, platform, strategy)

	var errors []error
	for _, gochart := range ui.Gocharts {
		if err := ui.generateGochart(gochart, platform, strategy); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (ui *unrealIntegrator) generateGochart(gochart *UnrealFile, platform UnrealPlatform, strategy GenerationStrategy) error {
	gfc, err := ui.NewGochartFileContext(gochart, platform, strategy)
	if err != nil {
		return fmt.Errorf("generating gochart context for %q: %w", gochart.Name(), err)
	}

	// See if we need to generated.
	if !gfc.ShouldGenerate() {
		fmt.Printf("Gochart %s up to date\n", gochart.Path)
		return nil
	}

	panic("REIMPLEMENT ME")

	// lf, err := files.LoadFileFromPath(gfc.GochartFile.Path)
	// if err != nil {
	// 	return fmt.Errorf("loading file: %w", err)
	// }

	// frontendData, err := ui.gf.Process(lf)
	// if err != nil {
	// 	return fmt.Errorf("processing statechart: %w", err)
	// }

	// irData, err := ir.ProcessStatechartData(frontendData)
	// if err != nil {
	// 	return fmt.Errorf("processing frontend: %w", err)
	// }

	// var opts []option.Option
	// opts = append(opts, cpp_backend.WithHeaderIncludePath(gfc.Header.Name()))

	// header, cpp, err := ui.gb.Generate(irData, opts...)
	// if err != nil {
	// 	return fmt.Errorf("generating files: %w", err)
	// }

	// if err := gfc.writeIfDifferent(gfc.Header, header); err != nil {
	// 	return fmt.Errorf("writing header if different: %w", err)
	// }

	// if err := gfc.writeIfDifferent(gfc.Cpp, cpp); err != nil {
	// 	return fmt.Errorf("writing cpp if different: %w", err)
	// }

	// return nil
}

type GochartFileContext struct {
	GochartFile *UnrealFile
	Platform    UnrealPlatform

	Header *UnrealFile
	Cpp    *UnrealFile
}

func (ui *unrealIntegrator) NewGochartFileContext(gochart *UnrealFile, platform UnrealPlatform, strategy GenerationStrategy) (*GochartFileContext, error) {
	headerPath, cppPath, err := ui.gochartOutputPaths(gochart, platform, strategy)
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

	if gfc.Header != nil {
		sb.WriteString(fmt.Sprintf("Header: %s\n", gfc.Header))
	} else {
		sb.WriteString("Header: <NONE>\n")
	}

	if gfc.Cpp != nil {
		sb.WriteString(fmt.Sprintf("Cpp: %s\n", gfc.Cpp))
	} else {
		sb.WriteString("Cpp: <NONE>\n")
	}

	return sb.String()
}

func (gfc *GochartFileContext) ShouldGenerate() bool {
	if gfc.isNewer(gfc.Header) || gfc.isNewer(gfc.Cpp) {
		return true
	}

	return false
}

func (gfc *GochartFileContext) isNewer(uf *UnrealFile) bool {
	// If there is no file info, it means the file does not exist.
	if uf.FileInfo == nil {
		return true
	}

	return gfc.GochartFile.FileInfo.ModTime().After(uf.FileInfo.ModTime())
}

func (gfc *GochartFileContext) writeIfDifferent(uf *UnrealFile, r io.Reader) error {
	// Read the contents of the generated.
	generatedData, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading generated data: %w", err)
	}

	// If the file exists, read the file and compare it to the generated data.
	if uf.FileInfo != nil {
		fileData, err := os.ReadFile(uf.Path)
		if err != nil {
			return fmt.Errorf("reading %q: %w", uf.Path, err)
		}

		// If they're equal, we just ignore.
		if bytes.Equal(fileData, generatedData) {
			return nil
		}
	}

	// If we're here, it means we need to write to the file.

	// Ensure the path exists.
	if err := os.MkdirAll(filepath.Dir(uf.Path), os.ModeDir); err != nil {
		return fmt.Errorf("creating base directories %q: %w", filepath.Dir(uf.Path), err)
	}

	// First we create or truncate the file.
	file, err := os.Create(uf.Path)
	if err != nil {
		return fmt.Errorf("creating/truncating %q: %w", uf.Path, err)
	}
	defer file.Close()

	if _, err := file.Write(generatedData); err != nil {
		return fmt.Errorf("writing to %q: %w", uf.Path, err)
	}
	fmt.Printf("Wrote %q\n", uf.Path)

	return nil
}

// gochartOutputPaths gets the output paths belonging to this particular gochart.
// Intermediate/Build/<PLATFORM>/UnrealEditor/Inc/Gochart/<MODULE>/<INTERNAL_PATH>/<NAME>.gochart.(h|cpp)
func (ui *unrealIntegrator) gochartOutputPaths(gochart *UnrealFile, platform UnrealPlatform, strategy GenerationStrategy) (_header string, _cpp string, _err error) {
	if ui.ProjectDir == "" {
		return "", "", fmt.Errorf("project dir is empty. Did you call Index?")
	}

	var base string
	switch strategy {
	case GenerationStrategy_Intermediate:
		// First it's built by the intermediate platform location.
		base = filepath.Join(ui.ProjectDir, "Intermediate", "Build", string(platform), "UnrealEditor", "Inc", "Gochart")

		// Then we get to the per module place and internal path.
		base = filepath.Join(base, gochart.Module.Name, filepath.Dir(gochart.InternalPath()))
	case GenerationStrategy_Inline:
		base = filepath.Dir(gochart.Path)
	case GenerationStrategy_ModuleGeneratedDir:
		base = filepath.Join(gochart.Module.BaseDir, "Generated")
		base = filepath.Join(base, filepath.Dir(gochart.InternalPath()))
	default:
		return "", "", fmt.Errorf("unknown generation strategy: %s", strategy)
	}

	// Now we return the output paths for the C++ backend.
	gochartName := gochart.Name()
	ext := filepath.Ext(gochartName)
	name := gochartName[0 : len(gochartName)-len(ext)]

	header := filepath.Join(base, fmt.Sprintf("%s.gochart.generated.h", name))
	cpp := filepath.Join(base, fmt.Sprintf("%s.gochart.generated.cpp", name))

	return header, cpp, nil
}
*/
