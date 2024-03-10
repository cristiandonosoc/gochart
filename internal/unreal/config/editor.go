package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cristiandonosoc/golib/pkg/files"

	goversion "github.com/hashicorp/go-version"
)

type GunrealEditorConfig struct {
	Version *goversion.Version

	// Installed determines whether this is considered an installed version (eg. installing via
	// Unreal Launcher) vs a source build.
	Installed bool

	// For internal tracking information mostly.
	BuildVersionFile *buildVersionJson
}

type buildVersionJson struct {
	MajorVersion         int    `json:"MajorVersion"`
	MinorVersion         int    `json:"MinorVersion"`
	PatchVersion         int    `json:"PatchVersion"`
	Changelist           int    `json:"Changelist"`
	CompatibleChangelist int    `json:"CompatibleChangelist"`
	IsLicenseeVersion    int    `json:"IsLicenseeVersion"`
	IsPromotedbuild      int    `json:"IsPromotedbuild"`
	BranchName           string `json:"BranchName"`
}

func (gec *GunrealEditorConfig) Describe() string {
	var sb strings.Builder

	sb.WriteString("EDITOR -------------------------------------------------------------------\n\n")
	sb.WriteString(fmt.Sprintf("- VERSION: %s\n", gec.Version))
	sb.WriteString(fmt.Sprintf("- INSTALLED: %t\n", gec.Installed))

	return sb.String()
}

func newEditorConfig(editorDir string) (*GunrealEditorConfig, error) {
	// Sanity check that there is an Engine directory.
	if exists, err := files.DirExists(filepath.Join(editorDir, "Engine")); err != nil {
		return nil, fmt.Errorf("checking if engine dir exists: %w", err)
	} else if !exists {
		return nil, fmt.Errorf("%q does not have an Engine directory. Is it an Unreal Editor installation?", editorDir)
	}

	version, bvj, err := readEditorVersion(editorDir)
	if err != nil {
		return nil, fmt.Errorf("reading editor version: %w", err)
	}

	installed, err := checkEngineInstalled(editorDir)
	if err != nil {
		return nil, fmt.Errorf("checking if engine is installed: %w", err)
	}

	return &GunrealEditorConfig{
		Version:          version,
		Installed:        installed,
		BuildVersionFile: bvj,
	}, nil
}

func readEditorVersion(path string) (*goversion.Version, *buildVersionJson, error) {
	jsonPath := filepath.Join(path, "Engine", "Build", "Build.version")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading Build.version at %q: %w", jsonPath, err)
	}

	bvj := &buildVersionJson{}
	if err := json.Unmarshal(data, bvj); err != nil {
		return nil, nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	semver := fmt.Sprintf("%d.%d.%d", bvj.MajorVersion, bvj.MinorVersion, bvj.PatchVersion)
	version, err := goversion.NewVersion(semver)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing semver %q: %w", semver, err)
	}

	return version, bvj, nil
}

func checkEngineInstalled(path string) (bool, error) {
	installedMarkerPath := filepath.Join(path, "Engine", "Build", "InstalledBuild.txt")

	_, exists, err := files.StatFile(installedMarkerPath)
	if err != nil {
		return false, fmt.Errorf("statting %q: %w", installedMarkerPath, err)
	}

	return exists, nil
}
