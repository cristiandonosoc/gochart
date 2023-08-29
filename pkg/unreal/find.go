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

func findAllGochartFiles(ctx context.Context, sourceDir, extension string) ([]string, error) {
	extension = strings.ToLower(extension)

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

						if filepath.Ext(file.Name()) != extension {
							continue
						}

						select {
						case foundGochartCh <- filepath.Join(dir, file.Name()):
							continue
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}

				return nil
			})
		}

		// Close the channel.
		g.Go(func() error {
			wg.Wait()
			close(foundGochartCh)
			return nil
		})
	}

	// Reduce: all the found directories into a list.
	var gocharts []string
	{
		g.Go(func() error {
			for gochart := range foundGochartCh {
				gocharts = append(gocharts, gochart)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Strings(gocharts)

	return gocharts, nil
}
