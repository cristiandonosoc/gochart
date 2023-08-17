// Package gochart_lang is the frontend for what we can "gochart_lang", which is a custom spec
// language to describe statecharts.
package gochart_lang

import (
	"bytes"
	"fmt"
	"io"

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
func NewGochartLangFrontend(opts ...option.Option) *gochartLangFrontend {
	// Get options.
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}


	return &gochartLangFrontend{
		options: options,
	}
}

func (glf *gochartLangFrontend) Process(r io.Reader) (*output.StatechartData, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input reader: %w", err)
	}

	scanner := NewScanner()
	tokens, errors := scanner.gatherTokens(bytes.NewReader(data))
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("Lexing error: %v\n", err)
		}
		return nil, fmt.Errorf("gochart: parsing input file")
	}

	parser := NewParser(glf.options)
	root, errors := parser.Parse(tokens)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("Parsing error: %v\n", err)
		}
		return nil, fmt.Errorf("gochart: parsing input file")
	}

	if len(root.statecharts) != 1 {
		return nil, fmt.Errorf("For now only one statechart is supported per file")
	}

	ft := newFrontendTranslator()
	scdata := ft.translateStatechart(root.statecharts[0])

	return scdata, nil
}
