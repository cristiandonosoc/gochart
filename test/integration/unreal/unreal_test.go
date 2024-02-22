package unreal_generation_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/build"
	"github.com/cristiandonosoc/gochart/internal/option"
	backend_common "github.com/cristiandonosoc/gochart/pkg/backend/common"
	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/cristiandonosoc/golib/pkg/test_support"
	"github.com/cristiandonosoc/gunreal/pkg/unreal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	strategies := []unreal_backend.GenerationStrategy{
		unreal_backend.GenerationStrategy_Inline,
		//unreal_backend.GenerationStrategy_ModuleGeneratedDir,
		//unreal_backend.GenerationStrategy_UHT,
	}

	for _, strategy := range strategies {
		t.Run(strategy.String(), func(t *testing.T) {
			// Copy the test project just for this test.
			tmpBase := test_support.TestTmpBase()
			tmp, err := os.MkdirTemp(tmpBase, "unreal-test-list")
			require.NoError(t, err)
			defer os.RemoveAll(tmp)

			copyProject(t, tmp)

			ctx := context.Background()
			ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, tmp)
			require.NoError(t, err)

			options := []option.Option{
				unreal_backend.WithStrategy(unreal_backend.GenerationStrategy_Inline),
			}

			got, err := ugb.List(ctx, options...)
			require.NoError(t, err)

			want := obtainProjectGoutputs(t, strategy, tmp)
			compareGoutputs(t, want, got)
		})
	}

}

// Helper functions --------------------------------------------------------------------------------

func copyProject(t *testing.T, dir string) {
	uprojectPath := filepath.Join("testdata", "GochartTest", "GochartTest.uproject")
	uproject, err := test_support.RunfilePath(uprojectPath)
	require.NoError(t, err)

	projectDir := filepath.Dir(uproject)

	err = files.CopyDirRecursive(projectDir, dir)
	require.NoError(t, err)
}

type testBuildDescription struct {
	Platform   unreal.Platform
	ProjectDir string
	Strategy   unreal_backend.GenerationStrategy
}

type testGochartDescription struct {
	Path       string
	FilePrefix string
	ModuleName string
	ModulePath string
}

func calculateStrategyBaseDir(t *testing.T, build *testBuildDescription, gochart *testGochartDescription) string {
	switch build.Strategy {
	case unreal_backend.GenerationStrategy_Inline:
		return filepath.Dir(gochart.Path)
	case unreal_backend.GenerationStrategy_ModuleGeneratedDir:
		return filepath.Join(build.ProjectDir, "Generated", filepath.Dir(gochart.Path))
	case unreal_backend.GenerationStrategy_UHT:
		base := filepath.Join(build.ProjectDir, "Intermediate", "Build", build.Platform.String(), "UnrealEditor", "Inc")
		return filepath.Join(base, gochart.ModuleName, "UHT", filepath.Dir(gochart.ModulePath))
	}

	t.Fatalf("unknown strategy %q", build.Strategy)
	return ""
}

func obtainProjectGoutputs(t *testing.T, strategy unreal_backend.GenerationStrategy, projectDir string) []*unreal_backend.GochartOutput {
	version := build.BuildVersion()

	build := &testBuildDescription{
		Platform:   unreal.Platform_Windows,
		ProjectDir: projectDir,
		Strategy:   strategy,
	}

	gocharts := []*testGochartDescription{
		{
			Path:       filepath.Join(projectDir, "Source", "GochartTest", "AnotherDir", "Counter.gochart"),
			FilePrefix: "Counter",
			ModuleName: "GochartTest",
			ModulePath: filepath.Join("AnotherDir", "Counter.gochart"),
		},
		{
			Path:       filepath.Join(projectDir, "Source", "GochartTest", "GochartTestActor.gochart"),
			FilePrefix: "GochartTestActor",
			ModuleName: "GochartTest",
			ModulePath: "GochartTestActor.gochart",
		},
	}

	generateBackendOutputs := func(t *testing.T, build *testBuildDescription, gochart *testGochartDescription) []*backend_common.BackendOutput {
		baseDir := calculateStrategyBaseDir(t, build, gochart)

		metadata := &backend_common.Metadata{
			Version:           version,
			IsUnrealGochart:   true,
			UnrealModule:      gochart.ModuleName,
			GochartModulePath: files.ToUnixPath(gochart.ModulePath),
		}

		return []*backend_common.BackendOutput{
			{
				Name:         "forward_declare",
				ExpectedPath: filepath.Join(baseDir, fmt.Sprintf("%s.gochart.forward_declare.h", gochart.FilePrefix)),
				Metadata:     metadata,
			},
			{
				Name:         "declare",
				ExpectedPath: filepath.Join(baseDir, fmt.Sprintf("%s.gochart.declare.h", gochart.FilePrefix)),
				Metadata:     metadata,
			},
			{
				Name:         "define",
				ExpectedPath: filepath.Join(baseDir, fmt.Sprintf("%s.gochart.define.h", gochart.FilePrefix)),
				Metadata:     metadata,
			},
		}
	}

	goutputs := make([]*unreal_backend.GochartOutput, 0, len(gocharts))

	for _, gochart := range gocharts {
		goutput := &unreal_backend.GochartOutput{
			Gochart: &unreal.File{
				Path: gochart.Path,
			},
		}
		goutput.Outputs = generateBackendOutputs(t, build, gochart)

		goutputs = append(goutputs, goutput)
	}

	return goutputs
}

func compareGoutputs(t *testing.T, want, got []*unreal_backend.GochartOutput) {
	require.Equal(t, len(want), len(got))

	for i := 0; i < len(want); i++ {
		want := want[i]
		got := got[i]

		require.Equal(t, want.Gochart.Path, got.Gochart.Path)

		require.Equal(t, len(want.Outputs), len(got.Outputs))

		for j := 0; j < len(want.Outputs); j++ {
			want := want.Outputs[j]
			got := got.Outputs[j]

			assert.Equal(t, want.Name, got.Name)
			assert.Equal(t, want.ExpectedPath, got.ExpectedPath)

			if assert.NotNil(t, got.Metadata) {
				assert.Equal(t, want.Metadata.Version, got.Metadata.Version)
				assert.Equal(t, want.Metadata.IsUnrealGochart, got.Metadata.IsUnrealGochart)
				assert.Equal(t, want.Metadata.UnrealModule, got.Metadata.UnrealModule)
				assert.Equal(t, want.Metadata.GochartModulePath, got.Metadata.GochartModulePath)
			}
		}
	}
}
