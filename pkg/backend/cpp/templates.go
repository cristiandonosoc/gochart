package cpp

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"text/template"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

//go:embed header.template.h body.template.cpp
var embeddedFS embed.FS

type embedPath string

const (
	headerFilename embedPath = "header.template.h"
	bodyFilename   embedPath = "body.template.cpp"
)

// templateManager is a helper struct to handle the common context for template loading.
type templateManager struct {
	headerTemplate *template.Template
	bodyTemplate   *template.Template

	sc *ir.Statechart
}

func newTemplateManager(sc *ir.Statechart) (*templateManager, error) {
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

func (tm *templateManager) generateBody(options *BackendOptions) (io.Reader, error) {
	context := newTemplateContext(tm.sc, options)

	var buf bytes.Buffer
	if err := tm.bodyTemplate.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &buf, nil
}

func (tm *templateManager) generateHeader(options *BackendOptions) (io.Reader, error) {
	context := newTemplateContext(tm.sc, options)

	var buf bytes.Buffer
	if err := tm.headerTemplate.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &buf, nil
}

// templateContext is a common struct that has helpers and information needed by the templates.
type templateContext struct {
	BackendOptions
	Statechart *ir.Statechart

	// Common Use strings.
	ImplName      string
	InterfaceName string
}

func newTemplateContext(sc *ir.Statechart, options *BackendOptions) *templateContext {
	tc := &templateContext{
		BackendOptions: *options,
		Statechart:     sc,

		ImplName:      fmt.Sprintf("Statechart%sImpl", sc.Name),
		InterfaceName: fmt.Sprintf("Statechart%s", sc.Name),
	}

	return tc
}
