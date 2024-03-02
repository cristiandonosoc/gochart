package unreal

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
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
	generatePositiveFilters []string // Only generate test files don't these words on them.
	generateNegativeFilters []string // Only generate test files don't these words on them.
	generateMetadata        bool
}

func TestList(t *testing.T) {
	strategies := []unreal_backend.GenerationStrategy{
		unreal_backend.GenerationStrategy_Inline,
		unreal_backend.GenerationStrategy_ModuleGeneratedDir,
		unreal_backend.GenerationStrategy_UHT,
	}

	fmt.Println(os.Args[0])

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
				compareGoutputs(t, want, got.Gocharts)

				// Should not have any orphaned gocharts.
				require.Empty(t, got.OrphanedGocharts)
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

			// Re-listing should give the existing files.
			var counterGochart *gunreal.File
			{
				config := &testConfig{
					version:          version,
					platform:         gunreal.Platform_Windows,
					strategy:         strategy,
					projectDir:       projectDir,
					generateMetadata: true,
				}

				got := listProject(t, config)

				want := obtainProjectGoutputs(t, config)
				compareGoutputs(t, want, got.Gocharts)

				// Should not have any orphaned gocharts.
				require.Empty(t, got.OrphanedGocharts)

				// Get the first gochart so it can be eliminated.
				for _, goutput := range got.Gocharts {
					if strings.Contains(goutput.Gochart.Path, "Counter") {
						counterGochart = goutput.Gochart
						break
					}
				}
			}

			// Delete the gochart file.
			{
				require.NotNil(t, counterGochart)

				err = os.Remove(counterGochart.Path)
				require.NoError(t, err)
			}

			// Listing now should give one found and one orphaned.
			var wantDeleted []string
			{
				config := &testConfig{
					version:                 version,
					platform:                gunreal.Platform_Windows,
					strategy:                strategy,
					projectDir:              projectDir,
					generateMetadata:        true,
					generateNegativeFilters: []string{"Counter"},
				}

				got := listProject(t, config)

				want := obtainProjectGoutputs(t, config)
				compareGoutputs(t, want, got.Gocharts)

				// Compare the orphaned.
				config.generateNegativeFilters = nil
				config.generatePositiveFilters = []string{"Counter"}

				wantOrphaned := obtainProjectGoutputs(t, config)
				compareGoutputs(t, wantOrphaned, got.OrphanedGocharts)

				for _, orphaned := range wantOrphaned {
					for _, output := range orphaned.Outputs {
						wantDeleted = append(wantDeleted, output.ExpectedPath)
					}
				}
				sort.Strings(wantDeleted)
			}

			// Clean the orphaned.
			{
				config := &testConfig{
					version:                 version,
					platform:                gunreal.Platform_Windows,
					strategy:                strategy,
					projectDir:              projectDir,
					generateMetadata:        true,
					generateNegativeFilters: []string{"Counter"},
				}

				got := cleanOrphaned(t, config, "")
				require.Equal(t, wantDeleted, got)
			}

			// List should not have any orphaned.
			{
				config := &testConfig{
					version:                 version,
					platform:                gunreal.Platform_Windows,
					strategy:                strategy,
					projectDir:              projectDir,
					generateMetadata:        true,
					generateNegativeFilters: []string{"Counter"},
				}

				got := listProject(t, config)

				want := obtainProjectGoutputs(t, config)
				compareGoutputs(t, want, got.Gocharts)

				require.Empty(t, got.OrphanedGocharts)
			}
		})
	}
}

func listProject(t *testing.T, config *testConfig) *unreal_backend.ListResult {
	ctx := context.Background()

	ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, config.projectDir)
	require.NoError(t, err)

	options := []option.Option{
		unreal_backend.WithStrategy(config.strategy),
	}

	result, err := ugb.List(ctx, options...)
	require.NoError(t, err)

	return result
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

func cleanOrphaned(t *testing.T, config *testConfig, filter string) []string {
	ctx := context.Background()

	ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, config.projectDir)
	require.NoError(t, err)

	options := []option.Option{
		unreal_backend.WithStrategy(config.strategy),
	}

	deleted, err := ugb.CleanOrphaned(ctx, filter, options...)
	require.NoError(t, err)

	return deleted
}

func compareGoutputs(t *testing.T, want, got []*unreal_backend.GochartOutput) {
	require.Equal(t, len(want), len(got))

	for i := 0; i < len(want); i++ {
		want := want[i]
		got := got[i]

		require.Equal(t, want.Gochart.Path, got.Gochart.Path)

		path := want.Gochart.Path
		require.Equal(t, len(want.Outputs), len(got.Outputs))
		for j := 0; j < len(want.Outputs); j++ {
			want := want.Outputs[j]
			got := got.Outputs[j]

			require.Equal(t, want.Name, got.Name, path)
			require.Equal(t, want.ExpectedPath, got.ExpectedPath)

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
