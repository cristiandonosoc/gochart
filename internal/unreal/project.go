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

func NewProject(projectDir string) (*Project, error) {
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, fmt.Errorf("making %q abs: %w", projectDir, err)
	}
	projectDir = abs

	sourceDir := filepath.Join(projectDir, "Source")
	if exists, err := files.DirExists(sourceDir); err != nil {
		return nil, fmt.Errorf("querying source dir %q: %w", sourceDir, err)
	} else if !exists {
		return nil, fmt.Errorf("source dir %q does not exists", sourceDir)
	}

	project := &Project{
		ProjectDir: projectDir,
	}

	// TODO(cdc): Determine unreal version.

	return project, nil
}

func (p *Project) IsIndexed() bool {
	return len(p.Modules) > 0
}

func (p *Project) SourceDir() string {
	return filepath.Join(p.ProjectDir, "Source")
}

// IndexModules goes and collects all the modules within the project.
func (p *Project) IndexModules(ctx context.Context) error {
	modules, err := collectModules(ctx, p.SourceDir())
	if err != nil {
		return fmt.Errorf("collecting modules: %w", err)
	}

	if len(modules) == 0 {
		return fmt.Errorf("no modules found at %q. Is it an Unreal project?", p.ProjectDir)
	}

	// Make sure all the modules point back to the project.
	for _, module := range modules {
		module.project = p
	}
	p.Modules = modules

	return nil
}

func (p *Project) NewFile(path string) (*File, error) {
	var module *Module

	isIntermediate := strings.Contains(path, "Intermediate")
	if !isIntermediate {
		m, err := p.identifyModule(path)
		if err != nil {
			return nil, fmt.Errorf("identifying module for %q: %w", path, err)
		}
		module = m
	}

	stat, found, err := files.StatFile(path)
	if err != nil || !found {
		return nil, fmt.Errorf("stating %q: %w (found: %t)", path, err, found)
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
// |extensions| should match a string.HasSuffix over the filepath.Base.
func (p *Project) SearchForFilesByExtension(ctx context.Context, platform Platform, extensions []string) ([]string, error) {
	if len(extensions) == 0 {
		return nil, fmt.Errorf("no extension given to search")
	}

	if len(p.Modules) == 0 {
		return nil, fmt.Errorf("no modules loaded. Is the project indexed?")
	}

	// We search the gocharts in a parallel fashion.
	g, ctx := errgroup.WithContext(ctx)

	// Produce: all the modules to check.
	modulesCh := make(chan *Module)
	{
		g.Go(func() error {
			defer close(modulesCh)

			for _, module := range p.Modules {
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

		for i := 0; i < kModuleSearcherWorkerCount; i++ {
			wg.Add(1)
			g.Go(func() error {
				defer wg.Done()

				for module := range modulesCh {
					// Ensure that we have loaded the UHT files.
					uhtFiles, err := module.LoadUHTFiles(platform, false)
					if err != nil {
						return fmt.Errorf("loading uht files for module %q: %w", module.Name, err)
					}

					// Collect the common + UHT files into one list.
					files := make([]string, 0, len(module.Files)+len(uhtFiles))
					files = append(files, module.Files...)
					files = append(files, uhtFiles...)

					for _, file := range files {
						base := filepath.Base(file)

						// We see if the file matches any of the required extensions.
						match := false
						for _, ext := range extensions {
							if strings.HasSuffix(base, ext) {
								match = true
								break
							}
						}

						if !match {
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

		// Make sure the channel will be closed.
		g.Go(func() error {
			wg.Wait()
			close(gochartsCh)
			return nil
		})
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

func (p *Project) identifyModule(path string) (*Module, error) {
	// Search over all the modules and see which one this belongs to.
	// Because a lot of modules could be "candidates", we keep the one with the longest path to be the
	// one that actually contains the file.
	// TODO(cdc): If this becomes a slow operation, it could be done in parallel.
	var candidate *Module
	for _, module := range p.Modules {
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
