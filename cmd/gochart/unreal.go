package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cristiandonosoc/gochart/internal/option"
	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"github.com/urfave/cli/v2"
)

const (
	kFlag_UnrealStrategy = "unreal_strategy"
)

func relFilePath(projectPath, filePath string) string {
	rel, err := filepath.Rel(projectPath, filePath)
	if err != nil {
		return filePath
	}
	return rel
}

func unrealInvocationFlags() []cli.Flag {
	var flags []cli.Flag

	flags = append(flags, frontend.SetFlags()...)

	// Add the strategy flag.
	{
		strategies := unreal_backend.ListGenerationStrategies()

		options := make([]string, 0, len(strategies))
		for _, strategy := range strategies {
			options = append(options, strategy.String())
		}

		flags = append(flags, &cli.StringFlag{
			Name:     kFlag_UnrealStrategy,
			Usage:    fmt.Sprintf("The `STRATEGY` about where to generate the output files. Values %s", options),
			Required: true,
		})
	}

	return flags
}

func unrealOptionsFromFlags(cliCtx *cli.Context) ([]option.Option, error) {
	var options []option.Option

	{
		strategyValue := cliCtx.String(kFlag_UnrealStrategy)
		strategy := unreal_backend.NewGenerationStrategy(strategyValue)
		if !strategy.Valid() {
			return nil, fmt.Errorf("Invalid strategy %q", strategyValue)
		}

		options = append(options, unreal_backend.WithStrategy(strategy))
	}

	return options, nil
}

func unrealCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:        "unreal",
			Description: "Commands associated with the Unreal integration",
			Subcommands: []*cli.Command{
				{
					Name:  "list",
					Usage: "list [PROJECT_DIR]",
					Flags: unrealInvocationFlags(),
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

						options, err := unrealOptionsFromFlags(cliCtx)
						if err != nil {
							return fmt.Errorf("parsing options from flags: %w", err)
						}

						result, err := ugb.List(ctx, options...)
						if err != nil {
							return fmt.Errorf("listing project: %w", err)
						}

						fmt.Printf("Listing gocharts at Unreal project at %s\n", ugb.Project.ProjectDir)
						fmt.Printf("Found %v gocharts:\n\n", len(result.Gocharts))
						for _, goutput := range result.Gocharts {
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
							fmt.Println("-----------------------------------------------------------------------")
						}

						if len(result.OrphanedGocharts) > 0 {
							fmt.Println("")
							fmt.Println("***********************************************************************")
							fmt.Printf("Found %d Orphaned Gocharts\n\n", len(result.OrphanedGocharts))

							for _, orphaned := range result.OrphanedGocharts {
								fmt.Printf("%s\n", orphaned.Gochart.Name())
								fmt.Printf("- PATH: %s [<file not found>]\n", relFilePath(ugb.Project.ProjectDir, orphaned.Gochart.Path))
								fmt.Printf("- POINTED BY:\n")
								for _, output := range orphaned.Outputs {
									fmt.Printf("   > %s\n", relFilePath(ugb.Project.ProjectDir, output.ExpectedPath))
								}
								fmt.Println("-----------------------------------------------------------------------")
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

						options, err := unrealOptionsFromFlags(cliCtx)
						if err != nil {
							return fmt.Errorf("parsing options from flags: %w", err)
						}

						goutputs, err := ugb.Generate(ctx, options...)
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
