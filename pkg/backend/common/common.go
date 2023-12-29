// Package common represents common types that is to be used by all backend packages.
package common

import (
	"io"
)

// BackendOutput represents an output file as calculated by a backend.
// Has all the information needed so that callers can take appropiate actions, such as printing
// them, writing them to files, etc.
type BackendOutput struct {
	Name string
	// ExpectedPath is the path where this output file should be written to.
	ExpectedPath string
	// Whether this file is already up to date in disk, so it was not generated.
	UpToDate bool
	// The content of the file if it was generated.
	Content io.Reader
}
