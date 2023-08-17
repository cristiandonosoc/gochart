// Package cpp is a Gochart backend meant to generate C++ code for statechart.
package cpp

import (
	"fmt"
	"io"
	"time"

	"github.com/cristiandonosoc/gochart/internal/option"
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

func NewCppGochartBackend(opts ...option.Option) *cppGochartBackend {
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
	}
}

func (cpp *cppGochartBackend) Generate(sc *ir.Statechart) (_header, _cpp io.Reader, _err error) {
	tm, err := newTemplateManager(sc)
	if err != nil {
		return nil, nil, fmt.Errorf("building new template manager: %w", err)
	}

	headerResult, err := tm.generateHeader(cpp.options)
	if err != nil {
		return nil, nil, fmt.Errorf("generating header: %w", err)
	}

	cppResult, err := tm.generateBody(cpp.options)
	if err != nil {
		return nil, nil, fmt.Errorf("generating cpp: %w", err)
	}

	return headerResult, cppResult, nil
}
