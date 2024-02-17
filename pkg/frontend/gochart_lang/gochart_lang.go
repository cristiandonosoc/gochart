// Package gochart_lang is the frontend for what we can "gochart_lang", which is a custom spec
// language to describe gocharts.
package gochart_lang

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/errors"
	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

// Options -----------------------------------------------------------------------------------------

type Options struct {
	Verbose bool
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

// Frontend ----------------------------------------------------------------------------------------

type gochartLangFrontend struct {
	options *Options
}

// NewFrontend returns a frontend capable of parting the gochart language.
func NewGochartLangFrontend(opts ...option.Option) (*gochartLangFrontend, error) {
	// Get options.
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	return &gochartLangFrontend{
		options: options,
	}, nil
}

func (glf *gochartLangFrontend) Process(lf *files.LoadedFile) (*common.GochartData, error) {
	scanner := NewScanner()
	tokens, scannerErrors := scanner.gatherTokens(lf)
	if len(scannerErrors) > 0 {
		return nil, errors.NewErrorContainer(scannerErrors)
	}

	parser := NewParser(lf, glf.options)
	root, parserErrors := parser.Parse(tokens)
	if len(parserErrors) > 0 {
		return nil, errors.NewErrorContainer(parserErrors)
	}

	if len(root.gocharts) != 1 {
		return nil, fmt.Errorf("For now only one gochart is supported per file")
	}

	ft := newFrontendTranslator(lf)
	gcdata := ft.translateGochart(root.gocharts[0])

	return gcdata, nil
}
