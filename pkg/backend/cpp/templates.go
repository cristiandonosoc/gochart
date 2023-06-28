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

// commonContext is context that is common to all templates (header, body, etc.)
type commonContext struct {
	// Version is the version of gochart used to generate the files.
	Version string
	Time    time.Time
}

// templateManager is a helper struct to handle the common context for template loading.
type templateManager struct {
	headerTemplate *template.Template
	bodyTemplate   *template.Template

	sc     *ir.Statechart
	common *commonContext
}

func newTemplateManager(sc *ir.Statechart, common *commonContext) (*templateManager, error) {
	headerTemplate, err := readTemplate(headerFilename)
	if err != nil {
		return nil, fmt.Errorf("reading header template: %w", err)
	}

	bodyTemplate, err := readTemplate(bodyFilename)
	if err != nil {
		return nil, fmt.Errorf("reading body template: %w", err)
	}

	return &templateManager{
		headerTemplate: headerTemplate,
		bodyTemplate:   bodyTemplate,
		sc:             sc,
		common:         common,
	}, nil
}

func readTemplate(ep embedPath) (*template.Template, error) {
	epstr := string(ep)

	tmpl, err := template.ParseFS(embeddedFS, epstr)
	if err != nil {
		return nil, fmt.Errorf("reading embedded template %q: %w", epstr, err)
	}
	return tmpl, nil
}

func (tm *templateManager) generateHeader() (io.Reader, error) {
	context := &struct {
		commonContext
		Statechart *ir.Statechart
	}{
		commonContext: *tm.common,
		Statechart:    tm.sc,
	}

	var buf bytes.Buffer
	if err := tm.headerTemplate.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &buf, nil
}

func (tm *templateManager) generateBody() (io.Reader, error) {
	context := &struct {
		commonContext
		Statechart *ir.Statechart
	}{
		commonContext: *tm.common,
		Statechart:    tm.sc,
	}

	var buf bytes.Buffer
	if err := tm.bodyTemplate.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &buf, nil
}
