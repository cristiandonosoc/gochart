package unreal

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	backend_common "github.com/cristiandonosoc/gochart/pkg/backend/common"
	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/cristiandonosoc/golib/pkg/test_support"
	gunreal "github.com/cristiandonosoc/gunreal/pkg/unreal"

	"github.com/stretchr/testify/require"
)

func copyProject(t *testing.T, dir string) {
	uprojectPath := filepath.Join("testdata", "GochartTest", "GochartTest.uproject")
	uproject, err := test_support.RunfilePath(uprojectPath)
	require.NoError(t, err)

	projectDir := filepath.Dir(uproject)

	err = files.CopyDirRecursive(projectDir, dir)
	require.NoError(t, err)
}

type testBuildDescription struct {
	Platform   gunreal.Platform
	Version    string
	ProjectDir string
	Strategy   unreal_backend.GenerationStrategy
}

type testGochartDescription struct {
	Path         string
	FilePrefix   string
	ModuleName   string
	InModulePath string
}

func calculateStrategyBaseDir(t *testing.T, bd *testBuildDescription, gd *testGochartDescription) string {
	switch bd.Strategy {
	case unreal_backend.GenerationStrategy_Inline:
		return filepath.Dir(gd.Path)
	case unreal_backend.GenerationStrategy_ModuleGeneratedDir:
		return filepath.Join(bd.ProjectDir, "Source", gd.ModuleName, "Generated", filepath.Dir(gd.InModulePath))
	case unreal_backend.GenerationStrategy_UHT:
		base := filepath.Join(bd.ProjectDir, "Intermediate", "Build", bd.Platform.String(), "UnrealEditor", "Inc")
		return filepath.Join(base, gd.ModuleName, "UHT", filepath.Dir(gd.InModulePath))
	}

	t.Fatalf("unknown strategy %q", bd.Strategy)
	return ""
}

func obtainProjectGoutputs(t *testing.T, config *testConfig) []*unreal_backend.GochartOutput {
	bd := &testBuildDescription{
		Platform:   config.platform,
		Version:    config.version,
		ProjectDir: config.projectDir,
		Strategy:   config.strategy,
	}

	gds := []*testGochartDescription{
		{
			Path:         filepath.Join(config.projectDir, "Source", "GochartTest", "AnotherDir", "Counter.gochart"),
			FilePrefix:   "Counter",
			ModuleName:   "GochartTest",
			InModulePath: filepath.Join("AnotherDir", "Counter.gochart"),
		},
		{
			Path:         filepath.Join(config.projectDir, "Source", "GochartTest", "GochartTestActor.gochart"),
			FilePrefix:   "GochartTestActor",
			ModuleName:   "GochartTest",
			InModulePath: "GochartTestActor.gochart",
		},
	}

	goutputs := make([]*unreal_backend.GochartOutput, 0, len(gds))

	for _, gd := range gds {
		// See if there are filters.
		if len(config.generatePositiveFilters) > 0 {
			// Only generate that match all of them.
			nonMatch := false
			for _, filter := range config.generatePositiveFilters {
				if !strings.Contains(gd.Path, filter) {
					nonMatch = true
					break
				}
			}

			if nonMatch {
				continue
			}
		}

		if len(config.generateNegativeFilters) > 0 {
			// Only generate that don't match any.
			match := false
			for _, filter := range config.generateNegativeFilters {
				if strings.Contains(gd.Path, filter) {
					match = true
					break
				}
			}

			if match {
				continue
			}
		}

		goutput := &unreal_backend.GochartOutput{
			Gochart: &gunreal.File{
				Path: gd.Path,
			},
		}
		goutput.Outputs = generateBackendOutputs(t, config, bd, gd)

		goutputs = append(goutputs, goutput)
	}

	sort.Slice(goutputs, func(i, j int) bool {
		return goutputs[i].Gochart.Path < goutputs[j].Gochart.Path
	})

	return goutputs
}

func generateBackendOutputs(t *testing.T, config *testConfig, bd *testBuildDescription, gd *testGochartDescription) []*backend_common.BackendOutput {
	baseDir := calculateStrategyBaseDir(t, bd, gd)

	var metadata *backend_common.Metadata
	if config.generateMetadata {
		metadata = &backend_common.Metadata{
			Version:           config.version,
			IsUnrealGochart:   true,
			UnrealModule:      gd.ModuleName,
			GochartModulePath: files.ToUnixPath(gd.InModulePath),
		}
	}

	outputs := []*backend_common.BackendOutput{
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

	sort.Slice(outputs, func(i, j int) bool {
		return outputs[i].Name < outputs[j].Name
	})

	return outputs
}
