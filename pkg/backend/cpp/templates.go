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

//go:embed header.template.h
var embeddedFS embed.FS

type embedPath string

const (
	headerFilename embedPath = "header.template.h"
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

func generateHeader(sc *ir.Statechart) (io.Reader, error) {
	tmpl, err := readTemplate(headerFilename)
	if err != nil {
		return nil, fmt.Errorf("reading header template: %w", err)
	}

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
