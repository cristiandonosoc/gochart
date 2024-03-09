package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cristiandonosoc/golib/pkg/files"

	"gopkg.in/yaml.v2"
)

type GunrealConfig struct {
	// *** Project fields ***

	UProject string `yaml:"uproject"`
	// (optional) Where the base project is.
	// If not set, it is calculated as the directory that holds the |UProject| file.
	ProjectDir string `yaml:"project_dir"`

	// *** Editor fields ***

	EditorDir string `yaml:"editor"`
	// UBT will normally be discovered via the editor, but this option can be used to override.
	UBT string `yaml:"ubt"`

	Path string
}

func LoadConfig(path string) (*GunrealConfig, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("abs %q: %w", path, err)
	}
	path = abs

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}

	gc := &GunrealConfig{
		Path: path,
	}
	if err := yaml.UnmarshalStrict(data, gc); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}

	if err := gc.resolve(); err != nil {
		return nil, fmt.Errorf("resolving config: %w", err)
	}

	return gc, err
}

func (gc *GunrealConfig) resolve() error {
	if err := gc.sanityCheck(); err != nil {
		return fmt.Errorf("sanity checking config: %w", err)
	}

	// Check the project key.
	if gc.ProjectDir == "" {
		gc.ProjectDir = filepath.Dir(gc.UProject)
	}

	return nil
}

func (gc *GunrealConfig) sanityCheck() error {
	if abs, err := checkFile(gc.UProject); err != nil {
		return fmt.Errorf("uproject: %w", err)
	} else {
		gc.UProject = abs
	}

	if abs, err := checkFile(gc.EditorDir); err != nil {
		return fmt.Errorf("editor_key: %w", err)
	} else {
		gc.EditorDir = abs
	}

	return nil
}

func checkFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("key not set")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("abs %q: %w", path, err)
	}
	path = abs

	// TODO(cdc): Use StatFileErrorf
	if _, found, err := files.StatFile(path); err != nil || !found {
		return "", files.StatFileErrorf(err, "statting %q", path)
	}

	return path, nil
}
