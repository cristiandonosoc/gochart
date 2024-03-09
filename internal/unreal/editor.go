package unreal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cristiandonosoc/golib/pkg/files"

	goversion "github.com/hashicorp/go-version"
)

type Editor struct {
	Version *goversion.Version

	// For internal tracking information. Exposed via functions if needed.
	bvj *buildVersionJson
}

func NewEditor(path string) (*Editor, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("abs %q: %w", path, err)
	}
	path = abs

	// Sanity check that there is an Engine directory.
	if exists, err := files.DirExists(filepath.Join(path, "Engine")); err != nil {
		return nil, fmt.Errorf("checking if engine dir exists: %w", err)
	} else if !exists {
		return nil, fmt.Errorf("%q does not have an Engine directory. Is it an Unreal Editor installation?", path)
	}

	version, bvj, err := readEditorVersion(path)
	if err != nil {
		return nil, fmt.Errorf("reading editor version: %w", err)
	}

	return &Editor{
		Version: version,
		bvj:     bvj,
	}, nil
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

func (e *Editor) Describe() (string, error) {
	var sb strings.Builder

	sb.WriteString("EDITOR -------------------------------------------------------------------\n\n")
	sb.WriteString(fmt.Sprintf("- VERSION: %s\n", e.Version))

	return sb.String(), nil
}
