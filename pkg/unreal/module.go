package unreal

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	UnrealBuildFileExtension = ".build.cs"
)

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

func IsUnrealBuildFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), UnrealBuildFileExtension)
}

func (um *UnrealModule) String() string {
	return fmt.Sprintf("%s (%s)", filepath.Base(um.BaseDir), um.BaseDir)
}

func (um *UnrealModule) Contains(path string) bool {
	return strings.HasPrefix(filepath.Clean(path), um.BaseDir)
}
