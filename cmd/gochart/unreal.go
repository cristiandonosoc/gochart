package main

import (
	"context"
	"fmt"

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
						ui, err := unreal.NewUnrealIntegrator(context.Background(), sourceDir)
						if err != nil {
							return fmt.Errorf("creating unreal integrator: %w", err)
						}

						for _, gochart := range ui.Gocharts {
							fmt.Println("GOCHART:", gochart)
						}

						return nil
					},
				},
			},
		},
	}
}
