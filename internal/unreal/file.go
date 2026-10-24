package unreal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cristiandonosoc/golib/pkg/files"
)

const (
	UnrealBuildFileExtension = ".build.cs"
)

// We hackily search for a ModuleRules class definition.
var identifier = "[a-zA-Z0-9_]"
var moduleRulesRegex = regexp.MustCompile(fmt.Sprintf(`public\s+class\s+(%s+)\s+:\s+ModuleRules`, identifier))

// IsUnrealBuildFile returns whether the path points to an unreal build file.
// These are normally C# files that end with the .Build.cs extension.
// We search for a common regex pattern than we should find.
// Returns the defined module name when found.
func IsUnrealBuildFile(path string) (string, bool, error) {
	if !strings.HasSuffix(strings.ToLower(path), UnrealBuildFileExtension) {
		return "", false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, fmt.Errorf("reading file %q: %w", path, err)
	}

	// If it doesn't match the regex, it means that we don't consider this to be a module file.
	matches := moduleRulesRegex.FindSubmatch(data)
	if len(matches) == 0 {
		return "", false, nil
	}

	moduleName := matches[1]
	return string(moduleName), true, nil
}

// File represents a file within an Unreal project.
type File struct {
	Path         string
	Module       *Module
	FileInfo     fs.FileInfo
	Intermediate bool
}

func (uf *File) String() string {
	var sb strings.Builder
	sb.WriteString(uf.Name())

	if uf.Module != nil {
		sb.WriteString(fmt.Sprintf(" [module:%s]", uf.Module.Name))
	}

	if uf.FileInfo != nil {
		sb.WriteString(fmt.Sprintf(" [mod:%v]", uf.FileInfo.ModTime()))
	}

	sb.WriteString(fmt.Sprintf(" [path:%s]", uf.Path))

	return sb.String()
}

// Name is just the filename of the file (akin to filepath.Base), without the extension.
func (uf *File) Name() string {
	base := filepath.Base(uf.Path)
	extension := filepath.Ext(base)
	return strings.TrimSuffix(base, extension)
}

// ModulePath is the path of the file _within_ the module (basically stripping the module path).
func (uf *File) ModulePath() string {
	path := files.ToUnixPath(strings.TrimPrefix(uf.Path, uf.Module.BaseDir))
	return strings.TrimPrefix(path, "/")
}

// Exists returns whether the given file exists in disk or not.
func (uf *File) Exists() bool {
	return uf.FileInfo != nil
}

// ModTime returns the latest modification time of the file.
// If the file does not exist, returns the earliest time possible.
func (uf *File) ModTime() time.Time {
	if !uf.Exists() {
		return time.Time{}
	}

	return uf.FileInfo.ModTime()
}
