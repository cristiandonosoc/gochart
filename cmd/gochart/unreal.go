package main

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/option"
	unreal_backend "github.com/cristiandonosoc/gochart/pkg/backend/unreal"
	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"github.com/urfave/cli/v2"
)

const (
	kFlag_UnrealStrategy = "unreal_strategy"
)

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
	unrealSection := &cli.Command{
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
						printGoutput(ugb, goutput)
						fmt.Println("-----------------------------------------------------------------------")
					}

					if len(result.OrphanedGocharts) > 0 {
						fmt.Println("")
						fmt.Println("***********************************************************************")
						fmt.Printf("Found %d Orphaned Gocharts\n\n", len(result.OrphanedGocharts))

						for _, orphaned := range result.OrphanedGocharts {
							printOrphanedGoutput(ugb, orphaned)
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
	}

	unrealSection.Subcommands = append(unrealSection.Subcommands, obtainCleanCommands()...)
	sort.Slice(unrealSection.Subcommands, func(i, j int) bool {
		return unrealSection.Subcommands[i].Name < unrealSection.Subcommands[j].Name
	})

	return []*cli.Command{unrealSection}
}

const (
	kFlag_Clean_Filter = "filter"
)

func obtainCleanCommands() []*cli.Command {
	flags := unrealInvocationFlags()
	flags = append(flags, &cli.StringFlag{
		Name:  kFlag_Clean_Filter,
		Usage: fmt.Sprintf("The `FILTER` that will be used to narrow which elements it will match. Substring matching"),
	})

	clean := &cli.Command{
		Name:        "clean",
		Description: "Commands associated with cleaning Gochart integration",
		Usage:       "Commands associated with cleaning Gochart integration",
		Subcommands: []*cli.Command{
			{
				Name:  "orphaned",
				Usage: "orphaned (OPTIONS) [PROJECT_DIR]",
				Flags: flags,
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

					if len(result.OrphanedGocharts) == 0 {
						fmt.Println("No orphaned Gocharts found")
						return nil
					}

					// Try to use the filter.
					filter := cliCtx.String(kFlag_Clean_Filter)
					if filter != "" {
						slices.DeleteFunc(result.OrphanedGocharts, func(goutput *unreal_backend.GochartOutput) bool {
							return !strings.Contains(goutput.Gochart.Name(), filter)
						})
					}

					if len(result.OrphanedGocharts) == 0 {
						return fmt.Errorf("No Gochart Name matched the filter %q. Perhaps relax the filter?\n", filter)
					}

					fmt.Println("Going to clean the following orphaned Gochart files:")
					fmt.Println("")
					for _, orphaned := range result.OrphanedGocharts {
						printOrphanedGoutput(ugb, orphaned)
						fmt.Println("-----------------------------------------------------------------------")
					}

					if ok, err := userPrompt("Delete the listed generated Gochart C++ files?"); err != nil {
						return err
					} else if !ok {
						fmt.Println("Aborting.")
						return nil
					}

					deleted, err := ugb.CleanOrphaned(ctx, filter, options...)
					if err != nil {
						return fmt.Errorf("cleaning orphaned files: %w", err)
					}

					fmt.Printf("Deleted %d files:\n", len(deleted))
					for _, d := range deleted {
						fmt.Printf("- %s\n", relFilePath(ugb.Project.ProjectDir, d))
					}

					return nil
				},
			},
		},
	}

	return []*cli.Command{clean}
}
