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

	ProjectName string `yaml:"project_name"`
	UProjectPath string `yaml:"uproject"`
	// (optional) Where the base project is.
	// If not set, it is calculated as the directory that holds the |UProject| file.
	ProjectDir string `yaml:"project_dir"`

	// *** Editor fields ***

	EditorConfig *GunrealEditorConfig `yaml:"editor"`

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

	sb.WriteString(fmt.Sprintf("CONFIG PATH: %s\n", gc.Path))
	sb.WriteString("\n")

	sb.WriteString("PROJECT ------------------------------------------------------------------\n\n")
	sb.WriteString(fmt.Sprintf("- Name: %s\n", gc.ProjectName))
	sb.WriteString(fmt.Sprintf("- UPROJECT: %s\n", gc.UProjectPath))
	sb.WriteString(fmt.Sprintf("- PROJECT DIR: %s\n", gc.ProjectDir))

	if gc.EditorConfig != nil {
		sb.WriteString("\n")
		sb.WriteString(gc.EditorConfig.Describe())
	}

	return sb.String()
}

func (gc *GunrealConfig) resolve() error {
	if err := gc.sanityCheck(); err != nil {
		return fmt.Errorf("sanity checking config: %w", err)
	}

	// Check the project dir.
	if gc.ProjectDir == "" {
		gc.ProjectDir = filepath.Dir(gc.UProjectPath)
	}

	if err := resolveEditorConfig(gc.Path, gc.EditorConfig); err != nil {
		return fmt.Errorf("reading editor config: %w", err)
	}

	return nil
}

func (gc *GunrealConfig) sanityCheck() error {
	if gc.ProjectName == "" {
		return fmt.Errorf("project_name not set")
	}

	if uproject, err := checkFile(gc.Path, gc.UProjectPath); err != nil {
		return fmt.Errorf("uproject: %w", err)
	} else {
		gc.UProjectPath = uproject
	}

	return nil
}

func checkFile(configPath, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("key not set")
	}

	abs, err := resolveConfigPath(configPath, path)
	if err != nil {
		return "", fmt.Errorf("resolving config path: %w", err)
	}
	path = abs

	// If the path given is not absolute, we will make it relative to the config file.
	if _, found, err := files.StatFile(path); err != nil || !found {
		return "", files.StatFileErrorf(err, "statting %q", path)
	}

	return path, nil
}

func resolveConfigPath(configPath, path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(configPath), path)

		abs, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("abs %q: %w", path, err)
		}
		path = abs
	}
	path = filepath.Clean(path)

	return path, nil

}
