package main

import (
	"fmt"
	"io"
	"os"

	// "github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/gochart"

	"github.com/urfave/cli/v2"
)

func getGochartCompiler(backendKind string, cliCtx *cli.Context) (*gochart.GochartCompiler, error) {
	// Get the frontend.
	gf, err := frontend.Parse(cliCtx)
	if err != nil {
		return nil, fmt.Errorf("determining frontend: %w", err)
	}

	// If the is a backend set, we get the backend.
	var gb backend.GochartBackend
	if backendKind != "" {
		result, err := backend.Parse(backendKind, cliCtx)
		if err != nil {
			return nil, fmt.Errorf("determining backend: %w", err)
		}
		gb = result
	}

	return gochart.NewGochartCompiler(gf, gb)
}

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
							gc, err := getGochartCompiler("", cliCtx)
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
							gc, err := getGochartCompiler("", cliCtx)
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
				Name:        "backend",
				Description: "Commands associated with the Gochart backends, which generate code/files.",
				Flags:       frontend.SetFlags(),
				Subcommands: []*cli.Command{
					{
						Name:        "cpp",
						Description: "Command associated with the C++ backend.",
						Subcommands: []*cli.Command{
							{
								Name:        "print",
								Description: "Print the generated C++ output.",
								Usage:       "print [STATECHART_FILE]",
								Action: func(cliCtx *cli.Context) error {
									gc, err := getGochartCompiler("cpp", cliCtx)
									if err != nil {
										return err
									}

									if cliCtx.Args().Len() != 1 {
										return fmt.Errorf("No input file set. Usage: %s", cliCtx.Command.Usage)
									}

									outputs, err := gc.Generate(cliCtx.Args().Get(0))
									if err != nil {
										return fmt.Errorf("generating C++ output files: %w", err)
									}

									for _, output := range outputs {
										fmt.Println(output.Tag)
										fmt.Println("---------------------------------------------------------------------")

										content, err := io.ReadAll(output.Content)
										if err != nil {
											return fmt.Errorf("reading content for output %q: %w", output.Tag, err)
										}

										fmt.Println(string(content))
									}

									return nil
								},
							},
							{
								Name:        "generate",
								Description: "Generate the C++ files",
								Usage:       "generate [STATECHART_FILE] [HEADER_OUTPUT] [CPP_OUTPUT]",
								Action: func(cliCtx *cli.Context) error {
									gc, err := getGochartCompiler("cpp", cliCtx)
									if err != nil {
										return err
									}

									if cliCtx.Args().Len() != 3 {
										return fmt.Errorf("Not correct amout of input files. Usage: %s", cliCtx.Command.Usage)
									}

									args := cliCtx.Args()
									if err := gc.GenerateToFiles(args.Get(0), args.Get(1), args.Get(2)); err != nil {
										return fmt.Errorf("generating C++ to files: %w", err)
									}

									return nil
								},
							},
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
