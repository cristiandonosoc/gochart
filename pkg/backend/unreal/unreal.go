// Package unreal is the backend generation that takes care of generating gochart files for an
// Unreal project.
package unreal

import (
	"context"
	"fmt"
	"time"

	"github.com/cristiandonosoc/gochart/internal/build"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/internal/unreal"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
)

const (
	kModuleSearcherWorkerCount = 10
	kGochartBuilderWorkerCount = 16
)

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

func (gs GenerationStrategy) String() string {
	switch gs {
	case GenerationStrategy_Intermediate:
		return "Intermediate"
	case GenerationStrategy_Inline:
		return "Inline"
	case GenerationStrategy_ModuleGeneratedDir:
		return "ModuleGeneratedDir"
	default:
		return "<unknown>"
	}
}

// Options -----------------------------------------------------------------------------------------

// TODO(cdc): Join options into a common backend options as well.
type Options struct {
	Version string
	Time    time.Time

	Platform unreal.UnrealPlatform
	Strategy GenerationStrategy
}

func defaultOptions() *Options {
	options := &Options{
		Version:  build.BuildVersion(),
		Time:     time.Now(),
		Platform: unreal.UnrealPlatform_Windows,
		Strategy: GenerationStrategy_Intermediate,
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

func WithPlatform(platform unreal.UnrealPlatform) option.Option {
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

type unrealGochartBackend struct {
	Project         *unreal.UnrealProject
	gochartFrontend frontend.GochartFrontend
	options         *Options
}

func NewUnrealGochartBackend(ctx context.Context, projectDir string) (*unrealGochartBackend, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("missing project dir")
	}

	project, err := unreal.IndexUnrealProject(ctx, projectDir)
	if err != nil {
		return nil, fmt.Errorf("indexing unreal project: %w", err)
	}

	gf, err := gochart_lang.NewGochartLangFrontend()
	if err != nil {
		return nil, fmt.Errorf("creating gochart lang frontend: %w", err)
	}

	return &unrealGochartBackend{
		Project:         project,
		gochartFrontend: gf,
	}, nil
}

type GochartOutput struct {
	Gochart *unreal.UnrealFile
	Outputs []*common.BackendOutput
}

func (o *GochartOutput) UpToDate() bool {
	for _, output := range o.Outputs {
		if !output.UpToDate {
			return false
		}
	}

	return true
}

// List goes over and encounters all the gocharts and the generated files, without generating them.
func (ugb *unrealGochartBackend) List(ctx context.Context, opts ...option.Option) ([]*GochartOutput, error) {
	// TODO(cdc): Don't make the options global to the backend as it is now.
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	ugb.options = options

	if err := ugb.validate(); err != nil {
		return nil, fmt.Errorf("validating: %w", err)
	}

	gochartPaths, err := ugb.Project.SearchForFilesByExtension(ctx, ".gochart")
	if err != nil {
		return nil, fmt.Errorf("searching for gochart files: %w", err)
	}

	if len(gochartPaths) == 0 {
		fmt.Println("No gocharts found.")
		return nil, nil
	}

	goutputs := make([]*GochartOutput, 0, len(gochartPaths))
	for _, gochartPath := range gochartPaths {
		goutput, _, err := ugb.calculateGochartExpectedOutputs(gochartPath)
		if err != nil {
			return nil, fmt.Errorf("calculating gochart expected outputs: %w", err)
		}

		goutputs = append(goutputs, goutput)
	}

	return goutputs, nil
}

func (ugb *unrealGochartBackend) Generate(ctx context.Context, opts ...option.Option) ([]*GochartOutput, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	ugb.options = options

	if err := ugb.validate(); err != nil {
		return nil, fmt.Errorf("validating: %w", err)
	}

	gocharts, err := ugb.Project.SearchForFilesByExtension(ctx, ".gochart")
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

	outputs, err := ugb.generateGocharts(ctx, ugb.Project, gocharts)
	if err != nil {
		return nil, fmt.Errorf("generating gocharts: %w", err)
	}

	return outputs, nil
}

func (ugb *unrealGochartBackend) validate() error {
	return nil
}
