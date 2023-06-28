package cpp

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

//go:embed header.template.h body.template.cpp
var embeddedFS embed.FS

type embedPath string

const (
	headerFilename embedPath = "header.template.h"
	bodyFilename   embedPath = "body.template.cpp"
)

func readTemplate(ep embedPath) (*template.Template, error) {
	epstr := string(ep)

	tmpl, err := template.ParseFS(embeddedFS, epstr)
	if err != nil {
		return nil, fmt.Errorf("reading embedded template %q: %w", epstr, err)
	}
	return tmpl, nil
}

type commonContext struct {
	Version    string
	Time       time.Time
	Statechart *ir.Statechart
}

type headerContext struct {
	commonContext
}

type bodyContext struct {
	commonContext
}

func generateHeader(sc *ir.Statechart) (io.Reader, error) {
	tmpl, err := readTemplate(headerFilename)
	if err != nil {
		return nil, fmt.Errorf("reading header template: %w", err)
	}

	// TODO(cdc): Context should come from a single place so that it's shared between header and body.
	context := &headerContext{
		commonContext{
			Version:    "DEVELOPMENT",
			Time:       time.Now(),
			Statechart: sc,
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing header template: %w", err)
	}

	return &buf, nil
}

func generateBody(sc *ir.Statechart) (io.Reader, error) {
	tmpl, err := readTemplate(bodyFilename)
	if err != nil {
		return nil, fmt.Errorf("reading body template: %w", err)
	}

	// TODO(cdc): Context should come from a single place so that it's shared between header and body.
	context := &bodyContext{
		commonContext{
			Version:    "DEVELOPMENT",
			Time:       time.Now(),
			Statechart: sc,
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing body template: %w", err)
	}

	return &buf, nil
}
