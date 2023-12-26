package unreal

/*

import (
	"context"
	"fmt"
	"path/filepath"
)

func (ui *unrealIntegrator) indexUnrealProject(ctx context.Context, projectDir, gochartExtension string) error {
	sourceDir := filepath.Join(projectDir, "Source")
	gochartPaths, modules, err := queryUnrealSources(ctx, sourceDir, GochartExtension)
	if err != nil {
		return fmt.Errorf("querying unreal project at %q: %w", projectDir, err)
	}

	ui.Modules = modules

	// Now that we have the modules, we can identify what module this belongs to.
	gocharts := make([]*UnrealFile, 0, len(gochartPaths))
	for _, path := range gochartPaths {
		gochart, err := ui.NewUnrealFile(path)
		if err != nil {
			return fmt.Errorf("creating unreal path for %q: %w", path, err)
		}
		gocharts = append(gocharts, gochart)
	}
	ui.Gocharts = gocharts

	return nil
}

func (ui *unrealIntegrator) identifyModule(path string) (*UnrealModule, error) {
	if len(ui.Modules) == 0 {
		return nil, fmt.Errorf("no modules found. Did you call indexUnrealProject before?")
	}

	// Search over all the modules and see which one this belongs to.
	// Because a lot of modules could be "candidates", we keep the one with the longest path to be the
	// one that actually contains the file.
	var candidate *UnrealModule
	for _, module := range ui.Modules {
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
*/
