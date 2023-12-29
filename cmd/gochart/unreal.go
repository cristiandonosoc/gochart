package main

import (
	"context"
	"fmt"

	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"github.com/urfave/cli/v2"
)

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

						gocharts, err := ugb.SearchGochartFiles(ctx)
						if err != nil {
							return fmt.Errorf("searching for gochart files: %w", err)
						}

						fmt.Printf("Listing gocharts at Unreal project at %s\n", ugb.Project.ProjectDir)
						fmt.Printf("Found %v gocharts:\n", len(gocharts))
						for _, gochart := range gocharts {
							fmt.Printf("- %s\n", gochart)
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
