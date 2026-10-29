// Package rust is a Gochart backend meant to generate Rust code for gochart.
package rust

import (
	"context"
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/build"
	rust_gen "github.com/cristiandonosoc/gochart/internal/generation/rust"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// Options -----------------------------------------------------------------------------------------

type Options struct {
	Version    string
	OutputDir  string
	FilePrefix string
}

func WithVersion(version string) option.Option {
	return optWrapper(func(options *Options) {
		options.Version = version
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

func optWrapper(fn func(options *Options)) option.Option {
	return func(opts interface{}) {
		casted := opts.(*Options)
		fn(casted)
	}
}

// Backend -----------------------------------------------------------------------------------------

type rustGochartBackend struct {
	options *Options
}

func NewRustGochartBackend(opts ...option.Option) (*rustGochartBackend, error) {
	options := &Options{
		Version: build.BuildVersion(),
	}
	for _, opt := range opts {
		opt(options)
	}

	return &rustGochartBackend{
		options: options,
	}, nil
}

func (rb *rustGochartBackend) Generate(ctx context.Context, gc *ir.Gochart, opts ...option.Option) ([]*common.BackendOutput, error) {
	for _, opt := range opts {
		opt(rb.options)
	}

	config := &rust_gen.Config{
		Metadata: &common.Metadata{
			Version: rb.options.Version,
		},
		OutputDir:  rb.options.OutputDir,
		FilePrefix: rb.options.FilePrefix,
	}

	outputs, err := rust_gen.Generate(gc, config)
	if err != nil {
		return nil, fmt.Errorf("generating files: %w", err)
	}

	return outputs, nil
}
