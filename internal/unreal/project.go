// Package unreal holds common functionality for querying an unreal project, searching for things
// like modules, files, etc.
package unreal

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/files"
)

const (
	kModuleSearcherWorkerCount    = 100
	kBuildFileSearcherWorkerCount = 100
	kModuleBuilderWorkerCount     = 100
)

// UnrealProject represets an indexed Unreal project.
type UnrealProject struct {
	ProjectDir string
	Modules    map[string]*UnrealModule
}

func IndexUnrealProject(ctx context.Context, projectDir string) (*UnrealProject, error) {
	sourceDir := filepath.Join(projectDir, "Source")
	if exists, err := files.DirExists(sourceDir); err != nil {
		return nil, fmt.Errorf("querying source dir %q: %w", sourceDir, err)
	} else if !exists {
		return nil, fmt.Errorf("source dir %q does not exists", sourceDir)
	}

	modules, err := collectModules(ctx, sourceDir)
	if err != nil {
		return nil, fmt.Errorf("collecting modules: %w", err)
	}

	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules found at %q. Is it an Unreal project?", projectDir)
	}

	return &UnrealProject{
		ProjectDir: projectDir,
		Modules:    modules,
	}, nil
}

func (up *UnrealProject) NewUnrealFile(path string) (*UnrealFile, error) {
	var module *UnrealModule

	isIntermediate := strings.Contains(path, "Intermediate")
	if !isIntermediate {
		m, err := up.identifyModule(path)
		if err != nil {
			return nil, fmt.Errorf("identifying module for %q: %w", path, err)
		}
		module = m
	}

	stat, err := files.StatFile(path)
	if err != nil {
		return nil, err
	}

	return &UnrealFile{
		Path:         path,
		Module:       module,
		FileInfo:     stat,
		Intermediate: isIntermediate,
	}, nil
}

func (up *UnrealProject) identifyModule(path string) (*UnrealModule, error) {
	// Search over all the modules and see which one this belongs to.
	// Because a lot of modules could be "candidates", we keep the one with the longest path to be the
	// one that actually contains the file.
	// TODO(cdc): If this becomes a slow operation, it could be done in parallel.
	var candidate *UnrealModule
	for _, module := range up.Modules {
		if module.Contains(path) {
			// If there is no current candidate, this is our current candidate.
			if candidate == nil {
				candidate = module
				continue
			}

			// Otherwise we want the candidate with the longest base directory.
			if len(module.BaseDir) > len(candidate.BaseDir) {
				candidate = module
				continue
			}
		}
	}

	if candidate == nil {
		return nil, fmt.Errorf("no module contains %q", path)
	}

	return candidate, nil
}
