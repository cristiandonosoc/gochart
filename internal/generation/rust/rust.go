// Package rust takes care of generating Rust code for a gochart.
package rust

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

const kOutputName = "rust"

type Config struct {
	Metadata *common.Metadata

	OutputDir  string
	FilePrefix string
}

func (c *Config) validate() error {
	if c.Metadata == nil {
		return fmt.Errorf("no metadata present!")
	}

	if err := c.Metadata.Validate(); err != nil {
		return fmt.Errorf("validating metadata: %w", err)
	}

	if c.OutputDir == "" {
		return fmt.Errorf("missing output dir")
	}

	if ok, err := files.DirExists(c.OutputDir); err != nil {
		return fmt.Errorf("checking if dir %q exists: %w", c.OutputDir, err)
	} else if !ok {
		return fmt.Errorf("path %q doesn't exist or is not a directory", c.OutputDir)
	}

	return nil
}

// GenerateContent returns the generated Rust code for |gc|.
func GenerateContent(gc *ir.Gochart, metadata *common.Metadata) (string, error) {
	m, err := newModel(gc)
	if err != nil {
		return "", fmt.Errorf("gochart %q: %w", gc.Name, err)
	}

	return m.generate(metadata), nil
}

func Generate(gc *ir.Gochart, config *Config) ([]*common.BackendOutput, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	content, err := GenerateContent(gc, config.Metadata)
	if err != nil {
		return nil, err
	}

	return []*common.BackendOutput{
		{
			Name:         kOutputName,
			ExpectedPath: ExpectedPath(config.OutputDir, config.FilePrefix),
			Metadata:     config.Metadata,
			Content:      strings.NewReader(content),
		},
	}, nil
}

// ExpectedPath is where the Rust file for a gochart is generated.
func ExpectedPath(outputDir, filePrefix string) string {
	return filepath.Join(outputDir, fmt.Sprintf("%s.gochart.rs", filePrefix))
}
