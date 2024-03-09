package unreal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
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

	project  *Project
	uhtFiles map[Platform][]string
}

// newUnrealModule generates a new module definition from the path to its build file.
// Assumes that the entry path has been cleaned with filepath.Clean
// Name can be empty and will be taken from the name of the file.
func newUnrealModule(name string, buildFile string, files []string) (*Module, error) {
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
		uhtFiles:  map[Platform][]string{},
	}, nil
}

func (m *Module) String() string {
	return fmt.Sprintf("%s (%s)", filepath.Base(m.BaseDir), m.BaseDir)
}

// Contains returns whether a particular path is within this module.
// Assumes that the entry |path| has been cleaned with filepath.Clean
func (m *Module) Contains(path string) bool {
	return strings.HasPrefix(path, m.BaseDir)
}

// LoadUHTFiles makes this module load the UHT files associated with this module for this platform.
// |reload| forces the previous cached results to be overwritten. Otherwise the previous results
// will be returned.
func (m *Module) LoadUHTFiles(platform Platform, reload bool) ([]string, error) {
	files, ok := m.uhtFiles[platform]
	if ok && !reload {
		return files, nil
	}

	// If we're here we need to query the list for this module.
	uhtDir := filepath.Join(m.project.ProjectDir(), "Intermediate", "Build", platform.String())
	uhtDir = filepath.Join(uhtDir, "UnrealEditor", "Inc", m.Name, "UHT")

	var uhtFiles []string
	err := filepath.WalkDir(uhtDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("path %q: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		uhtFiles = append(uhtFiles, path)
		return nil
	})

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("walking UHT dir %q: %w", uhtDir, err)
		}
	}

	m.uhtFiles[platform] = uhtFiles
	return uhtFiles, nil
}
