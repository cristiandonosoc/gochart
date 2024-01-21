// Binary docs_generator is an utility to generate documentation markdown from the IR options.
package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

//go:embed options.md.template
var embeddedFS embed.FS

const (
	kTemplateFilePath = "options.md.template"
)

// OptionData describes the option information so that it can be generated.
type OptionData struct {
	Name        string
	FlagName    string
	Kind        string
	Description string
	Synopsis    string
}

type TemplateContext struct {
	GochartOptions []*OptionData
	StateOptions   []*OptionData
}

func collectGochartOptions() []*OptionData {
	irOpts := ir.GochartOptionDefinitions()

	options := make([]*OptionData, 0, len(irOpts))
	for _, irOpt := range irOpts {
		options = append(options, &OptionData{
			Name:        irOpt.Name,
			FlagName:    irOpt.FlagName,
			Kind:        irOpt.Kind.String(),
			Description: irOpt.Description,
			Synopsis:    strings.TrimSpace(irOpt.Synopsis),
		})
	}

	return options
}

func collectStateOptions() []*OptionData {
	return nil
}

func internalMain() error {
	fmt.Println(os.Args)
	if len(os.Args) != 2 {
		return fmt.Errorf("Usage: docs_generator [REPO_ROOT] [OUTPUT_MARKDOWN_FILEPATH]")
	}

	output := os.Args[1]

	tmpl, err := template.New(kTemplateFilePath).ParseFS(embeddedFS, kTemplateFilePath)
	if err != nil {
		return fmt.Errorf("reading template %q: %w", kTemplateFilePath, err)
	}

	context := &TemplateContext{}
	context.GochartOptions = collectGochartOptions()
	context.StateOptions = collectStateOptions()

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	if err := files.WriteToFile(output, &buf); err != nil {
		return fmt.Errorf("writing generated docs to %q: %w", output, err)
	}

	fmt.Printf("Wrote docs to %q\n", output)
	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
