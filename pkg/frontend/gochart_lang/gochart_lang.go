// Package gochart_lang is the frontend for what we can "gochart_lang", which is a custom spec
// language to describe statecharts.
package gochart_lang

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
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

func (glf *gochartLangFrontend) Process(lf *output.LoadedFile) (*output.StatechartData, error) {
	scanner := NewScanner()
	tokens, errors := scanner.gatherTokens(lf)
	if len(errors) > 0 {
		return nil, NewFrontendErrorContainer(lf, errors)
	}

	parser := NewParser(glf.options)
	root, errors := parser.Parse(tokens)
	if len(errors) > 0 {
		return nil, NewFrontendErrorContainer(lf, errors)
	}

	if len(root.statecharts) != 1 {
		return nil, fmt.Errorf("For now only one statechart is supported per file")
	}

	ft := newFrontendTranslator()
	scdata := ft.translateStatechart(root.statecharts[0])

	return scdata, nil
}
