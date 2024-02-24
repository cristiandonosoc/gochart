package unreal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/cristiandonosoc/gochart/internal/build"
	cpp_gen "github.com/cristiandonosoc/gochart/internal/generation/cpp"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/cristiandonosoc/gunreal/pkg/unreal"

	"golang.org/x/sync/errgroup"
)

func (ugb *unrealGochartBackend) generateGocharts(ctx context.Context,
	project *unreal.Project, gocharts []string) ([]*GochartOutput, error) {

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
	outputsCh := make(chan *GochartOutput)
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
					output, err := ugb.generateGochartIfNeeded(ctx, project, gochart)
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
	var outputs []*GochartOutput
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
		return outputs[i].Gochart.Path < outputs[j].Gochart.Path
	})

	return outputs, nil
}

func (ugb *unrealGochartBackend) generateGochartIfNeeded(ctx context.Context,
	project *unreal.Project, gochartPath string) (*GochartOutput, error) {

	if err := checkIfContextIsValid(ctx); err != nil {
		return nil, fmt.Errorf("context is not longer valid: %w", err)
	}

	expectedGoutput, cppGenConfig, err := ugb.calculateGochartExpectedOutputs(gochartPath)
	if err != nil {
		return nil, fmt.Errorf("calculating gochart expected output paths: %w", err)
	}

	// If the existing files are valid, there is nothing to generate.
	// We simply generated the cached outputs.
	if expectedGoutput.UpToDate() {
		return expectedGoutput, nil
	}

	outputs, err := ugb.generateGochart(ctx, expectedGoutput.Gochart, cppGenConfig)
	if err != nil {
		return nil, fmt.Errorf("generting files for gochart: %w", err)
	}

	return &GochartOutput{
		Gochart: expectedGoutput.Gochart,
		Outputs: outputs,
	}, nil
}

// calculateGochartExpectedOutputs returns the expected gochart output as if it were to be
// generated. This also returns the cpp_gen.Config used, so it the same can be used if the indeed
// we want to generate the files.
func (ugb *unrealGochartBackend) calculateGochartExpectedOutputs(gochartPath string) (*GochartOutput, *cpp_gen.Config, error) {
	gochart, err := ugb.Project.NewFile(gochartPath)
	if err != nil {
		return nil, nil, fmt.Errorf("getting gochart unreal file: %w", err)
	}

	outputDir, err := calculateGochartOutputDir(ugb.Project, ugb.options.Platform, ugb.options.Strategy, gochart)
	if err != nil {
		return nil, nil, fmt.Errorf("calculating output directory: %w", err)
	}

	// Get where the files would be generated.

	metadata := &common.Metadata{
		Version:           ugb.options.Version,
		IsUnrealGochart:   true,
		UnrealModule:      gochart.Module.Name,
		GochartModulePath: gochart.ModulePath(),
	}

	cppGenConfig := &cpp_gen.Config{
		Metadata:          metadata,
		Time:              ugb.options.Time,
		OutputDir:         outputDir,
		FilePrefix:        gochart.Name(),
		GenerateOutputDir: true,
	}

	if err := cppGenConfig.Metadata.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validating config: %w", err)
	}

	expectedOutputs, err := cpp_gen.CalculateExpectedPaths(cppGenConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("calculating expected outputs: %w", err)
	}

	for _, expectedOutput := range expectedOutputs {
		// Attempt to load the files into the cache.
		if _, err := files.LoadFileFromPath(expectedOutput.ExpectedPath); err != nil {
			// If the file doesn't exist, it gochart headers will need to be generated.
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return nil, nil, fmt.Errorf("loading file %q: %w", expectedOutput.ExpectedPath, err)
		}

		// Since the file is in disk, we check if it was made by the same version.
		metadata, err := cpp_gen.ReadMetadataFromBackendOutput(expectedOutput)
		if err != nil {
			return nil, nil, fmt.Errorf("loading metadata for %q: %w", expectedOutput.ExpectedPath, err)
		}
		expectedOutput.Metadata = metadata

		// If they were not made by the same version, we know it must be regenerated.
		if expectedOutput.Metadata.Version != build.BuildVersion() {
			continue
		}

		expectedOutput.Metadata.IsUnrealGochart = true
	}

	goutput := &GochartOutput{
		Gochart: gochart,
		Outputs: expectedOutputs,
	}

	return goutput, cppGenConfig, nil
}

