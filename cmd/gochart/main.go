package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

func internalMain() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("Usage: gochart <PATH>")
	}

	fmt.Println("-----------------------------------------------------------------------------------")

	yf := yaml.NewYamlFrontend()

	path := os.Args[1]
	scdata, err := yf.ProcessFromFile(path)
	if err != nil {
		return fmt.Errorf("processign statechart yaml %q: %w", path, err)
	}

	{
		encoded, err := json.MarshalIndent(scdata, "", "  ")
		if err != nil {
			return fmt.Errorf("marshalling to json: %w", err)
		}

		fmt.Println(string(encoded))
	}

	fmt.Println("-----------------------------------------------------------------------------------")

	sc, err := ir.ProcessStatechartData(scdata)
	if err != nil {
		return fmt.Errorf("processing statechart data: %w", err)
	}

	{
		for _, state := range sc.States {
			fmt.Printf("State %s\n", state.Name)
		}
	}

	fmt.Println("-----------------------------------------------------------------------------------")

	backend := cpp.NewCppGochartBackend()
	headerData, err := backend.Generate(sc)
	if err != nil {
		return fmt.Errorf("generating header: %w", err)
	}

	header, err := io.ReadAll(headerData)
	if err != nil {
		return fmt.Errorf("reading the header data: %w", err)
	}

	fmt.Println(string(header))

	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
