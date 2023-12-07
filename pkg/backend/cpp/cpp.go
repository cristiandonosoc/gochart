// Package cpp is a Gochart backend meant to generate C++ code for statechart.
package cpp

import (
	"fmt"
	"time"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// Options -----------------------------------------------------------------------------------------

type Options struct {
	HeaderIncludePath string
	Time              time.Time
	Version           string

	Verbose bool
}

func WithHeaderIncludePath(path string) option.Option {
	return optWrapper(func(options *Options) {
		options.HeaderIncludePath = path
	})
}

func WithTime(t time.Time) option.Option {
	return optWrapper(func(options *Options) {
		options.Time = t
	})
}

func WithVersion(version string) option.Option {
	return optWrapper(func(options *Options) {
		options.Version = version
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
	options *Options
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

func (cpp *cppGochartBackend) Generate(sc *ir.Statechart, opts ...option.Option) ([]*common.BackendOutput, error) {
	for _, opt := range opts {
		opt(cpp.options)
	}

	tm, err := newTemplateManager(sc)
	if err != nil {
		return nil, fmt.Errorf("building new template manager: %w", err)
	}

	forwardDeclare, declare, define, err := tm.generateTemplates(cpp.options)
	if err != nil {
		return nil, fmt.Errorf("generating templates: %w", err)
	}

	outputs := []*common.BackendOutput{
		{
			Name:    "forward_declare",
			Content: forwardDeclare,
		},
		{
			Name:    "declare",
			Content: declare,
		},
		{
			Name:    "define",
			Content: define,
		},
	}
	return outputs, nil
}
