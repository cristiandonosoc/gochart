// Package formatter is a backend that generates a formated version of the given statechart.
package formatter

import (
	"context"
	"fmt"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
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

func (fgb *formatterGochartBackend) Generate(ctx context.Context, sc *ir.Statechart, opts ...option.Option) ([]*common.BackendOutput, error) {
	// TODO(cdc): Don't make the options global to the backend as it is now.
	for _, opt := range opts {
		opt(fgb.options)
	}

	var fe frontend_common.FrontendEntry = sc.FrontendData
	iw := NewIndentedWriter()

	f := newFormatterEntry(fe)

	if err := fgb.formatEntry(ctx, f, iw, fgb.options.Debug); err != nil {
		return nil, fmt.Errorf("formatting entries: %w", err)
	}

	content := iw.String()
	output := &common.BackendOutput{
		Content: strings.NewReader(content),
	}

	return []*common.BackendOutput{output}, nil
}

func (fgb *formatterGochartBackend) formatEntry(ctx context.Context, f formatterEntry, iw *IndentedWriter, debug bool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue.
	}

	if f.Indents() {
		iw.MustIncreaseIndent()
	}

	if !debug {
		f.Write(iw)
	} else {
		f.DebugWrite(iw)
	}

	for _, cfe := range f.ChildEntries() {
		if err := fgb.formatEntry(ctx, cfe, iw, debug); err != nil {
			return err
		}
	}

	if f.Indents() {
		iw.MustDecreateIndent()
	}

	return nil
}
