package unreal_test

import (
	"context"
	"os"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/build"
	"github.com/cristiandonosoc/gochart/internal/option"
	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
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
			projectDir := tmp // Alias.

			version := build.BuildVersion()

			// First listing should be empty.
			{
				config := &obtainProjectGoutputsConfig{
					version: version,
					platform: unreal.Platform_Windows,
					strategy: strategy,
					projectDir: projectDir,
					generateMetadata: false,
				}

				got := listProject(t, projectDir)
				want := obtainProjectGoutputs(t, config)
				compareGoutputs(t, want, got)
			}
		})
	}
}

func listProject(t *testing.T, projectDir string) []*unreal_backend.GochartOutput {
	ctx := context.Background()
	ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, projectDir)
	require.NoError(t, err)

	options := []option.Option{
		unreal_backend.WithStrategy(unreal_backend.GenerationStrategy_Inline),
	}

	listing, err := ugb.List(ctx, options...)
	require.NoError(t, err)

	return listing
}

func compareGoutputs(t *testing.T, want, got []*unreal_backend.GochartOutput) {
	require.Equal(t, len(want), len(got))

	for i := 0; i < len(want); i++ {
		want := want[i]
		got := got[i]

		require.Equal(t, want.Gochart.Path, got.Gochart.Path)
		require.Equal(t, want.UpToDate(), got.UpToDate())

		require.Equal(t, len(want.Outputs), len(got.Outputs))
		for j := 0; j < len(want.Outputs); j++ {
			want := want.Outputs[j]
			got := got.Outputs[j]

			assert.Equal(t, want.Name, got.Name)
			assert.Equal(t, want.ExpectedPath, got.ExpectedPath)

			if assert.Equal(t, want.Metadata != nil, got.Metadata != nil) {
				if want.Metadata != nil {
					assert.Equal(t, want.Metadata.Version, got.Metadata.Version)
					assert.Equal(t, want.Metadata.IsUnrealGochart, got.Metadata.IsUnrealGochart)
					assert.Equal(t, want.Metadata.UnrealModule, got.Metadata.UnrealModule)
					assert.Equal(t, want.Metadata.GochartModulePath, got.Metadata.GochartModulePath)
				}
			}
		}
	}
}
