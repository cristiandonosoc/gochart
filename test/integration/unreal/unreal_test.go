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
	Version    string
	ProjectDir string
	Strategy   unreal_backend.GenerationStrategy
}

type testGochartDescription struct {
	Path       string
	FilePrefix string
	ModuleName string
	ModulePath string
}

func calculateStrategyBaseDir(t *testing.T, bd *testBuildDescription, gd *testGochartDescription) string {
	switch bd.Strategy {
	case unreal_backend.GenerationStrategy_Inline:
		return filepath.Dir(gd.Path)
	case unreal_backend.GenerationStrategy_ModuleGeneratedDir:
		return filepath.Join(bd.ProjectDir, "Generated", filepath.Dir(gd.Path))
	case unreal_backend.GenerationStrategy_UHT:
		base := filepath.Join(bd.ProjectDir, "Intermediate", "Build", bd.Platform.String(), "UnrealEditor", "Inc")
		return filepath.Join(base, gd.ModuleName, "UHT", filepath.Dir(gd.ModulePath))
	}

	t.Fatalf("unknown strategy %q", bd.Strategy)
	return ""
}

func generateBackendOutputs(t *testing.T, bd *testBuildDescription, gd *testGochartDescription) []*backend_common.BackendOutput {
	version := build.BuildVersion()
	baseDir := calculateStrategyBaseDir(t, bd, gd)

	metadata := &backend_common.Metadata{
		Version:           version,
		IsUnrealGochart:   true,
		UnrealModule:      gd.ModuleName,
		GochartModulePath: files.ToUnixPath(gd.ModulePath),
	}

	return []*backend_common.BackendOutput{
		{
			Name:         "forward_declare",
			ExpectedPath: filepath.Join(baseDir, fmt.Sprintf("%s.gochart.forward_declare.h", gd.FilePrefix)),
			Metadata:     metadata,
		},
		{
			Name:         "declare",
			ExpectedPath: filepath.Join(baseDir, fmt.Sprintf("%s.gochart.declare.h", gd.FilePrefix)),
			Metadata:     metadata,
		},
		{
			Name:         "define",
			ExpectedPath: filepath.Join(baseDir, fmt.Sprintf("%s.gochart.define.h", gd.FilePrefix)),
			Metadata:     metadata,
		},
	}
}

func obtainProjectGoutputs(t *testing.T, strategy unreal_backend.GenerationStrategy, projectDir string) []*unreal_backend.GochartOutput {
	bd := &testBuildDescription{
		Platform:   unreal.Platform_Windows,
		Version:    build.BuildVersion(),
		ProjectDir: projectDir,
		Strategy:   strategy,
	}

	gds := []*testGochartDescription{
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

	goutputs := make([]*unreal_backend.GochartOutput, 0, len(gds))

	for _, gd := range gds {
		goutput := &unreal_backend.GochartOutput{
			Gochart: &unreal.File{
				Path: gd.Path,
			},
		}
		goutput.Outputs = generateBackendOutputs(t, bd, gd)

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
