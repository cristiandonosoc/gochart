package main

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"github.com/urfave/cli/v2"
)

const (
	CPPInvocationFlag_HeaderIncludePath = "header-include-path"
)

func backendCppInvocationFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  CPPInvocationFlag_HeaderIncludePath,
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

								outputs, err := gc.Generate(cliCtx.Args().Get(0))
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
							Usage:       "generate [STATECHART_FILE] [FORWARD_DECLARE] [DECLARE] [DEFINE]",
							Action: func(cliCtx *cli.Context) error {
								gc, err := getGochartCompiler("cpp", cliCtx)
								if err != nil {
									return err
								}

								if cliCtx.Args().Len() != 4 {
									return fmt.Errorf("Incorrect amount of input files. Usage: %s", cliCtx.Command.Usage)
								}

								args := cliCtx.Args()

								forwardDeclarePath := args.Get(1)
								declarePath := args.Get(2)
								definePath := args.Get(3)

								var options []option.Option

								// If there is no header path, we pass it as an option.

								var headerIncludePath string
								if path := cliCtx.String(CPPInvocationFlag_HeaderIncludePath); path != "" {
									headerIncludePath = path
								} else {
									// path, err := files.CppToHeader(cppPath)
									// if err != nil {
									// 	return fmt.Errorf("getting header from input C++ file: %w", err)
									// }
									// headerIncludePath = filepath.Base(path)
								}

								options = append(options, cpp.WithHeaderIncludePath(headerIncludePath))

								outputs, err := gc.Generate(args.Get(0), options...)
								if err != nil {
									return fmt.Errorf("generating C++ output files: %w", err)
								}

								outputCount := 0
								for _, output := range outputs {
									if output.Name == "forward_declare" {
										if err := files.WriteToFile(forwardDeclarePath, output.Content); err != nil {
											return fmt.Errorf("outputing forward declare to %q: %w", forwardDeclarePath, err)
										}

										outputCount++
										fmt.Printf("Output forward declare to %q\n", forwardDeclarePath)
									}

									if output.Name == "declare" {
										if err := files.WriteToFile(declarePath, output.Content); err != nil {
											return fmt.Errorf("outputing declare  to %q: %w", declarePath, err)
										}

										outputCount++
										fmt.Printf("Output declare to %q\n", declarePath)
									}

									if output.Name == "define" {
										if err := files.WriteToFile(definePath, output.Content); err != nil {
											return fmt.Errorf("outputing define  to %q: %w", definePath, err)
										}

										outputCount++
										fmt.Printf("Output define to %q\n", definePath)
									}

								}

								if outputCount != 3 {
									return fmt.Errorf("expected to output 3 files, outputted %v", outputCount)
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
