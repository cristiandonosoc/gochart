package main

import (
	"context"
	"fmt"
	"path/filepath"

	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"github.com/urfave/cli/v2"
)

func relFilePath(projectPath, filePath string) string {
	rel, err := filepath.Rel(projectPath, filePath)
	if err != nil {
		return filePath
	}
	return rel
}

func unrealInvocationFlags() []cli.Flag {
	return []cli.Flag{}
}

func unrealCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:        "unreal",
			Description: "Commands associated with the Unreal integration",
			Flags:       frontend.SetFlags(),
			Subcommands: []*cli.Command{
				{
					Name:  "list",
					Usage: "list [PROJECT_DIR]",
					Action: func(cliCtx *cli.Context) error {
						if cliCtx.Args().Len() != 1 {
							return fmt.Errorf("Usage: %s", cliCtx.Command.Usage)
						}

						projectDir := cliCtx.Args().Get(0)

						ctx := context.Background()
						ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, projectDir)
						if err != nil {
							return fmt.Errorf("creating unreal backend: %w", err)
						}

						goutputs, err := ugb.List(ctx)
						if err != nil {
							return fmt.Errorf("listing project: %w", err)
						}

						fmt.Printf("Listing gocharts at Unreal project at %s\n", ugb.Project.ProjectDir)
						fmt.Printf("Found %v gocharts:\n", len(goutputs))
						for _, goutput := range goutputs {
							fmt.Println("-----------------------------------------------------------------------")
							uptodate := ""
							if !goutput.UpToDate() {
								uptodate = "*WILL REGENERATE*"
							}

							fmt.Printf("%s %s\n", goutput.Gochart.Name(), uptodate)
							fmt.Printf("- PATH: %s\n", relFilePath(ugb.Project.ProjectDir, goutput.Gochart.Path))
							fmt.Printf("- EXPECTED OUTPUTS:\n")
							for _, output := range goutput.Outputs {
								fmt.Printf("  > %s\n", relFilePath(ugb.Project.ProjectDir, output.ExpectedPath))
							}
						}

						return nil
					},
				},
				{
					Name:        "generate",
					Usage:       "generate (OPTIONS) [SOURCE_DIR]",
					Description: "Generate all the gochart files for an Unreal project",
					Flags:       unrealInvocationFlags(),
					Action: func(cliCtx *cli.Context) error {
						if cliCtx.Args().Len() != 1 {
							return fmt.Errorf("No source dir set. Usage: %s", cliCtx.Command.Usage)
						}

						projectDir := cliCtx.Args().Get(0)

						ctx := context.Background()
						ugb, err := unreal_backend.NewUnrealGochartBackend(ctx, projectDir)
						if err != nil {
							return fmt.Errorf("creating unreal backend: %w", err)
						}

						goutputs, err := ugb.Generate(ctx)
						if err != nil {
							return fmt.Errorf("generating gocharts: %w", err)
						}

						fmt.Println("Generation result:")
						for _, goutput := range goutputs {
							if goutput.UpToDate() {
								fmt.Printf("- %s UP TO DATE\n", goutput.Gochart.Name())
								continue
							}

							fmt.Printf("- %s:\n", goutput.Gochart.Name())
							for _, output := range goutput.Outputs {
								fmt.Printf("  > Generated %s\n", output.ExpectedPath)
							}
						}

						return nil
					},
				},
			},
		},
	}
}
