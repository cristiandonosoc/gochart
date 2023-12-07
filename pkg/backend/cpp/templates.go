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

//go:embed forward_declare.h.template declare.h.template define.h.template
var embeddedFS embed.FS

type embedPath string

const (
	kForwardDeclareFilename embedPath = "forward_declare.h.template"
	kDeclareFilename        embedPath = "declare.h.template"
	kDefineFilename         embedPath = "define.h.template"
)

// templateManager is a helper struct to handle the common context for template loading.
type templateManager struct {
	forwardDeclareTemplate *template.Template
	declareTemplate        *template.Template
	defineTemplate         *template.Template

	sc *ir.Statechart
}

func newTemplateManager(sc *ir.Statechart) (*templateManager, error) {
	forwardDeclareTemplate, err := readTemplate(kForwardDeclareFilename)
	if err != nil {
		return nil, fmt.Errorf("reading forward declare template %q: %w", kForwardDeclareFilename, err)
	}

	declareTemplate, err := readTemplate(kDeclareFilename)
	if err != nil {
		return nil, fmt.Errorf("reading declare template %q: %w", kDeclareFilename, err)
	}

	defineTemplate, err := readTemplate(kDefineFilename)
	if err != nil {
		return nil, fmt.Errorf("reading define template %q: %w", kDefineFilename, err)
	}

	return &templateManager{
		forwardDeclareTemplate: forwardDeclareTemplate,
		declareTemplate:        declareTemplate,
		defineTemplate:         defineTemplate,
		sc:                     sc,
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

func (tm *templateManager) generateTemplates(options *Options) (_forwardDeclare, _declare, _define io.Reader, _err error) {
	context := newTemplateContext(tm.sc, options)

	var forwardDeclareBuf bytes.Buffer
	if err := tm.forwardDeclareTemplate.Execute(&forwardDeclareBuf, context); err != nil {
		return nil, nil, nil, fmt.Errorf("executing forward declare template: %w", err)
	}

	var declareBuf bytes.Buffer
	if err := tm.declareTemplate.Execute(&declareBuf, context); err != nil {
		return nil, nil, nil, fmt.Errorf("executing declare template: %w", err)
	}

	var defineBuf bytes.Buffer
	if err := tm.defineTemplate.Execute(&defineBuf, context); err != nil {
		return nil, nil, nil, fmt.Errorf("executing define template: %w", err)
	}

	return &forwardDeclareBuf, &declareBuf, &defineBuf, nil
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
