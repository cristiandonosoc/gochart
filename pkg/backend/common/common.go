// Package common represents common types that is to be used by all backend packages.
package common

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/cristiandonosoc/gunreal/pkg/unreal"
)

type Metadata struct {
	// Version establishes of Gochart was used to generate this file.
	Version string

	// IsUnrealGochart determines whether this is a gochart generated as part of an unreal project.
	IsUnrealGochart bool

	// UnrealModule represents the module name where this gochart has been installed in.
	// Only required when |IsUnrealGochart| is true.
	UnrealModule string

	// ModulePath is the module internal path to the gochart file.
	// This can be used to determine whether the generation gochart file still exists.
	// Only required when |IsUnrealGochart| is true.
	GochartModulePath string

	Stat fs.FileInfo
}

func (m *Metadata) Validate() error {
	if m.Version == "" {
		return fmt.Errorf("missing version")
	}

	if m.IsUnrealGochart {
		if m.UnrealModule == "" {
			return fmt.Errorf("missing unreal module")
		}

		if m.GochartModulePath == "" {
			return fmt.Errorf("missing module path")
		}
	}

	return nil
}

func (m *Metadata) FullGochartPath(project *unreal.Project) (string, error) {
	// Find the module where the gochart is supposed to be.
	// TODO(cdc): This might not considered be an error in the future, depending on the clean
	//			  semantics that we want to express.
	module, ok := project.Modules[m.UnrealModule]
	if !ok {
		return "", fmt.Errorf("module %q not found", m.UnrealModule)
	}

	path := filepath.Join(module.BaseDir, m.GochartModulePath)
	return path, nil
}

// BackendOutput represents an output file as calculated by a backend.
// Has all the information needed so that callers can take appropriate actions, such as printing
// them, writing them to files, etc.
type BackendOutput struct {
	Name string
	// ExpectedPath is the path where this output file should be written to.
	ExpectedPath string

	Metadata *Metadata

	Content     io.Reader
	DiskContent *files.LoadedFile
}

func (bo *BackendOutput) IsLoaded() bool {
	return bo.DiskContent != nil
}

// EnsureLoad makes sure that the file and metadata are loaded.
func (bo *BackendOutput) EnsureLoadFromDisk() error {
	if bo.IsLoaded() {
		return nil
	}

	lf, err := files.LoadFileFromPath(bo.ExpectedPath)
	if err != nil {
		return fmt.Errorf("loading file from path %q: %w", bo.ExpectedPath, err)
	}

	bo.DiskContent = lf
	return nil
}