func (ugb *unrealGochartBackend) generateGochart(ctx context.Context,
	gochart *unreal.File, cppGenConfig *cpp_gen.Config) ([]*common.BackendOutput, error) {

	if err := checkIfContextIsValid(ctx); err != nil {
		return nil, fmt.Errorf("context is no longer valid: %w", err)
	}

	lf, err := files.LoadFileFromPath(gochart.Path)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %w", gochart.Path, err)
	}

	gcdata, err := ugb.gochartFrontend.Process(lf)
	if err != nil {
		return nil, fmt.Errorf("processing gochart frontend: %w", err)
	}

	gc, err := ir.ProcessGochartData(gcdata)
	if err != nil {
		return nil, fmt.Errorf("processing IR data: %w", err)
	}

	outputs, err := cpp_gen.Generate(gc, cppGenConfig)
	if err != nil {
		return nil, fmt.Errorf("Generating gochart output files: %w", err)
	}

	// Now that we have the files generated, we check if we need to write them to disk.
	for _, output := range outputs {
		_, err := ugb.writeOutputIfDifferentOnDisk(ctx, output)
		if err != nil {
			return nil, fmt.Errorf("writing gochart output to disk: %w", err)
		}
	}

	return outputs, nil
}

// gochartOutputPaths gets the output paths belonging to this particular gochart.
// Intermediate/Build/<PLATFORM>/UnrealEditor/Inc/Gochart/<MODULE>/<INTERNAL_PATH>/<NAME>.gochart.(h|cpp)
func calculateGochartOutputDir(project *unreal.Project, platform unreal.Platform, strategy GenerationStrategy,
	gochart *unreal.File) (string, error) {

	var dir string
	switch strategy {
	case GenerationStrategy_UHT:
		// First it's built by the intermediate platform location.
		dir = filepath.Join(project.ProjectDir, "Intermediate", "Build", string(platform), "UnrealEditor", "Inc")

		// Then we get to the per module place and internal path.
		dir = filepath.Join(dir, gochart.Module.Name, "UHT", filepath.Dir(gochart.ModulePath()))
	case GenerationStrategy_Inline:
		dir = filepath.Dir(gochart.Path)
	case GenerationStrategy_ModuleGeneratedDir:
		dir = filepath.Join(gochart.Module.BaseDir, "Generated")
		dir = filepath.Join(dir, filepath.Dir(gochart.ModulePath()))
	default:
		return "", fmt.Errorf("unknown generation strategy: %s", strategy)
	}

	return dir, nil
}

// writeOutputIfDifferentOnDisk will check the contents of the file on disk and compare them to the
// generated data. If there was no error, the returned boolean states whether the file was written
// to disk or not. Returning false means that the contents on disk are the same, so there was no
// need to write it.
func (ugb *unrealGochartBackend) writeOutputIfDifferentOnDisk(ctx context.Context, output *common.BackendOutput) (bool, error) {
	if err := checkIfContextIsValid(ctx); err != nil {
		return false, fmt.Errorf("context is no longer valid: %w", err)
	}

	if output.Content == nil {
		return false, fmt.Errorf("no generated content found")
	}

	// Read the contents of the generated.
	generatedData, err := io.ReadAll(output.Content)
	if err != nil {
		return false, fmt.Errorf("reading generated data: %w", err)
	}

	fileData, err := os.ReadFile(output.ExpectedPath)
	if err != nil {
		// We ignore file not exists error, because we will need to write it anyway.
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("reading file %q: %w", output.ExpectedPath, err)
		}
	}

	// If the file contents are equal, we omit generating, stating that we did not.
	if bytes.Equal(fileData, generatedData) {
		return false, nil
	}

	// If we're here, it means we need to write to the file.

	// Ensure the path exists.
	if err := os.MkdirAll(filepath.Dir(output.ExpectedPath), os.ModeDir); err != nil {
		return false, fmt.Errorf("creating base directories %q: %w", filepath.Dir(output.ExpectedPath), err)
	}

	// First we create or truncate the file.
	file, err := os.Create(output.ExpectedPath)
	if err != nil {
		return false, fmt.Errorf("creating/truncating %q: %w", output.ExpectedPath, err)
	}
	defer file.Close()

	if _, err := file.Write(generatedData); err != nil {
		return false, fmt.Errorf("writing to %q: %w", output.ExpectedPath, err)
	}

	return true, nil
}

func checkIfContextIsValid(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
