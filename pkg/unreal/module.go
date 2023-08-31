package unreal

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	UnrealBuildFileExtension = ".build.cs"
)

// UnrealFile represents a file within an Unreal project.
type UnrealFile struct {
	Path     string
	Module   *UnrealModule
	FileInfo fs.FileInfo
}

// Name is just the filename of the file (akin to filepath.Base).
func (uf *UnrealFile) Name() string {
	return filepath.Base(uf.Path)
}

// InternalPath is the path of the file _within_ the module (basically stripping the module path).
func (uf *UnrealFile) InternalPath() string {
	return strings.TrimPrefix(uf.Path, uf.Module.BaseDir)
}

// UnrealModule represents an unreal module.
type UnrealModule struct {
	Name      string
	BaseDir   string
	BuildFile string
}

func NewUnrealModule(buildFile string) *UnrealModule {
	bf := filepath.Clean(buildFile)
	return &UnrealModule{
		Name:      filepath.Base(bf),
		BaseDir:   filepath.Dir(bf),
		BuildFile: bf,
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
