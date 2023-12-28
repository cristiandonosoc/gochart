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
				// {
				// 	Name:        "generate",
				// 	Usage:       "generate (OPTIONS) [SOURCE_DIR]",
				// 	Description: "Generate all the gochart files for an Unreal project",
				// 	Flags:       unrealInvocationFlags(),
				// 	Action: func(cliCtx *cli.Context) error {
				// 		if cliCtx.Args().Len() != 1 {
				// 			return fmt.Errorf("No source dir set. Usage: %s", cliCtx.Command.Usage)
				// 		}

				// 		gc, err := getGochartCompiler("cpp", cliCtx)
				// 		if err != nil {
				// 			return fmt.Errorf("getting gochart compiler: %w", err)
				// 		}

				// 		sourceDir := cliCtx.Args().Get(0)
				// 		ui, err := unreal.NewUnrealIntegrator(gc.Frontend, gc.Backend)
				// 		if err != nil {
				// 			return fmt.Errorf("creating unreal integrator: %w", err)
				// 		}

				// 		ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
				// 		defer cancel()

				// 		if err := ui.Index(ctx, sourceDir); err != nil {
				// 			return fmt.Errorf("indexing %q: %w", sourceDir, err)
				// 		}

				// 		if errors := ui.Generate("Win64", unreal.GenerationStrategy_ModuleGeneratedDir); len(errors) != 0 {
				// 			for _, err := range errors {
				// 				fmt.Printf("generating gocharts: %v\n", err)
				// 			}
				// 			return fmt.Errorf("FAILED")
				// 		}

				// 		return nil
				// 	},
				// },
			},
		},
	}
}
