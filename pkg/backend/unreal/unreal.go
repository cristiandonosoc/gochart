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

type unrealGochartBackend struct {
	Project         *unreal.Project
	gochartFrontend frontend.GochartFrontend
	options         *Options
}

func NewUnrealGochartBackend(ctx context.Context, projectDir string) (*unrealGochartBackend, error) {
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

	return &unrealGochartBackend{
		Project:         project,
		gochartFrontend: gf,
	}, nil
}

type GochartOutput struct {
	Gochart *unreal.File
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
	if !ugb.options.Strategy.Valid() {
		return fmt.Errorf("invalid generation strategy. Is the --unreal_strategy set?")
	}

	return nil
}
