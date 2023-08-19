package main

import (
	"fmt"
	"os"

	// "github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/gochart"

	"github.com/urfave/cli/v2"
)

func internalMain() error {
	app := cli.App{
		Name:  "gochart",
		Usage: "Generate statechart",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Value: false,
				Usage: "Whether to print more debug output",
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "frontend",
				Description: "Commands associated with the frontends. Normally used for debugging.",
				Flags:       frontend.SetFlags(),
				Subcommands: []*cli.Command{
					{
						Name:        "print",
						Description: "Print a common description of the output of the frontend. Useful for debugging.",
						Usage:       "print [STATECHART_FILE]",
						Action: func(cliCtx *cli.Context) error {
							gf, err := frontend.Parse(cliCtx)
							if err != nil {
								return fmt.Errorf("parsing frontend: %w", err)
							}

							gc, err := gochart.NewGochartCompiler(gf, nil)
							if err != nil {
								return err
							}

							if cliCtx.Args().Len() != 1 {
								return fmt.Errorf("No input file set. Usage: %s", cliCtx.Command.Usage)
							}

							if err := gc.PrintFrontend(cliCtx.Args().Get(0), os.Stdout); err != nil {
								return fmt.Errorf("printing frontend: %w", err)
							}

							return nil
						},
					},
				},
			},
			{
				Name:        "ir",
				Description: "Commands associated with the IR (Intermediate Representation). Normally used for debugging.",
				Flags:       frontend.SetFlags(),
				Subcommands: []*cli.Command{
					{
						Name:        "print",
						Description: "Prints a common description of the output after processing the frontend's output.",
						Usage:       "print [STATECHART_FILE]",
						Action: func(cliCtx *cli.Context) error {
							gf, err := frontend.Parse(cliCtx)
							if err != nil {
								return fmt.Errorf("parsing frontend: %w", err)
							}

							gc, err := gochart.NewGochartCompiler(gf, nil)
							if err != nil {
								return err
							}

							if cliCtx.Args().Len() != 1 {
								return fmt.Errorf("No input file set. Usage: %s", cliCtx.Command.Usage)
							}

							if err := gc.PrintIR(cliCtx.Args().Get(0), os.Stdout); err != nil {
								return fmt.Errorf("printing frontend: %w", err)
							}

							return nil
						},
					},
				},
			},
			{
				Name:  "generate",
				Flags: frontend.SetFlags(),
				Subcommands: []*cli.Command{
					{
						Name:  "cpp",
						Usage: "Generate C++ backend",
						Action: func(cliCtx *cli.Context) error {
							return fmt.Errorf("GENERATE C++ IMPLEMENT ME")
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
