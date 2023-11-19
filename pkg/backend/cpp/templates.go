package cpp

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"text/template"

	"github.com/cristiandonosoc/gochart/pkg/ir"

	"github.com/Masterminds/sprig/v3"
)

//go:embed statechart.h.template statechart.cpp.template
var embeddedFS embed.FS

type embedPath string

const (
	headerFilename embedPath = "statechart.h.template"
	cppFilename    embedPath = "statechart.cpp.template"
)

// templateManager is a helper struct to handle the common context for template loading.
type templateManager struct {
	headerTemplate *template.Template
	cppTemplate    *template.Template

	sc *ir.Statechart
}

func newTemplateManager(sc *ir.Statechart) (*templateManager, error) {
	headerTemplate, err := readTemplate(headerFilename)
	if err != nil {
		return nil, fmt.Errorf("reading header template: %w", err)
	}

	cppTemplate, err := readTemplate(cppFilename)
	if err != nil {
		return nil, fmt.Errorf("reading cpp template: %w", err)
	}

	return &templateManager{
		headerTemplate: headerTemplate,
		cppTemplate:    cppTemplate,
		sc:             sc,
	}, nil
}

func collectTemplateHelperFunctions() template.FuncMap {
	return map[string]any{
		"TriggerCallArgs": TemplateHelper_TriggerCallArgs,
	}
}

func readTemplate(ep embedPath) (*template.Template, error) {
	// Load sprig functions.
	epstr := string(ep)
	tmpl, err := template.New(epstr).
		Funcs(sprig.FuncMap()).
		Funcs(collectTemplateHelperFunctions()).
		ParseFS(embeddedFS, epstr)
	if err != nil {
		return nil, fmt.Errorf("reading embedded template %q: %w", epstr, err)
	}
	return tmpl, nil
}

func (tm *templateManager) generateBody(options *Options) (io.Reader, error) {
	context := newTemplateContext(tm.sc, options)

	var buf bytes.Buffer
	if err := tm.cppTemplate.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &buf, nil
}

func (tm *templateManager) generateHeader(options *Options) (io.Reader, error) {
	context := newTemplateContext(tm.sc, options)

	var buf bytes.Buffer
	if err := tm.headerTemplate.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return &buf, nil
}

// templateContext is a common struct that has helpers and information needed by the templates.
type templateContext struct {
	Options

	Statechart *ir.Statechart

	// Common Use strings.
	ImplName      string
	InterfaceName string
}

func newTemplateContext(sc *ir.Statechart, options *Options) *templateContext {
	tc := &templateContext{
		Options:    *options,
		Statechart: sc,

		ImplName:      fmt.Sprintf("Statechart%sImpl", sc.Name),
		InterfaceName: fmt.Sprintf("Statechart%s", sc.Name),
	}

	return tc
}

// TEMPLATE HELPER FUNCTIONS -----------------------------------------------------------------------

func TemplateHelper_TriggerCallArgs(trigger *ir.Trigger, prefix string) []string {
	argsStrings := make([]string, 0, len(trigger.Args))

	for _, arg := range trigger.Args {
		argsStrings = append(argsStrings, fmt.Sprintf("%s%s", prefix, arg.Name))
	}

	return argsStrings
}
