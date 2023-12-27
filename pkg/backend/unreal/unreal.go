// Package unreal is the backend generation that takes care of generating gochart files for an
// Unreal project.
package unreal

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/internal/unreal"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"

	"golang.org/x/sync/errgroup"
)

const (
	kModuleSearcherWorkerCount = 10
	kGochartBuilderWorkerCount = 16
)

type GenerationStrategy int

const (
	// Intermediate means generating files into the project's intermediate files.
	GenerationStrategy_Intermediate GenerationStrategy = iota

	// Inline means generating the files right next to the gochart file.
	GenerationStrategy_Inline

	// ModuleGeneratedDir creates a Generated directory in the module root and append the internal
	// route to it.
	GenerationStrategy_ModuleGeneratedDir
)

func (gs GenerationStrategy) String() string {
	switch gs {
	case GenerationStrategy_Intermediate:
		return "Intermediate"
	case GenerationStrategy_Inline:
		return "Inline"
	case GenerationStrategy_ModuleGeneratedDir:
		return "ModuleGeneratedDir"
	default:
		return "<unknown>"
	}
}

// Options -----------------------------------------------------------------------------------------

// TODO(cdc): Join options into a common backend options as well.
type Options struct {
	Version string
	Time    time.Time

	ProjectDir string
	Platform   unreal.UnrealPlatform
	Strategy   GenerationStrategy
}

func defaultOptions() *Options {
	options := &Options{
		Version:  "DEVELOPMENT",
		Time:     time.Now(),
		Platform: unreal.UnrealPlatform_Windows,
		Strategy: GenerationStrategy_ModuleGeneratedDir,
	}

	return options
}

func optWrapper(fn func(options *Options)) option.Option {
	return func(opts interface{}) {
		casted := opts.(*Options)
		fn(casted)
	}
}

func WithVersion(version string) option.Option {
	return optWrapper(func(options *Options) {
		options.Version = version
	})
}

func WithTime(t time.Time) option.Option {
	return optWrapper(func(options *Options) {
		options.Time = t
	})
}

func WithProjectDir(projectDir string) option.Option {
	return optWrapper(func(options *Options) {
		options.ProjectDir = projectDir
	})
}

func WithPlatform(platform unreal.UnrealPlatform) option.Option {
	return optWrapper(func(options *Options) {
		options.Platform = platform
	})
}

func WithStrategy(strategy GenerationStrategy) option.Option {
	return optWrapper(func(options *Options) {
		options.Strategy = strategy
	})
}

// Backend -----------------------------------------------------------------------------------------

type unrealGochartBackend struct {
	gochartFrontend frontend.GochartFrontend
	options         *Options
}

func NewUnrealGochartBackend(opts ...option.Option) (*unrealGochartBackend, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	gf, err := gochart_lang.NewGochartLangFrontend()
	if err != nil {
		return nil, fmt.Errorf("creating gochart lang frontend: %w", err)
	}

	return &unrealGochartBackend{
		gochartFrontend: gf,
		options:         options,
	}, nil
}

type gochartOutput struct {
	Gochart *unreal.UnrealFile
	Outputs []*common.BackendOutput
}

func (ugb *unrealGochartBackend) Generate(ctx context.Context, sc *ir.Statechart, opts ...option.Option) ([]*gochartOutput, error) {
	// Override options.
	for _, opt := range opts {
		opt(ugb.options)
	}

	if err := ugb.validate(); err != nil {
		return nil, fmt.Errorf("validating: %w", err)
	}

	if ugb.options.ProjectDir == "" {
		return nil, fmt.Errorf("missing project dir")
	}

	project, err := unreal.IndexUnrealProject(ctx, ugb.options.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("indexing unreal project: %w", err)
	}

	gocharts, err := ugb.searchGochartFiles(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("searching for gochart files: %w", err)
	}

	outputs, err := ugb.generateGocharts(ctx, project, gocharts)
	if err != nil {
		return nil, fmt.Errorf("generating gocharts: %w", err)
	}

	return outputs, nil
}

func (ugb *unrealGochartBackend) validate() error {
	if ok, err := files.DirExists(ugb.options.ProjectDir); err != nil {
		return fmt.Errorf("checking if dir %q exists: %w", ugb.options.ProjectDir, err)
	} else if !ok {
		return fmt.Errorf("dir %q does not exists (or is not a directory): %w", ugb.options.ProjectDir, err)
	}

	return nil
}

func (ugb *unrealGochartBackend) searchGochartFiles(ctx context.Context,
	project *unreal.UnrealProject) ([]string, error) {

	// We search the gocharts in a parallel fashion.
	g, ctx := errgroup.WithContext(ctx)

	// Produce: all the modules to check.
	modulesCh := make(chan *unreal.UnrealModule)
	{
		g.Go(func() error {
			defer close(modulesCh)

			for _, module := range project.Modules {
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
						if ext != ".gochart" {
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
