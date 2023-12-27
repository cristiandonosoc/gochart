// Package cpp is a Gochart backend meant to generate C++ code for statechart.
package cpp

import (
	"context"
	"fmt"
	"time"

	cpp_gen "github.com/cristiandonosoc/gochart/internal/generation/cpp"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// Options -----------------------------------------------------------------------------------------

type Options struct {
	Version    string
	Time       time.Time
	OutputDir  string
	FilePrefix string

	Verbose bool
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

func WithOutputDir(outputDir string) option.Option {
	return optWrapper(func(options *Options) {
		options.OutputDir = outputDir
	})
}

func WithFilePrefix(filePrefix string) option.Option {
	return optWrapper(func(options *Options) {
		options.FilePrefix = filePrefix
	})
}

func WithVerbose(verbose bool) option.Option {
	return optWrapper(func(options *Options) {
		options.Verbose = verbose
	})
}

func optWrapper(fn func(options *Options)) option.Option {
	return func(opts interface{}) {
		casted := opts.(*Options)
		fn(casted)
	}
}

// Backend -----------------------------------------------------------------------------------------

type cppGochartBackend struct {
	outputDir string
	options   *Options
}

func NewCppGochartBackend(opts ...option.Option) (*cppGochartBackend, error) {
	// Get options.
	options := &Options{
		Version: "DEVELOPMENT",
		Time:    time.Now(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return &cppGochartBackend{
		options: options,
	}, nil
}

func (cpp *cppGochartBackend) Generate(ctx context.Context, sc *ir.Statechart, opts ...option.Option) ([]*common.BackendOutput, error) {
	for _, opt := range opts {
		opt(cpp.options)
	}

	config := &cpp_gen.Config{
		Version:           cpp.options.Version,
		Time:              cpp.options.Time,
		OutputDir:         cpp.options.OutputDir,
		FilePrefix:        cpp.options.FilePrefix,
		GenerateOutputDir: false,
		Verbose:           cpp.options.Verbose,
	}

	gens, err := cpp_gen.Generate(sc, config)
	if err != nil {
		return nil, fmt.Errorf("Generating files: %w", err)
	}

	outputs := make([]*common.BackendOutput, 0, len(gens))
	for _, gen := range gens {
		outputs = append(outputs, common.NewBackendOutputFromCppOutput(gen))
	}

	return outputs, nil
}
