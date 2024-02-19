package unreal_generation_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/build"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/internal/test_support"
	backend_common "github.com/cristiandonosoc/gochart/pkg/backend/common"
	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/cristiandonosoc/gunreal/pkg/unreal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	tmpBase := test_support.TestTmpBase()
	tmp, err := os.MkdirTemp(tmpBase, "unreal-test-list")
	require.NoError(t, err)
	defer os.RemoveAll(tmp)

	copyProject(t, tmp)

	wantVersion := build.BuildVersion()

	want := []*unreal_backend.GochartOutput{
		{
			Gochart: &unreal.File{
				Path: filepath.Join(tmp, "Source", "GochartTest", "AnotherDir", "Counter.gochart"),
			},
			Outputs: []*backend_common.BackendOutput{
				{
					Name:         "forward_declare",
					ExpectedPath: filepath.Join(tmp, "Source", "GochartTest", "AnotherDir", "Counter.gochart.forward_declare.h"),
					Metadata: &backend_common.Metadata{
						Version:           wantVersion,
						IsUnrealGochart:   true,
						UnrealModule:      "GochartTest",
						GochartModulePath: files.ToUnixPath(filepath.Join("AnotherDir", "Counter.gochart")),
					},
				},
				{
					Name:         "declare",
					ExpectedPath: filepath.Join(tmp, "Source", "GochartTest", "AnotherDir", "Counter.gochart.declare.h"),
					Metadata: &backend_common.Metadata{
						Version:           wantVersion,
						IsUnrealGochart:   true,
						UnrealModule:      "GochartTest",
						GochartModulePath: files.ToUnixPath(filepath.Join("AnotherDir", "Counter.gochart")),
					},
				},
				{
					Name:         "define",
					ExpectedPath: filepath.Join(tmp, "Source", "GochartTest", "AnotherDir", "Counter.gochart.define.h"),
					Metadata: &backend_common.Metadata{
						Version:           wantVersion,
						IsUnrealGochart:   true,
						UnrealModule:      "GochartTest",
						GochartModulePath: files.ToUnixPath(filepath.Join("AnotherDir", "Counter.gochart")),
					},
				},
			},
		},

		{
			Gochart: &unreal.File{
				Path: filepath.Join(tmp, "Source", "GochartTest", "GochartTestActor.gochart"),
			},
			Outputs: []*backend_common.BackendOutput{
				{
					Name:         "forward_declare",
					ExpectedPath: filepath.Join(tmp, "Source", "GochartTest", "GochartTestActor.gochart.forward_declare.h"),
					Metadata: &backend_common.Metadata{
						Version:           wantVersion,
						IsUnrealGochart:   true,
						UnrealModule:      "GochartTest",
						GochartModulePath: files.ToUnixPath(filepath.Join("GochartTestActor.gochart")),
					},
				},
				{
					Name:         "declare",
					ExpectedPath: filepath.Join(tmp, "Source", "GochartTest", "GochartTestActor.gochart.declare.h"),
					Metadata: &backend_common.Metadata{
						Version:           wantVersion,
						IsUnrealGochart:   true,
						UnrealModule:      "GochartTest",
						GochartModulePath: files.ToUnixPath(filepath.Join("GochartTestActor.gochart")),
					},
				},
				{
					Name:         "define",
					ExpectedPath: filepath.Join(tmp, "Source", "GochartTest", "GochartTestActor.gochart.define.h"),
					Metadata: &backend_common.Metadata{
						Version:           wantVersion,
						IsUnrealGochart:   true,
						UnrealModule:      "GochartTest",
						GochartModulePath: files.ToUnixPath(filepath.Join("GochartTestActor.gochart")),
					},
				},
			},
		},
	}

	ctx := context.Background()
	ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, tmp)
	require.NoError(t, err)

	options := []option.Option{
		unreal_backend.WithStrategy(unreal_backend.GenerationStrategy_Inline),
	}

	got, err := ugb.List(ctx, options...)
	require.NoError(t, err)

	compareGoutputs(t, want, got)
}

// Helper functions --------------------------------------------------------------------------------

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

func copyProject(t *testing.T, dir string) {
	uprojectPath := filepath.Join("testdata", "GochartTest", "GochartTest.uproject")
	uproject, err := test_support.RunfilePath(uprojectPath)
	require.NoError(t, err)

	projectDir := filepath.Dir(uproject)

	err = files.CopyDirRecursive(projectDir, dir)
	require.NoError(t, err)
}
