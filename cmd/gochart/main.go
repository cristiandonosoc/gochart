package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/gochart"

	"github.com/mitchellh/cli"
)

var (
	flags = struct {
		input         string
		printFrontend bool
		printIR       bool
		printOutput   bool
		headerOutput  string
		bodyOutput    string

		verbose bool
	}{}
)

func parseFlags() error {
	flag.BoolVar(&flags.printFrontend, "print-frontend", false,
		"Whether we should just print the result of the frontend.")
	flag.BoolVar(&flags.printIR, "print-ir", false,
		"Whether we should just print the result of the IR pass")
	flag.BoolVar(&flags.printOutput, "print-output", false,
		"Whether we should just print eh result of the backend, instead of generating files")

	flag.StringVar(&flags.headerOutput, "header-output", "", "Where to output the header (.h) file. Cpp backend only.")
	flag.StringVar(&flags.bodyOutput, "body-output", "", "Where to output the body (.cpp) file. Cpp backend only.")
	flag.BoolVar(&flags.verbose, "verbose", false, "Print verbose debug output")

	flag.Parse()

	if !flags.printFrontend && !flags.printIR {
		if !flags.printOutput {
			if flags.headerOutput == "" || flags.bodyOutput == "" {
				return fmt.Errorf("if --print-only is not set, then --header-output and --body-output must be set")
			}
		}
	}

	if len(flag.Args()) != 1 {
		return fmt.Errorf("Wrong amount of arguments")
	}
	flags.input = flag.Args()[0]

	return nil
}

func writeToFile(path string, r io.Reader) error {
	// Create or truncate the target file.
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating/truncating file %q: %w", path, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		return fmt.Errorf("copying contents to %q: %w", path, err)
	}

	// Ensure the file is written.
	if err := out.Sync(); err != nil {
		return fmt.Errorf("calling sync on %q: %w", path, err)
	}

	return nil
}

var (
	baseFlags = struct {
	}{}
)

type cppCommand struct{}

func (yc *cppCommand) run(args []string) error {
	fs := flag.NewFlagSet("cpp", flag.ContinueOnError)

	frontendFlags := frontend.SetFlags(fs)
	backendFlags, err := backend.SetFlags("cpp", fs)
	if err != nil {
		return fmt.Errorf("setting flags for backend: %w", err)
	}

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	gf, err := frontend.Parse(frontendFlags)
	if err != nil {
		return fmt.Errorf("parsing frontend flags: %w", err)
	}

	bf, err := backend.Parse(backendFlags)
	if err != nil {
		return fmt.Errorf("parsing backend flags: %w", err)
	}

	gc, err := gochart.NewGochartCompiler(gf, bf)
	if err != nil {
		return fmt.Errorf("creating gochart compiler: %w", err)
	}

	gc.PrintIR(nil)
	return fmt.Errorf("CPPCOMMAND: IMPLEMENT ME")
}

func (yc *cppCommand) Run(args []string) int {
	if err := yc.run(args); err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

func (c *cppCommand) Help() string {
	help := fmt.Sprintf(`
Usage: gochart [global options] yaml [options] [PATH]

	%s

Options:
	-frontend=(gochart|yaml)	The frontend language to read the statechart from.
														Default: gochart
	`, c.Synopsis())

	return strings.TrimSpace(help)
}

func (c *cppCommand) Synopsis() string {
	return "Generates C++ code from statechart."
}

func internalMain() error {
	c := cli.NewCLI("gochart", "1.0.0")
	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"generate cpp": func() (cli.Command, error) {
			return &cppCommand{}, nil
		},
	}

	exitCode, err := c.Run()
	if err != nil {
		return fmt.Errorf("running cli: %w", err)
	}

	if exitCode != 0 {
		return fmt.Errorf("Error running gochart")
	}

	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
