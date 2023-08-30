package unreal

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// GochartFile represents a Gochart file found within an Unreal module.
type UnrealGochart struct {
	Path string
}

func indexUnrealProject(ctx context.Context, sourceDir, gochartExtension string) ([]string, []*UnrealModule, error) {
	gochartExtension = strings.ToLower(gochartExtension)

	g, ctx := errgroup.WithContext(ctx)

	// Produce: all the directories to check.
	dirsToCheckCh := make(chan string)
	{
		g.Go(func() error {
			defer close(dirsToCheckCh)

			err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					return nil
				}

				select {
				case dirsToCheckCh <- path:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})

			if err != nil {
				return fmt.Errorf("walking dir %q: %w", sourceDir, err)
			}

			return nil
		})
	}

	// Map: Verify each directory for gochart files.
	foundGochartCh := make(chan string)
	foundBuildFileCh := make(chan string)
	{
		workers := 100
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			g.Go(func() error {
				wg.Add(1)
				defer wg.Done()

				for dir := range dirsToCheckCh {
					files, err := os.ReadDir(dir)
					if err != nil {
						return fmt.Errorf("reading dir %q: %w", dir, err)
					}

					for _, file := range files {
						if file.IsDir() {
							continue
						}

						// If we found an gochart file, we send it to the channel.
						if filepath.Ext(file.Name()) == gochartExtension {
							select {
							case foundGochartCh <- filepath.Join(dir, file.Name()):
								continue
							case <-ctx.Done():
								return ctx.Err()
							}
						}

						// If we find an build file, we send it to the channel.
						if IsUnrealBuildFile(file.Name()) {
							select {
							case foundBuildFileCh <- filepath.Join(dir, file.Name()):
								continue
							case <-ctx.Done():
								return ctx.Err()
							}
						}
					}
				}

				return nil
			})
		}

		// Close the channels.
		g.Go(func() error {
			wg.Wait()
			close(foundGochartCh)
			close(foundBuildFileCh)
			return nil
		})
	}

	// Reduce: all the found gocharts and build files into a list.
	gochartSet := make(map[string]struct{})
	moduleSet := make(map[string]struct{})
	{
		g.Go(func() error {
			for gochart := range foundGochartCh {
				if _, ok := gochartSet[gochart]; ok {
					return fmt.Errorf("gochart %q indexed twice!", gochart)
				}
				gochartSet[gochart] = struct{}{}
			}
			return nil
		})

		g.Go(func() error {
			for buildFile := range foundBuildFileCh {
				buildFile := filepath.Clean(buildFile)
				if _, ok := moduleSet[buildFile]; ok {
					return fmt.Errorf("module build file %q indexed twice!", buildFile)
				}
				moduleSet[buildFile] = struct{}{}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	gocharts := make([]string, 0, len(gochartSet))
	for gochart := range gochartSet {
		gocharts = append(gocharts, gochart)
	}
	sort.Strings(gocharts)

	modules := make([]*UnrealModule, 0, len(moduleSet))
	for buildFile := range moduleSet {
		modules = append(modules, NewUnrealModule(buildFile))
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].BaseDir < modules[j].BaseDir
	})

	return gocharts, modules, nil
}
