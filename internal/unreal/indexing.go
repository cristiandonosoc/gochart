package unreal

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// collectModules scans the whole |Source| directory of an unreal project in a parallel fashion.
// Indexes all the files within a project, for faster in memory searching afterwards.
func collectModules(ctx context.Context, sourceDir string) (map[string]*Module, error) {
	// collect all the files in the unreal project.
	result, err := collectFiles(ctx, sourceDir)
	if err != nil {
		return nil, fmt.Errorf("collecting files in %q: %w", sourceDir, err)
	}

	// For each build file, we generate a worker that will collect all the files to it.
	g, ctx := errgroup.WithContext(ctx)

	// Produce: all the build files to check.
	buildFileDescriptionsCh := make(chan *buildFileDescription)
	{
		g.Go(func() error {
			defer close(buildFileDescriptionsCh)

			for _, bfd := range result.buildFiles {
				select {
				case buildFileDescriptionsCh <- bfd:
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})
	}

	// Map: each module collect all the files that belong to it.
	modulesCh := make(chan *Module)
	{
		var wg sync.WaitGroup

		// Create a new worker to consume the channel.
		for i := 0; i < kModuleBuilderWorkerCount; i++ {
			wg.Add(1)
			g.Go(func() error {
				defer wg.Done()

				for bfd := range buildFileDescriptionsCh {
					baseDir := filepath.Dir(bfd.Path)

					// Search for all the files that are under this path.
					// We do it by searching for the build file in a binary search and then searching "upwards
					// and downwards" for the files in the same dir. At soon as we find one that doens't belong,
					// we can stop searching because the file list is sorted.
					index, ok := slices.BinarySearch(result.allFiles, bfd.Path)
					if !ok {
						return fmt.Errorf("build file %q not found in the all file list", bfd.Path)
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
						return fmt.Errorf("creating unreal module %q: %w", bfd.ModuleName, err)
					}

					select {
					case modulesCh <- um:
						continue
					case <-ctx.Done():
						return ctx.Err()
					}
				}

				return nil
			})
		}

		// Make sure the channel will be closed.
		g.Go(func() error {
			wg.Wait()
			close(modulesCh)
			return nil
		})
	}

	// Reduce: Collect all the generated modules.
	modules := make(map[string]*Module)
	{
		g.Go(func() error {
			for module := range modulesCh {
				if _, ok := modules[module.Name]; ok {
					return fmt.Errorf("module %q found more than once", module.Name)
				}

				modules[module.Name] = module
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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

// func collectFiles(ctx context.Context, sourceDir string) (_buildFiles []*buildFileDescription, _allFiles []string, _err error) {
func collectFiles(ctx context.Context, sourceDir string) (*collectFilesResult, error) {
	g, ctx := errgroup.WithContext(ctx)

	// Produce: generate all the directories to be checked.
	dirsToCheckCh := make(chan string)
	{
		g.Go(func() error {
			defer close(dirsToCheckCh)

			err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("path %q: %w", path, err)
				}

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

	// Map: Generate all the valid files found in the directories.
	foundFilesCh := make(chan string)
	{
		var wg sync.WaitGroup

		// Ensure the channel gets closed.
		g.Go(func() error {
			wg.Wait()
			close(foundFilesCh)
			return nil
		})

		// Create a new worker to search for the files.
		for i := 0; i < kModuleSearcherWorkerCount; i++ {
			wg.Add(1)
			g.Go(func() error {
				defer wg.Done()

				for dir := range dirsToCheckCh {
					files, err := os.ReadDir(dir)
					if err != nil {
						return fmt.Errorf("reading dir %q: %w", dir, err)
					}

					for _, file := range files {
						// If it's a directory, it will be grabbed by another worker.
						if file.IsDir() {
							continue
						}

						// Send the file
						select {
						case foundFilesCh <- filepath.Join(dir, file.Name()):
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

	// Router: detect if it's a build file (we read the file, so we should do it in parallel).
	buildFilesCh := make(chan *buildFileDescription)
	allFilesCh := make(chan string)
	{
		var wg sync.WaitGroup

		// Ensure the channel gets closed.
		g.Go(func() error {
			wg.Wait()
			close(buildFilesCh)
			close(allFilesCh)
			return nil
		})

		// Create a new worker to search for build files.
		for i := 0; i < kBuildFileSearcherWorkerCount; i++ {
			wg.Add(1)
			g.Go(func() error {
				defer wg.Done()

				for file := range foundFilesCh {
					// We send it to the all files channel.
					select {
					case allFilesCh <- file:
						// Sent.
					case <-ctx.Done():
						return ctx.Err()
					}

					// Now we check if it's an unreal file and send it to the specific channel.
					moduleName, ok, err := IsUnrealBuildFile(file)
					if err != nil {
						return fmt.Errorf("checking if %q is unreal file: %w", file, err)
					}

					bfd := &buildFileDescription{
						ModuleName: moduleName,
						Path:       file,
					}

					if ok {
						select {
						case buildFilesCh <- bfd:
							// Sent.
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}

				return nil
			})
		}
	}

	// Reduce: Collect all the files in an array to be sorted and also collects the module.
	var buildFiles []*buildFileDescription
	var allFiles []string
	{
		g.Go(func() error {
			for bfd := range buildFilesCh {
				buildFiles = append(buildFiles, bfd)
			}
			return nil
		})

		g.Go(func() error {
			for file := range allFilesCh {
				// Add it to the generic array.
				allFiles = append(allFiles, file)
			}
			return nil
		})
	}

	// We run the whole script.
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// We sort all the files so we can do faster searches on it.
	sort.Strings(allFiles)

	return &collectFilesResult{
		buildFiles: buildFiles,
		allFiles:   allFiles,
	}, nil
}
