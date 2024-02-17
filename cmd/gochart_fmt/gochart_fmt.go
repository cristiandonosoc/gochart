// Binary gochart_fmt is a standalone application that is meant to format gocharts.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cristiandonosoc/gochart/pkg/backend/formatter"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"
	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/urfave/cli/v2"
)

func internalMain() error {
	app := cli.App{
		Name:           "gochart_fmt",
		Usage:          "Utility to format gochart files",
		Flags:          []cli.Flag{},
		DefaultCommand: "format",
		Commands: []*cli.Command{
			{
				Name:        "format",
				Usage:       "Format a given format file",
				ArgsUsage:   "(flags) [GOCHART_FILE]",
				Description: "Parse a gochart file and print the formatted result into stdio.",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "debug",
						Value: false,
						Usage: "Whether to use the debug printer (mostly for development)",
					},
				},
				Action: func(cliCtx *cli.Context) error {
					if cliCtx.Args().Len() != 1 {
						return fmt.Errorf("Usage for %q: %s", cliCtx.Command.FullName(), cliCtx.Command.ArgsUsage)
					}

					path := cliCtx.Args().Get(0)

					gf, err := gochart_lang.NewGochartLangFrontend()
					if err != nil {
						return fmt.Errorf("creating gochart lang frontend: %w", err)
					}

					gb, err := formatter.NewFormatterGochartBackend(formatter.WithDebug(cliCtx.Bool("debug")))
					if err != nil {
						return fmt.Errorf("creating formatter backend: %w", err)
					}

					lf, err := files.LoadFileFromPath(path)
					if err != nil {
						return fmt.Errorf("loading file %q: %w", path, err)
					}

					gcdata, err := gf.Process(lf)
					if err != nil {
						return fmt.Errorf("processing gochart frontend: %w", err)
					}

					gc, err := ir.ProcessGochartData(gcdata)
					if err != nil {
						return fmt.Errorf("processing IR data: %w", err)
					}

					outputs, err := gb.Generate(context.Background(), gc)
					if err != nil {
						return fmt.Errorf("formatting file: %w", err)
					}

					if len(outputs) != 1 {
						return fmt.Errorf("expected only one output, got %d", len(outputs))
					}

					content, err := io.ReadAll(outputs[0].Content)
					if err != nil {
						return fmt.Errorf("reading all content: %w", err)
					}
					fmt.Println(string(content))

					return nil
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
