package common

import (
	"github.com/cristiandonosoc/gochart/internal/files"
)

type FrontendEntry interface {
	// Index is the frontend order in which they appeared in the frontend file.
	// This is comparison to other FrontendEntries.
	Index() int

	// Position is the location of this entry in the *original* source file.
	Position() *files.LoadedFilePosition
}
