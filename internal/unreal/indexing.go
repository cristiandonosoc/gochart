package unreal

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// collectModules scans the whole |Source| directory of an unreal project.
// Indexes all the files within a project, for faster in memory searching afterwards.
func collectModules(sourceDir string) (map[string]*Module, error) {
	// collect all the files in the unreal project.
	result, err := collectFiles(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("collecting files in %q: %w", sourceDir, err)
	}

	// For each build file, we collect all the files that belong to it.
	modules := make(map[string]*Module)
	for _, bfd := range result.buildFiles {
		baseDir := filepath.Dir(bfd.Path)

		// Search for all the files that are under this path.
		// We do it by searching for the build file in a binary search and then searching "upwards
		// and downwards" for the files in the same dir. At soon as we find one that doens't belong,
		// we can stop searching because the file list is sorted.
		index, ok := slices.BinarySearch(result.allFiles, bfd.Path)
		if !ok {
			return nil, fmt.Errorf("build file %q not found in the all file list", bfd.Path)
		}

		moduleFiles := []string{
			bfd.Path,
		}

		// Search backwards from the build file index.
		for i := index - 1; i >= 0; i-- {
			file := result.allFiles[i]
			if strings.Contains(file, baseDir) {
				moduleFiles = append(moduleFiles, file)
				continue
			}

			// As soon as we don't find a file, it means that we're "out of this dir range".
			break
		}

		// Search forward from the build file index.
		for i := index + 1; i < len(result.allFiles); i++ {
			file := result.allFiles[i]
			if strings.Contains(file, baseDir) {
				moduleFiles = append(moduleFiles, file)
				continue
			}

			// As soon as we don't find a file, it means that we're "out of this dir range".
			break
		}

		// Now moduleFiles has all the files that belong to this module.
		// We sort because sorted lists are cool.
		sort.Strings(moduleFiles)

		um, err := newUnrealModule(bfd.ModuleName, bfd.Path, moduleFiles)
		if err != nil {
			return nil, fmt.Errorf("creating unreal module %q: %w", bfd.ModuleName, err)
		}

		if _, ok := modules[um.Name]; ok {
			return nil, fmt.Errorf("module %q found more than once", um.Name)
		}

		modules[um.Name] = um
	}

	return modules, nil
}

type buildFileDescription struct {
	ModuleName string
	Path       string
}

type collectFilesResult struct {
	buildFiles []*buildFileDescription
	allFiles   []string
}

func collectFiles(sourceDir string) (*collectFilesResult, error) {
	var buildFiles []*buildFileDescription
	var allFiles []string

	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("path %q: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		allFiles = append(allFiles, path)

		// Now we check if it's an unreal build file.
		moduleName, ok, err := IsUnrealBuildFile(path)
		if err != nil {
			return fmt.Errorf("checking if %q is unreal file: %w", path, err)
		}

		if ok {
			buildFiles = append(buildFiles, &buildFileDescription{
				ModuleName: moduleName,
				Path:       path,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking dir %q: %w", sourceDir, err)
	}

	// We sort all the files so we can do faster searches on it.
	sort.Strings(allFiles)

	return &collectFilesResult{
		buildFiles: buildFiles,
		allFiles:   allFiles,
	}, nil
}
