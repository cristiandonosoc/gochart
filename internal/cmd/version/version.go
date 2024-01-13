// Package version holds common cli logic to handle with versioning.
package version

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/build"

	"github.com/urfave/cli/v2"
)

func printVersion() {
	fmt.Println("Version:", build.BuildVersion())
	fmt.Println("Build Date:", build.BuildTime())
	fmt.Println("Git Commit:", build.GitCommit())
}

func AppendVersionData(app *cli.App) {
	app.Flags = append(app.Flags, &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Value:   false,
		Action: func(cliCtx *cli.Context, value bool) error {
			printVersion()
			return fmt.Errorf("")
		},
	})

	app.Commands = append(app.Commands, &cli.Command{
		Name:        "version",
		Description: "Print the version of the app",
		Action: func(cliCtx *cli.Context) error {
			printVersion()
			return fmt.Errorf("")
		},
	})
}
