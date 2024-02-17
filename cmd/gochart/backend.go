package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"github.com/urfave/cli/v2"
)

const (
	CPPInvocationFlag_HeaderIncludePath = "header-include-path"
)

func backendCppInvocationFlags() []cli.Flag {
	return []cli.Flag{}
}

func backendCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:        "backend",
			Description: "Commands associated with the Gochart backends, which generate code/files.",
			Flags:       frontend.SetFlags(),
			Subcommands: []*cli.Command{
				{
					Name:        "cpp",
					Flags:       backendCppInvocationFlags(),
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

								// TODO(cdc): Define a better context strategy.
								ctx := context.Background()

								outputs, err := gc.Generate(ctx, cliCtx.Args().Get(0))
								if err != nil {
									return fmt.Errorf("generating C++ output files: %w", err)
								}

								for _, output := range outputs {
									fmt.Println(output.Name)
									fmt.Println("---------------------------------------------------------------------")

									content, err := io.ReadAll(output.Content)
									if err != nil {
										return fmt.Errorf("reading content for output %q: %w", output.Name, err)
									}

									fmt.Println(string(content))
								}

								return nil
							},
						},
						{
							Name:        "generate",
							Description: "Generate the C++ files",
							Usage:       "generate [STATECHART_FILE] [OUTPUT_PATH] [FILE_PREFIX]",
							Action: func(cliCtx *cli.Context) error {
								gc, err := getGochartCompiler("cpp", cliCtx)
								if err != nil {
									return err
								}

								if cliCtx.Args().Len() != 3 {
									return fmt.Errorf("Incorrect amount of input files. Usage: %s", cliCtx.Command.Usage)
								}

								args := cliCtx.Args()

								outputDir, err := filepath.Abs(args.Get(1))
								if err != nil {
									return fmt.Errorf("getting absolute path from %q: %w", args.Get(1), err)
								}

								options := []option.Option{
									cpp.WithOutputDir(outputDir),
									cpp.WithFilePrefix(args.Get(2)),
								}

								// TODO(cdc): Define a better context strategy.
								ctx := context.Background()

								outputs, err := gc.Generate(ctx, args.Get(0), options...)
								if err != nil {
									return fmt.Errorf("generating C++ output files: %w", err)
								}

								if len(outputs) != 3 {
									return fmt.Errorf("expected to output 3 files, outputted %d (%v)", len(outputs), outputs)
								}

								for _, output := range outputs {
									data, err := io.ReadAll(output.Content)
									if err != nil {
										return fmt.Errorf("reading content: %w", err)
									}

									if err := os.WriteFile(output.ExpectedPath, data, 0644); err != nil {
										return fmt.Errorf("outputing %q to %q: %w", output.Name, output.ExpectedPath, err)
									}
									fmt.Println("Generated", output.ExpectedPath)
								}

								return nil
							},
						},
					},
				},
			},
		},
	}

}
