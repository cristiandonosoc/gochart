package unreal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UnrealPlatform string

const (
	UnrealPlatform_Windows = "Win64"
)

// NewUnrealPlatform attempts to unify the unreal platform from identifiers that might come from the
// outside.
func NewUnrealPlatform(id string) (UnrealPlatform, error) {
	switch strings.ToLower(id) {
	case "win64", "windows":
		return UnrealPlatform_Windows, nil
	default:
		return "", fmt.Errorf("unrecognized unreal platform %q", id)
	}
}

func (up *UnrealPlatform) String() string {
	return string(*up)
}

const (
	UnrealBuildFileExtension = ".build.cs"
)

// UnrealFile represents a file within an Unreal project.
type UnrealFile struct {
	Path         string
	Module       *UnrealModule
	FileInfo     fs.FileInfo
	Intermediate bool
}

func (ui *unrealIntegrator) NewUnrealFile(path string) (*UnrealFile, error) {
	intermediate := strings.Contains(path, "Intermediate")

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			stat = nil
		} else {
			return nil, fmt.Errorf("stat %q: %w", path, err)
		}
	}

	var module *UnrealModule
	if !intermediate {
		m, err := ui.identifyModule(path)
		if err != nil {
			return nil, fmt.Errorf("identifying module for %q: %w", path, err)
		}
		module = m
	}

	return &UnrealFile{
		Path:         path,
		Module:       module,
		FileInfo:     stat,
		Intermediate: intermediate,
	}, nil
}

// Name is just the filename of the file (akin to filepath.Base).
func (uf *UnrealFile) Name() string {
	return filepath.Base(uf.Path)
}

// InternalPath is the path of the file _within_ the module (basically stripping the module path).
func (uf *UnrealFile) InternalPath() string {
	return strings.TrimPrefix(uf.Path, uf.Module.BaseDir)
}

// Exists returns whether the given file exists in disk or not.
func (uf *UnrealFile) Exists() bool {
	return uf.FileInfo != nil
}

// ModTime returns the latest modification time of the file.
// If the file does not exist, returns the earliest time possible.
func (uf *UnrealFile) ModTime() time.Time {
	if !uf.Exists() {
		return time.Time{}
	}

	return uf.FileInfo.ModTime()
}

// UnrealModule represents an unreal module.
type UnrealModule struct {
	Name      string
	BaseDir   string
	BuildFile string
}

// NewUnrealModule generates a new module definition from the path to its build file.
// Assumes that the entry path has been cleaned with filepath.Clean
func NewUnrealModule(buildFile string) *UnrealModule {
	baseDir := filepath.Dir(buildFile)
	return &UnrealModule{
		Name:      filepath.Base(baseDir),
		BaseDir:   baseDir,
		BuildFile: buildFile,
	}
}

// IsUnrealBuildFile returns whether the path points to an unreal build file.
// These are normally C# files that end with the .Build.cs extension.
func IsUnrealBuildFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), UnrealBuildFileExtension)
}

func (um *UnrealModule) String() string {
	return fmt.Sprintf("%s (%s)", filepath.Base(um.BaseDir), um.BaseDir)
}

// Contains returns whether a particular path is within this module.
// Assumes that the entry |path| has been cleaned with filepath.Clean
func (um *UnrealModule) Contains(path string) bool {
	return strings.HasPrefix(path, um.BaseDir)
}
