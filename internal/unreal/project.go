// Package unreal holds common functionality for querying an unreal project, searching for things
// like modules, files, etc.
package unreal

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/cristiandonosoc/golib/pkg/files"

	"golang.org/x/sync/errgroup"
)

const (
	kModuleSearcherWorkerCount    = 100
	kBuildFileSearcherWorkerCount = 100
	kModuleBuilderWorkerCount     = 100
)

// Project represents an indexed Unreal project.
type Project struct {
	ProjectDir string
	Modules    map[string]*Module
}

func IndexProject(ctx context.Context, projectDir string) (*Project, error) {
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

	return &Project{
		ProjectDir: projectDir,
		Modules:    modules,
	}, nil
}

func (up *Project) NewFile(path string) (*File, error) {
	var module *Module

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

	return &File{
		Path:         path,
		Module:       module,
		FileInfo:     stat,
		Intermediate: isIntermediate,
	}, nil
}

// SearchForFilesByExtension goes over all the loaded modules in parallel and finds all the found
// files with that extension. Useful for things like finding all gochart files.
// |extension| should be akin to the result of a filepath.Ext call, which includes the dot.
func (up *Project) SearchForFilesByExtension(ctx context.Context, extension string) ([]string, error) {
	if len(up.Modules) == 0 {
		return nil, fmt.Errorf("no modules loaded. Is the project indexed?")
	}

	// We search the gocharts in a parallel fashion.
	g, ctx := errgroup.WithContext(ctx)

	// Produce: all the modules to check.
	modulesCh := make(chan *Module)
	{
		g.Go(func() error {
			defer close(modulesCh)

			for _, module := range up.Modules {
				select {
				case modulesCh <- module:
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})
	}

	// Map: for each module, search for gochart files.
	gochartsCh := make(chan string)
	{
		var wg sync.WaitGroup

		// Make sure the channel will be closed.
		g.Go(func() error {
			wg.Wait()
			close(gochartsCh)
			return nil
		})

		for i := 0; i < kModuleSearcherWorkerCount; i++ {
			wg.Add(1)
			g.Go(func() error {
				defer wg.Done()

				for module := range modulesCh {
					for _, file := range module.Files {
						ext := filepath.Ext(file)
						if ext != extension {
							continue
						}

						select {
						case gochartsCh <- file:
							continue
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}

				return nil
			})
		}
	}

	// Reduce: collect gochart files.
	gochartSet := map[string]struct{}{}
	{
		g.Go(func() error {
			for gochart := range gochartsCh {
				if _, ok := gochartSet[gochart]; ok {
					return fmt.Errorf("gochart %q found more than once", gochart)
				}

				gochartSet[gochart] = struct{}{}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Collect the results in an array and sort it because sorting is cool.
	gocharts := make([]string, 0, len(gochartSet))
	for gochart := range gochartSet {
		gocharts = append(gocharts, gochart)
	}
	sort.Strings(gocharts)

	return gocharts, nil
}

func (up *Project) identifyModule(path string) (*Module, error) {
	// Search over all the modules and see which one this belongs to.
	// Because a lot of modules could be "candidates", we keep the one with the longest path to be the
	// one that actually contains the file.
	// TODO(cdc): If this becomes a slow operation, it could be done in parallel.
	var candidate *Module
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
