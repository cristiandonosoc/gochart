package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/unreal"

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
					Name:        "generate",
					Usage:       "generate (OPTIONS) [SOURCE_DIR]",
					Description: "Generate all the gochart files for an Unreal project",
					Flags:       unrealInvocationFlags(),
					Action: func(cliCtx *cli.Context) error {
						if cliCtx.Args().Len() != 1 {
							return fmt.Errorf("No source dir set. Usage: %s", cliCtx.Command.Usage)
						}

						sourceDir := cliCtx.Args().Get(0)
						ui, err := unreal.NewUnrealIntegrator(nil, nil)
						if err != nil {
							return fmt.Errorf("creating unreal integrator: %w", err)
						}

						ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
						defer cancel()

						if err := ui.Index(ctx, sourceDir); err != nil {
							return fmt.Errorf("indexing %q: %w", sourceDir, err)
						}

						for _, gochart := range ui.Gocharts {
							fmt.Println("GOCHART:", gochart)
						}

						return fmt.Errorf("IMPLEMENT ME!")
					},
				},
			},
		},
	}
}
