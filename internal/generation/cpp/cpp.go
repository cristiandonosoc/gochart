// Package cpp takes care of generating cpp code depending on the specifications.
package cpp

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

const (
	kOutputName_ForwardDeclare = "forward_declare"
	kOutputName_Declare        = "declare"
	kOutputName_Define         = "define"
)

type Config struct {
	Version           string
	Time              time.Time
	OutputDir         string
	FilePrefix        string
	GenerateOutputDir bool
	Verbose           bool
}

func (c *Config) validate() error {
	if c.Version == "" {
		return fmt.Errorf("missing version")
	}

	if c.Time.IsZero() {
		return fmt.Errorf("non set time")
	}

	if c.OutputDir == "" {
		return fmt.Errorf("missing output dir")
	}

	if !c.GenerateOutputDir {
		if ok, err := files.DirExists(c.OutputDir); err != nil {
			return fmt.Errorf("checking if dir %q exists: %w", c.OutputDir, err)
		} else if !ok {
			return fmt.Errorf("path %q doesn't exist or is not a directory", c.OutputDir)
		}
	}

	return nil
}

type Output struct {
	Name string
	// ExpectedPath is where this file should be.
	ExpectedPath string

	Content io.Reader
}

func Generate(sc *ir.Statechart, config *Config) ([]*Output, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	tm, err := newTemplateManager(config, sc)
	if err != nil {
		return nil, fmt.Errorf("building new template manager: %w", err)
	}

	forwardDeclare, declare, define, err := tm.generateTemplates()
	if err != nil {
		return nil, fmt.Errorf("generating templates: %w", err)
	}

	outputs, err := CalculateExpectedPaths(config)
	if err != nil {
		return nil, fmt.Errorf("calculating expected output paths: %w", err)
	}

	for _, output := range outputs {
		switch output.Name {
		case kOutputName_ForwardDeclare:
			output.Content = forwardDeclare
		case kOutputName_Declare:
			output.Content = declare
		case kOutputName_Define:
			output.Content = define
		default:
			return nil, fmt.Errorf("unknown output %q", output.Name)
		}
	}

	return outputs, nil
}

// CalculateExpectedPaths will return some outputs without content, just predicting where the
// output files are going to be. This can be useful to compare with the existing paths, for example
// to see if any generation has to occur.
func CalculateExpectedPaths(config *Config) ([]*Output, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	outputs := []*Output{
		{
			Name:         kOutputName_ForwardDeclare,
			ExpectedPath: filepath.Join(config.OutputDir, fmt.Sprintf("%s.forward_declare.h", config.FilePrefix)),
		},
		{
			Name:         kOutputName_Declare,
			ExpectedPath: filepath.Join(config.OutputDir, fmt.Sprintf("%s.declare.h", config.FilePrefix)),
		},
		{
			Name:         kOutputName_Define,
			ExpectedPath: filepath.Join(config.OutputDir, fmt.Sprintf("%s.define.h", config.FilePrefix)),
		},
	}

	return outputs, nil
}
