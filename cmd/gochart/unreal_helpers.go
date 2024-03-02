package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
)

func relFilePath(projectPath, filePath string) string {
	rel, err := filepath.Rel(projectPath, filePath)
	if err != nil {
		return filePath
	}
	return rel
}

func printGoutput(ugb *unreal_backend.UnrealGochartBackend, goutput *unreal_backend.GochartOutput) {
	uptodate := "[UP TO DATE]"
	if !goutput.UpToDate() {
		uptodate = "[*** WILL REGENERATE ***]"
	}

	fmt.Printf("%s %s\n", goutput.Gochart.Name(), uptodate)

	fmt.Printf("- PATH: %s", relFilePath(ugb.Project.ProjectDir, goutput.Gochart.Path))
	if modtime := goutput.Gochart.ModTime(); !modtime.IsZero() {
		fmt.Printf(" [modtime: %v]", modtime.Format("2006/01/02 15:04:05"))
	}
	fmt.Printf("\n")

	fmt.Printf("- EXPECTED OUTPUTS:\n")
	for _, output := range goutput.Outputs {
		fmt.Printf("  > %s", relFilePath(ugb.Project.ProjectDir, output.ExpectedPath))
		if modtime := output.ModTime(); !modtime.IsZero() {
			fmt.Printf(" [modtime: %v]", modtime.Format("2006/01/02 15:04:05"))
		} else {
			fmt.Printf(" [<file not found>]")
		}

		fmt.Printf("\n")
	}
}

func printOrphanedGoutput(ugb *unreal_backend.UnrealGochartBackend, orphaned *unreal_backend.GochartOutput) {
	fmt.Printf("%s\n", orphaned.Gochart.Name())
	fmt.Printf("- PATH: %s [<file not found>]\n", relFilePath(ugb.Project.ProjectDir, orphaned.Gochart.Path))
	fmt.Printf("- POINTED BY:\n")
	for _, output := range orphaned.Outputs {
		fmt.Printf("   > %s\n", relFilePath(ugb.Project.ProjectDir, output.ExpectedPath))
	}
}

func userPrompt(prompt string) (bool, error) {
	fmt.Printf("%s (y/N): ", prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading input: %w", err)
	}

	response := strings.ToLower(strings.TrimSpace(input))
	if response == "" {
		return false, nil
	}

	switch response {
	case "y":
		return true, nil
	case "n":
		return false, nil
	default:
		return false, fmt.Errorf("wrong option %q. Please answer 'y' or 'n'", response)
	}
}
