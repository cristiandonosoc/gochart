// Package unreal is the backend generation that takes care of generating gochart files for an
// Unreal project.
package unreal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cristiandonosoc/gochart/internal/build"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gunreal/pkg/unreal"
)

const (
	kModuleSearcherWorkerCount = 10
	kGochartBuilderWorkerCount = 16
)

type GenerationStrategy int

const (
	GenerationStrategy_Invalid GenerationStrategy = iota

	// Inline means generating the files right next to the gochart file.
	GenerationStrategy_Inline

	// ModuleGeneratedDir creates a Generated directory in the module root and append the internal
	// route to it.
	GenerationStrategy_ModuleGeneratedDir

	// UHT means generating files into the project's intermediate files, in the UHT folder.
	GenerationStrategy_UHT
)

func (gs GenerationStrategy) String() string {
	switch gs {
	case GenerationStrategy_Inline:
		return "Inline"
	case GenerationStrategy_ModuleGeneratedDir:
		return "ModuleGeneratedDir"
	case GenerationStrategy_UHT:
		return "UHT"
	default:
		return "<unknown>"
	}
}

func (gs GenerationStrategy) Valid() bool {
	return gs != GenerationStrategy_Invalid
}

func ListGenerationStrategies() []GenerationStrategy {
	return []GenerationStrategy{
		GenerationStrategy_Inline,
		GenerationStrategy_ModuleGeneratedDir,
		GenerationStrategy_UHT,
	}
}

func NewGenerationStrategy(value string) GenerationStrategy {
	for _, gs := range ListGenerationStrategies() {
		if strings.ToLower(gs.String()) == strings.ToLower(value) {
			return gs
		}
	}

	return GenerationStrategy_Invalid
}

// Options -----------------------------------------------------------------------------------------

// TODO(cdc): Join options into a common backend options as well.
type Options struct {
	Version string
	Time    time.Time

	Platform unreal.Platform
	Strategy GenerationStrategy
}

func defaultOptions() *Options {
	options := &Options{
		Version:  build.BuildVersion(),
		Time:     time.Now(),
		Platform: unreal.Platform_Windows,
		Strategy: GenerationStrategy_Invalid,
	}

	return options
}

func optWrapper(fn func(options *Options)) option.Option {
	return func(opts interface{}) {
		casted := opts.(*Options)
		fn(casted)
	}
}

func WithVersion(version string) option.Option {
	return optWrapper(func(options *Options) {
		options.Version = version
	})
}

func WithTime(t time.Time) option.Option {
	return optWrapper(func(options *Options) {
		options.Time = t
	})
}

func WithPlatform(platform unreal.Platform) option.Option {
	return optWrapper(func(options *Options) {
		options.Platform = platform
	})
}

func WithStrategy(strategy GenerationStrategy) option.Option {
	return optWrapper(func(options *Options) {
		options.Strategy = strategy
	})
}

// Backend -----------------------------------------------------------------------------------------

type UnrealGochartBackend struct {
	Project         *unreal.Project
	gochartFrontend frontend.GochartFrontend
	options         *Options
}

func NewUnrealGochartBackend(ctx context.Context, projectDir string) (*UnrealGochartBackend, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("missing project dir")
	}

	project, err := unreal.IndexProject(ctx, projectDir)
	if err != nil {
		return nil, fmt.Errorf("indexing unreal project: %w", err)
	}

	gf, err := gochart_lang.NewGochartLangFrontend()
	if err != nil {
		return nil, fmt.Errorf("creating gochart lang frontend: %w", err)
	}

	return &UnrealGochartBackend{
		Project:         project,
		gochartFrontend: gf,
	}, nil
}

type GochartOutput struct {
	Gochart *unreal.File
	Outputs []*common.BackendOutput
}

func (o *GochartOutput) Orphan() bool {
	return !o.Gochart.Exists()
}

func (o *GochartOutput) UpToDate() bool {
	if !o.Gochart.Exists() {
		return false
	}

	modTime := o.Gochart.FileInfo.ModTime()
	for _, output := range o.Outputs {
		if !output.IsLoaded() {
			return false
		}

		if output.DiskContent.Stat.ModTime().Before(modTime) {
			return false
		}
	}

	return true
}

func (ugb *UnrealGochartBackend) Generate(ctx context.Context, opts ...option.Option) ([]*GochartOutput, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	ugb.options = options

	if err := ugb.validate(); err != nil {
		return nil, fmt.Errorf("validating: %w", err)
	}

	gocharts, err := ugb.Project.SearchForFilesByExtension(ctx, ugb.options.Platform, []string{".gochart"})
	if err != nil {
		return nil, fmt.Errorf("searching for gochart files: %w", err)
	}

	if len(gocharts) == 0 {
		fmt.Println("No gocharts found.")
		return nil, nil
	}

	fmt.Printf("Found %v gocharts:\n", len(gocharts))
	for _, gochart := range gocharts {
		fmt.Printf("- %s\n", gochart)
	}

	goutputs, err := ugb.generateGocharts(ctx, ugb.Project, gocharts)
	if err != nil {
		return nil, fmt.Errorf("generating gocharts: %w", err)
	}

	// Mark them as generated by unreal.
	for _, goutput := range goutputs {
		for _, output := range goutput.Outputs {
			output.Metadata.IsUnrealGochart = true
		}
	}

	return goutputs, nil
}

func (ugb *UnrealGochartBackend) validate() error {
	if !ugb.options.Strategy.Valid() {
		return fmt.Errorf("invalid generation strategy. Is the --unreal_strategy set?")
	}

	return nil
}
