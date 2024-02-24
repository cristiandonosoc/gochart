package unreal

import (
	"context"
	"os"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/build"
	"github.com/cristiandonosoc/gochart/internal/option"
	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/golib/pkg/test_support"
	gunreal "github.com/cristiandonosoc/gunreal/pkg/unreal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	version    string
	platform   gunreal.Platform
	strategy   unreal_backend.GenerationStrategy
	projectDir string

	// generation options.
	generateMetadata bool
}

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
				config := &testConfig{
					version:          version,
					platform:         gunreal.Platform_Windows,
					strategy:         strategy,
					projectDir:       projectDir,
					generateMetadata: false,
				}

				got := listProject(t, config)

				want := obtainProjectGoutputs(t, config)
				compareGoutputs(t, want, got)
			}

			// Generate should work.
			{
				config := &testConfig{
					version:          version,
					platform:         gunreal.Platform_Windows,
					strategy:         strategy,
					projectDir:       projectDir,
					generateMetadata: true,
				}

				got := generateProject(t, config)

				want := obtainProjectGoutputs(t, config)
				compareGoutputs(t, want, got)
			}
		})
	}
}

func listProject(t *testing.T, config *testConfig) []*unreal_backend.GochartOutput {
	ctx := context.Background()

	ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, config.projectDir)
	require.NoError(t, err)

	options := []option.Option{
		unreal_backend.WithStrategy(config.strategy),
	}

	goutputs, err := ugb.List(ctx, options...)
	require.NoError(t, err)

	return goutputs
}

func generateProject(t *testing.T, config *testConfig) []*unreal_backend.GochartOutput {
	ctx := context.Background()

	ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, config.projectDir)
	require.NoError(t, err)

	options := []option.Option{
		unreal_backend.WithStrategy(config.strategy),
	}

	goutputs, err := ugb.Generate(ctx, options...)
	require.NoError(t, err)

	return goutputs
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

			if assert.Equal(t, want.Metadata != nil, got.Metadata != nil, "GOTMETADATA:", got.Metadata) {
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
