package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/option"

	"github.com/urfave/cli/v2"
)

const (
	CPPInvocationFlag_HeaderIncludePath = "header-include-path"
)

func backendCppInvocationFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag {
			Name: CPPInvocationFlag_HeaderIncludePath,
			Usage: "Change the `HEADER_INCLUDE_PATH` for the generated .cpp file. If not set, will asume same dir.",
		},
	}
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
					Flags: backendCppInvocationFlags(),
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

								headerPath := args.Get(1)
								cppPath := args.Get(2)

								var options []option.Option

								// If there is no header path, we pass it as an option.

								var headerIncludePath string
								if path := cliCtx.String(CPPInvocationFlag_HeaderIncludePath); path != "" {
									headerIncludePath = path
								} else {
									path, err := files.CppToHeader(cppPath)
									if err != nil {
										return fmt.Errorf("getting header from input C++ file: %w", err)
									}
									headerIncludePath = filepath.Base(path)
								}

								options = append(options, cpp.WithHeaderIncludePath(headerIncludePath))

								outputs, err := gc.Generate(args.Get(0), options...)
								if err != nil {
									return fmt.Errorf("generating C++ output files: %w", err)
								}

								outputCount := 0
								for _, output := range outputs {
									if output.Tag == "header" {
										if err := files.WriteToFile(headerPath, output.Content); err != nil {
											return fmt.Errorf("outputing header to %q: %w", headerPath, err)
										}

										outputCount++
										fmt.Printf("Output header to %q\n", headerPath)
									}

									if output.Tag == "cpp" {
										if err := files.WriteToFile(cppPath, output.Content); err != nil {
											return fmt.Errorf("outputing cpp to %q: %w", cppPath, err)
										}

										outputCount++
										fmt.Printf("Output cpp to %q\n", cppPath)
									}
								}

								if outputCount != 2 {
									return fmt.Errorf("expected to output 2 files, outputted %v", outputCount)
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
