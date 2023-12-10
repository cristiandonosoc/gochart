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
	Name    string
	Content io.Reader
	// ExpectedPath is where this file should be.
	ExpectedPath string
}

func Generate(config *Config, sc *ir.Statechart) ([]*Output, error) {
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

	outputs := []*Output{
		{
			Name:         "forward_declare",
			Content:      forwardDeclare,
			ExpectedPath: filepath.Join(config.OutputDir, fmt.Sprintf("%s.forward_declare.h", config.FilePrefix)),
		},
		{
			Name:         "declare",
			Content:      declare,
			ExpectedPath: filepath.Join(config.OutputDir, fmt.Sprintf("%s.declare.h", config.FilePrefix)),
		},
		{
			Name:         "define",
			Content:      define,
			ExpectedPath: filepath.Join(config.OutputDir, fmt.Sprintf("%s.define.h", config.FilePrefix)),
		},
	}
	return outputs, nil

}
