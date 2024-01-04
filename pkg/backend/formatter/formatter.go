// Package formatter is a backend that generates a formated version of the given gochart.
package formatter

import (
	"context"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// Options -----------------------------------------------------------------------------------------

type Options struct {
	Debug bool
}

func defaultOptions() *Options {
	options := &Options{
		Debug: false,
	}

	return options
}

func WithDebug(debug bool) option.Option {
	return optWrapper(func(options *Options) {
		options.Debug = debug
	})
}

func optWrapper(fn func(options *Options)) option.Option {
	return func(opts interface{}) {
		casted := opts.(*Options)
		fn(casted)
	}
}

// Backend -----------------------------------------------------------------------------------------

type formatterGochartBackend struct {
	options *Options
}

func NewFormatterGochartBackend(opts ...option.Option) (*formatterGochartBackend, error) {
	// TODO(cdc): Don't make the options global to the backend as it is now.
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &formatterGochartBackend{
		options: options,
	}, nil
}

func (fgb *formatterGochartBackend) Generate(ctx context.Context, gc *ir.Gochart, opts ...option.Option) ([]*common.BackendOutput, error) {
	// TODO(cdc): Don't make the options global to the backend as it is now.
	for _, opt := range opts {
		opt(fgb.options)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Continue.
	}

	iw := NewIndentedWriter()
	gf := newGochartFormatter(iw, fgb.options.Debug)
	gf.formatGochart(gc)

	content := iw.String()
	output := &common.BackendOutput{
		Content: strings.NewReader(content),
	}

	return []*common.BackendOutput{output}, nil
}
