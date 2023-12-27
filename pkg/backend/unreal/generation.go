package unreal

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"github.com/cristiandonosoc/gochart/internal/unreal"

	"golang.org/x/sync/errgroup"
)

func (ugb *unrealGochartBackend) generateGocharts(ctx context.Context,
	project *unreal.UnrealProject, gocharts []string) ([]*gochartOutput, error) {

	// We generate the gocharts in a parallel fashion.
	g, ctx := errgroup.WithContext(ctx)

	// Produce: all the gocharts to generate.
	gochartsCh := make(chan string)
	{
		g.Go(func() error {
			defer close(gochartsCh)

			for _, gochart := range gocharts {
				select {
				case gochartsCh <- gochart:
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})
	}

	// Map: for each gochart, generate the backend files.
	outputsCh := make(chan *gochartOutput)
	{
		var wg sync.WaitGroup

		// Make sure the channel will be closed.
		g.Go(func() error {
			wg.Wait()
			close(outputsCh)
			return nil
		})

		for i := 0; i < kGochartBuilderWorkerCount; i++ {
			wg.Add(1)
			g.Go(func() error {
				defer wg.Done()

				for gochart := range gochartsCh {
					output, err := generateGochart(ctx, project, gochart)
					if err != nil {
						return fmt.Errorf("generating gochart %q: %w", gochart, err)
					}

					select {
					case outputsCh <- output:
						continue
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return nil
			})
		}
	}

	// Reduce: collect gochart outputs.
	var outputs []*gochartOutput
	{
		g.Go(func() error {
			for output := range outputsCh {
				outputs = append(outputs, output)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(outputs, func(i, j int) bool {
		return outputs[i].Gochart < outputs[j].Gochart
	})

	return outputs, nil
}

func generateGochart(ctx context.Context, project *unreal.UnrealProject, gochart string) (*gochartOutput, error) {
	return nil, fmt.Errorf("IMPLEMENT ME")
}

// gochartOutputPaths gets the output paths belonging to this particular gochart.
// Intermediate/Build/<PLATFORM>/UnrealEditor/Inc/Gochart/<MODULE>/<INTERNAL_PATH>/<NAME>.gochart.(h|cpp)
func gochartOutputPaths(project *unreal.UnrealProject, platform unreal.UnrealPlatform, strategy GenerationStrategy,
	gochart *unreal.UnrealFile) (_header string, _cpp string, _err error) {

	var base string
	switch strategy {
	case GenerationStrategy_Intermediate:
		// First it's built by the intermediate platform location.
		base = filepath.Join(project.ProjectDir, "Intermediate", "Build", string(platform), "UnrealEditor", "Inc", "Gochart")

		// Then we get to the per module place and internal path.
		base = filepath.Join(base, gochart.Module.Name, filepath.Dir(gochart.InternalPath()))
	case GenerationStrategy_Inline:
		base = filepath.Dir(gochart.Path)
	case GenerationStrategy_ModuleGeneratedDir:
		base = filepath.Join(gochart.Module.BaseDir, "Generated")
		base = filepath.Join(base, filepath.Dir(gochart.InternalPath()))
	default:
		return "", "", fmt.Errorf("unknown generation strategy: %s", strategy)
	}

	// Now we return the output paths for the C++ backend.
	gochartName := gochart.Name()
	ext := filepath.Ext(gochartName)
	name := gochartName[0 : len(gochartName)-len(ext)]

	header := filepath.Join(base, fmt.Sprintf("%s.gochart.generated.h", name))
	cpp := filepath.Join(base, fmt.Sprintf("%s.gochart.generated.cpp", name))

	return header, cpp, nil
}
