package unreal

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var moduleFilenameRegexPattern = `(.+?)\.(?i:build\.cs)$`
var moduleFilenameRegex = regexp.MustCompile(moduleFilenameRegexPattern)

// Module represents an unreal module.
type Module struct {
	Name      string
	BaseDir   string
	BuildFile string
	Files     []string
}

// NewUnrealModule generates a new module definition from the path to its build file.
// Assumes that the entry path has been cleaned with filepath.Clean
// Name can be empty and will be taken from the name of the file.
func NewUnrealModule(name string, buildFile string, files []string) (*Module, error) {
	baseDir := filepath.Dir(buildFile)

	if name == "" {
		matches := moduleFilenameRegex.FindStringSubmatch(filepath.Base(buildFile))
		if len(matches) == 0 {
			return nil, fmt.Errorf("build file path %q does not match regex %q", buildFile, moduleFilenameRegexPattern)
		}

		name = matches[0]
	}

	return &Module{
		Name:      name,
		BaseDir:   baseDir,
		BuildFile: buildFile,
		Files:     files,
	}, nil
}

func (um *Module) String() string {
	return fmt.Sprintf("%s (%s)", filepath.Base(um.BaseDir), um.BaseDir)
}

// Contains returns whether a particular path is within this module.
// Assumes that the entry |path| has been cleaned with filepath.Clean
func (um *Module) Contains(path string) bool {
	return strings.HasPrefix(path, um.BaseDir)
}
