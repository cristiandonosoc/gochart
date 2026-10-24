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

var (
	gVersion_5_2 = goversion.Must(goversion.NewVersion("5.2.0"))
	gVersion_5_3 = goversion.Must(goversion.NewVersion("5.3.0"))
	gVersion_5_4 = goversion.Must(goversion.NewVersion("5.4.0"))

	gVersionConstraints = goversion.MustConstraints(goversion.NewConstraint(">= 5.2, < 5.4"))
)

type GunrealEditorConfig struct {
	EditorDir string `yaml:"editor_dir"`

	// (optional) Which dotnet to use for invoking the tooling.
	// If not set, it will tried to be found in the Unreal installation.
	Dotnet string `yaml:"dotnet"`

	Version *goversion.Version

	// Installed determines whether this is considered an installed version (eg. installing via
	// Unreal Launcher) vs a source build.
	Installed bool

	// UBTDll will normally be discovered via the editor.
	UBTDll string

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
	sb.WriteString(fmt.Sprintf("- EDITOR DIR: %s\n", gec.EditorDir))
	sb.WriteString(fmt.Sprintf("- VERSION: %s\n", gec.Version))
	sb.WriteString(fmt.Sprintf("- INSTALLED: %t\n", gec.Installed))
	sb.WriteString(fmt.Sprintf("- DOTNET: %s\n", gec.Dotnet))
	sb.WriteString(fmt.Sprintf("- UBT DLL: %s\n", gec.UBTDll))

	return sb.String()
}

func resolveEditorConfig(configPath string, gec *GunrealEditorConfig) error {
	if gec == nil {
		return fmt.Errorf("no editor key found")
	}

	// Sanity check that there is an Engine directory.
	if exists, err := files.DirExists(filepath.Join(gec.EditorDir, "Engine")); err != nil {
		return fmt.Errorf("checking if engine dir exists: %w", err)
	} else if !exists {
		return fmt.Errorf("%q does not have an Engine directory. Is it an Unreal Editor installation?", gec.EditorDir)
	}

	version, bvj, err := readEditorVersion(gec)
	if err != nil {
		return fmt.Errorf("reading editor version: %w", err)
	}
	gec.Version = version
	gec.BuildVersionFile = bvj

	installed, err := checkEngineInstalled(gec)
	if err != nil {
		return fmt.Errorf("checking if engine is installed: %w", err)
	}
	gec.Installed = installed

	dotnet, err := resolveDotnet(configPath, gec)
	if err != nil {
		return fmt.Errorf("resolving dotnet: %w", err)
	}
	gec.Dotnet = dotnet

	ubt, err := resolveUBT(configPath, gec)
	if err != nil {
		return fmt.Errorf("Resolving ubt: %w", err)
	}
	gec.UBTDll = ubt

	return nil
}

func readEditorVersion(gec *GunrealEditorConfig) (*goversion.Version, *buildVersionJson, error) {
	jsonPath := filepath.Join(gec.EditorDir, "Engine", "Build", "Build.version")

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

	if !gVersionConstraints.Check(version) {
		return nil, nil, fmt.Errorf("editor version %q does not comply constraints %q", version, gVersionConstraints)
	}

	return version, bvj, nil
}

func checkEngineInstalled(gec *GunrealEditorConfig) (bool, error) {
	installedMarkerPath := filepath.Join(gec.EditorDir, "Engine", "Build", "InstalledBuild.txt")

	_, exists, err := files.StatFile(installedMarkerPath)
	if err != nil {
		return false, fmt.Errorf("statting %q: %w", installedMarkerPath, err)
	}

	return exists, nil
}

func resolveDotnet(configPath string, gec *GunrealEditorConfig) (string, error) {
	// Check if user set their own specific dotnet.
	if gec.Dotnet != "" {
		dotnet, err := checkFile(configPath, gec.Dotnet)
		if err != nil {
			return "", fmt.Errorf("checking user provided dotnet: %w", err)
		}
		return dotnet, nil
	}

	if gec.Version.LessThan(gVersion_5_4) {
		dotnetPath := filepath.Join(gec.EditorDir, "Engine", "Binaries", "ThirdParty", "DotNET", "6.0.302", "windows", "dotnet.exe")
		dotnet, err := checkFile(configPath, dotnetPath)
		if err != nil {
			if gec.Installed {
				return "", fmt.Errorf("checking dotnet: %w. Was the installation done correctly?", err)
			} else {
				return "", fmt.Errorf("checking dotnet: %w. Did you run Setup.bat?", err)
			}
		}

		return dotnet, nil
	}

	return "", fmt.Errorf("unsupported version %q", gec.Version)
}

func resolveUBT(configPath string, gec *GunrealEditorConfig) (string, error) {
	if gec.Version.LessThan(gVersion_5_4) {
		ubtPath := filepath.Join(gec.EditorDir, "Engine", "Binaries", "DotNet", "UnrealBuildTool", "UnrealBuildTool.dll")

		ubt, err := checkFile(configPath, ubtPath)
		if err != nil {
			if gec.Installed {
				return "", fmt.Errorf("checking ubt: %w. Was the installation done correctly?", err)
			} else {
				return "", fmt.Errorf("checking ubt: %w. Did you run Setup.bat?", err)
			}
		}

		return ubt, nil
	}

	return "", fmt.Errorf("unsupported version %q", gec.Version)
}
