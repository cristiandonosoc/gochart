package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (gc *GunrealConfig) Describe() string {
	var sb strings.Builder

	sb.WriteString("CONFIG FILE --------------------------------------------------------------\n\n")
	sb.WriteString(fmt.Sprintf("- PATH: %s\n", gc.Path))
	sb.WriteString(fmt.Sprintf("- UPROJECT: %s\n", gc.UProject))
	sb.WriteString(fmt.Sprintf("- PROJECT DIR: %s\n", gc.ProjectDir))

	if gc.EditorDir != "" {
		sb.WriteString(fmt.Sprintf("- EDITOR DIR: %s\n", gc.EditorDir))
	}

	return sb.String()
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
	if abs, err := gc.checkFile(gc.UProject); err != nil {
		return fmt.Errorf("uproject: %w", err)
	} else {
		gc.UProject = abs
	}

	if abs, err := gc.checkFile(gc.EditorDir); err != nil {
		return fmt.Errorf("editor_key: %w", err)
	} else {
		gc.EditorDir = abs
	}

	return nil
}

func (gc *GunrealConfig) checkFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("key not set")
	}

	// If the path given is not absolute, we will make it relative to the config file.
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(gc.Path), path)

		abs, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("abs %q: %w", path, err)
		}
		path = abs
	}
	path = filepath.Clean(path)

	// TODO(cdc): Use StatFileErrorf
	if _, found, err := files.StatFile(path); err != nil || !found {
		return "", files.StatFileErrorf(err, "statting %q", path)
	}

	return path, nil
}
