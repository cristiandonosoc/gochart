package main

import (
	"fmt"
	"os"

	// "github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/gochart"

	"github.com/urfave/cli/v2"
)

// func (yc *cppCommand) run(args []string) error {
// 	fs := flag.NewFlagSet("cpp", flag.ContinueOnError)
//
// 	frontendFlags := frontend.SetFlags(fs)
// 	backendFlags, err := backend.SetFlags("cpp", fs)
// 	if err != nil {
// 		return fmt.Errorf("setting flags for backend: %w", err)
// 	}
//
// 	if err := fs.Parse(args); err != nil {
// 		return fmt.Errorf("parsing flags: %w", err)
// 	}
//
// 	gf, err := frontend.Parse(frontendFlags)
// 	if err != nil {
// 		return fmt.Errorf("parsing frontend flags: %w", err)
// 	}
//
// 	bf, err := backend.Parse(backendFlags)
// 	if err != nil {
// 		return fmt.Errorf("parsing backend flags: %w", err)
// 	}
//
// 	gc, err := gochart.NewGochartCompiler(gf, bf)
// 	if err != nil {
// 		return fmt.Errorf("creating gochart compiler: %w", err)
// 	}
//
// 	gc.PrintIR(nil)
// 	return fmt.Errorf("CPPCOMMAND: IMPLEMENT ME")
// }

// func (c *cppCommand) Help() string {
// 	help := fmt.Sprintf(`
// Usage: gochart [global options] yaml [options] [PATH]
//
// 	%s
//
// Options:
// 	-frontend=(gochart|yaml)	The frontend language to read the statechart from.
// 														Default: gochart
// 	`, c.Synopsis())
//
// 	return strings.TrimSpace(help)
// }

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
				Name: "frontend",
				Flags: frontend.SetFlags(),
				Subcommands: []*cli.Command {
					{
						Name: "print",
						Description: "Print a common description of the output of the frontend. Useful for debugging.",
						Usage: "print [STATECHART_FILE]",
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
				Name: "generate",
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
