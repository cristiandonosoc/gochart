package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cristiandonosoc/gochart/pkg/backend"
	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

var (
	flags = struct {
		input         string
		frontendKind  frontend.FrontendKind
		backendKind   backend.BackendKind
		printFrontend bool
		printIR       bool
		printOutput   bool
		headerOutput  string
		bodyOutput    string
	}{}
)

func parseFlags() error {
	frontendFlag := flag.String("frontend", "yaml", "What frontend to use")
	backendFlag := flag.String("backend", "cpp", "Want backend to generate to")
	flag.BoolVar(&flags.printFrontend, "print-frontend", false,
		"Whether we should just print the result of the frontend.")
	flag.BoolVar(&flags.printIR, "print-ir", false,
		"Whether we should just print the result of the IR pass")
	flag.BoolVar(&flags.printOutput, "print-backend", false,
		"Whether we should just print eh result of the backend, instead of generating files")
	flag.StringVar(&flags.headerOutput, "header-output", "", "Where to output the header (.h) file. Cpp backend only.")
	flag.StringVar(&flags.bodyOutput, "body-output", "", "Where to output the body (.cpp) file. Cpp backend only.")

	flag.Parse()

	if err := flags.frontendKind.Set(*frontendFlag); err != nil {
		return err
	}

	if err := flags.backendKind.Set(*backendFlag); err != nil {
		return err
	}

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

func internalMain() error {
	if err := parseFlags(); err != nil {
		flag.Usage()
		fmt.Println("Usage: gochart [FLAGS] INPUT")
		return fmt.Errorf("parsing commandline: %w", err)
	}

	// Get the frontend and read the yaml.
	var gf frontend.GochartFrontend
	switch flags.frontendKind {
	case frontend.FrontendKind_Gochart:
		gf = gochart_lang.NewGochartLangFrontend()
	case frontend.FrontendKind_Yaml:
		gf = yaml.NewYamlFrontend()
	}

	frontendData, err := frontend.ProcessFromFile(gf, flags.input)
	if err != nil {
		return fmt.Errorf("parsing input file: %w", err)
	}

	if flags.printFrontend {
		printer := frontend.NewPrinter(os.Stdout)
		printer.PrintStatechart(frontendData)
		return nil
	}

	// Process the generated data into intermediary format.
	irData, err := ir.ProcessStatechartData(frontendData)
	if err != nil {
		return fmt.Errorf("processing statechart data: %w", err)
	}

	if flags.printIR {
		printer := ir.NewPrinter(os.Stdout)
		printer.PrintStatechart(irData)
		return nil
	}

	var gb backend.GochartBackend
	if flags.backendKind == backend.BackendKind_Cpp {
		gb = cpp.NewCppGochartBackend(func(o *cpp.BackendOptions) {
			// For now we just assume the include is in the same directory.
			o.HeaderInclude = filepath.Base(flags.headerOutput)
		})
	}

	headerData, bodyData, err := gb.Generate(irData)
	if err != nil {
		return fmt.Errorf("generating backend: %w", err)
	}

	if flags.printOutput {
		header, err := io.ReadAll(headerData)
		if err != nil {
			return fmt.Errorf("reading the header data: %w", err)
		}

		body, err := io.ReadAll(bodyData)
		if err != nil {
			return fmt.Errorf("reading the body data: %w", err)
		}

		fmt.Println("HEADER *****")
		fmt.Println(string(header))
		fmt.Println("BODY *****")
		fmt.Println(string(body))

		return nil
	}

	if err := writeToFile(flags.headerOutput, headerData); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	fmt.Printf("Wrote header to %s\n", flags.headerOutput)

	if err := writeToFile(flags.bodyOutput, bodyData); err != nil {
		return fmt.Errorf("writing body: %w", err)
	}
	fmt.Printf("Wrote body to %s\n", flags.bodyOutput)

	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
