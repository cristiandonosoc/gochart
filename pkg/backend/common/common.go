// Package common represents common types that is to be used by all backend packages.
package common

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/internal/files"
)

type Metadata struct {
	// What version of Gochart was used to generate this file.
	Version string
}

// BackendOutput represents an output file as calculated by a backend.
// Has all the information needed so that callers can take appropiate actions, such as printing
// them, writing them to files, etc.
type BackendOutput struct {
	Name string
	// ExpectedPath is the path where this output file should be written to.
	ExpectedPath string
	// Whether this file is already up to date in disk, so it was not generated.
	UpToDate bool

	Metadata *Metadata

	Content     io.Reader
	DiskContent *files.LoadedFile
}

func (bo *BackendOutput) IsLoadedFromDisk() bool {
	return bo.DiskContent != nil
}

func (bo *BackendOutput) LoadFromDisk() error {
	if bo.IsLoadedFromDisk() {
		return nil
	}

	lf, err := files.LoadFileFromPath(bo.ExpectedPath)
	if err != nil {
		return fmt.Errorf("loading file from path %q: %w", bo.ExpectedPath, err)
	}

	bo.DiskContent = lf
	return nil
}
