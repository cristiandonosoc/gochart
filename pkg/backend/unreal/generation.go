package unreal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/cristiandonosoc/gochart/internal/files"
	cpp_gen "github.com/cristiandonosoc/gochart/internal/generation/cpp"
	"github.com/cristiandonosoc/gochart/internal/unreal"
	"github.com/cristiandonosoc/gochart/pkg/backend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"

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
		return outputs[i].Gochart.Path < outputs[j].Gochart.Path
	})

	return outputs, nil
}

func (ugb *unrealGochartBackend) generateGochartIfNeeded(ctx context.Context,
	project *unreal.UnrealProject, gochartPath string) (*gochartOutput, error) {

	if err := checkIfContextIsValid(ctx); err != nil {
		return nil, fmt.Errorf("context is not longer valid: %w", err)
	}

	gochart, err := project.NewUnrealFile(gochartPath)
	if err != nil {
		return nil, fmt.Errorf("getting gochart unreal file: %w", err)
	}

	outputDir, err := calculateGochartOutputDir(project, ugb.options.Platform, ugb.options.Strategy, gochart)
	if err != nil {
		return nil, fmt.Errorf("calculating output directory: %w", err)
	}

	// Get where the files would be generated.
	cppGenConfig := &cpp_gen.Config{
		Version:           ugb.options.Version,
		Time:              ugb.options.Time,
		OutputDir:         outputDir,
		FilePrefix:        gochart.Name(),
		GenerateOutputDir: true,
	}
	expectedOutputs, err := cpp_gen.CalculateExpectedPaths(cppGenConfig)
	if err != nil {
		return nil, fmt.Errorf("calculating expected outputs: %w", err)
	}

	// Now we check if we need to generate the output files.
	// They get generated all or nothing.
	valid := true
	for _, output := range expectedOutputs {
		stat, err := files.StatFile(output.ExpectedPath)
		if err != nil {
			return nil, fmt.Errorf("stating %q: %w", output.ExpectedPath, err)
		}

		// If the file doens't exist, it is invalid.
		if stat == nil {
			valid = false
			break
		}

		// We check if the source gochart file is newer.
		// If so, the existing statechart files are invalid and need to be regenerated.
		if gochart.FileInfo.ModTime().After(stat.ModTime()) {
			valid = false
			break
		}
	}

	// If the existing files are valid, there is nothing to generat.
	// We simply generated the cached outputs.
	if valid {
		outputs := make([]*common.BackendOutput, 0, len(expectedOutputs))
		for _, expectedOutput := range expectedOutputs {
			output := common.NewBackendOutputFromCppOutput(expectedOutput)
			output.UpToDate = true
			outputs = append(outputs, output)
		}

		return &gochartOutput{
			Gochart: gochart,
			Outputs: outputs,
		}, nil
	}

	// Here we need to generate the files.
	outputs, err := ugb.generateGochart(ctx, gochart, cppGenConfig)
	if err != nil {
		return nil, fmt.Errorf("generting files for gochart: %w", err)
	}

	return &gochartOutput{
		Gochart: gochart,
		Outputs: outputs,
	}, nil
}

func (ugb *unrealGochartBackend) generateGochart(ctx context.Context,
	gochart *unreal.UnrealFile, cppGenConfig *cpp_gen.Config) ([]*common.BackendOutput, error) {

	if err := checkIfContextIsValid(ctx); err != nil {
		return nil, fmt.Errorf("context is no longer valid: %w", err)
	}

	lf, err := files.LoadFileFromPath(gochart.Path)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %w", gochart.Path, err)
	}

	scdata, err := ugb.gochartFrontend.Process(lf)
	if err != nil {
		return nil, fmt.Errorf("processing gochart frontend: %w", err)
	}

	sc, err := ir.ProcessStatechartData(scdata)
	if err != nil {
		return nil, fmt.Errorf("processing IR data: %w", err)
	}

	cppOutputs, err := cpp_gen.Generate(sc, cppGenConfig)
	if err != nil {
		return nil, fmt.Errorf("Generating gochart output files: %w", err)
	}

	// Now that we have the files generated, we check if we need to write them to disk.
	outputs := make([]*common.BackendOutput, 0, len(cppOutputs))
	for _, cppOutput := range cppOutputs {
		output := common.NewBackendOutputFromCppOutput(cppOutput)

		written, err := ugb.writeOutputIfDifferentOnDisk(ctx, output)
		if err != nil {
			return nil, fmt.Errorf("writing gochart output to disk: %w", err)
		}

		// If we didn't write, means that even though it was generated, the output was up to date.
		// This can happen when someone makes non-structural changes to a statechart (like adding
		// comments), which will invalidate the mod time, but not change the output.
		if !written {
			output.UpToDate = true
		}

		outputs = append(outputs, common.NewBackendOutputFromCppOutput(cppOutput))
	}

	return outputs, nil
}

// gochartOutputPaths gets the output paths belonging to this particular gochart.
// Intermediate/Build/<PLATFORM>/UnrealEditor/Inc/Gochart/<MODULE>/<INTERNAL_PATH>/<NAME>.gochart.(h|cpp)
func calculateGochartOutputDir(project *unreal.UnrealProject, platform unreal.UnrealPlatform, strategy GenerationStrategy,
	gochart *unreal.UnrealFile) (string, error) {

	var dir string
	switch strategy {
	case GenerationStrategy_Intermediate:
		// First it's built by the intermediate platform location.
		dir = filepath.Join(project.ProjectDir, "Intermediate", "Build", string(platform), "UnrealEditor", "Inc", "Gochart")

		// Then we get to the per module place and internal path.
		dir = filepath.Join(dir, gochart.Module.Name, filepath.Dir(gochart.InternalPath()))
	case GenerationStrategy_Inline:
		dir = filepath.Dir(gochart.Path)
	case GenerationStrategy_ModuleGeneratedDir:
		dir = filepath.Join(gochart.Module.BaseDir, "Generated")
		dir = filepath.Join(dir, filepath.Dir(gochart.InternalPath()))
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
		if !os.IsNotExist(err) {
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
