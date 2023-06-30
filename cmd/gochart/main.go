package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

func readFrontend(path string) (*frontend.StatechartData, error) {
	yf := yaml.NewYamlFrontend()

	scdata, err := yf.ProcessFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("processign statechart yaml %q: %w", path, err)
	}

	return scdata, nil
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

func ensureDirectoriesExists(headerPath, bodyPath string) error {
	if ok, err := ensureDirExists(filepath.Dir(headerPath)); err != nil {
		return fmt.Errorf("ensuring %q owning directory exists: %w", headerPath, err)
	} else if !ok {
		return fmt.Errorf("parent path for %q is not a directory", headerPath)
	}

	if ok, err := ensureDirExists(filepath.Dir(bodyPath)); err != nil {
		return fmt.Errorf("ensuring %q owning directory exists: %w", bodyPath, err)
	} else if !ok {
		return fmt.Errorf("parent path for %q is not a directory", bodyPath)
	}

	return nil
}

func ensureDirExists(path string) (bool, error) {
	if info, err := os.Stat(path); err != nil {
		return false, fmt.Errorf("stat %q: %w", path, err)
	} else {
		if !info.IsDir() {
			return false, nil
		}
	}

	return true, nil
}

func internalMain() error {
	var yamlPath string
	var headerPath string
	var bodyPath string
	onlyPrint := true

	if len(os.Args) == 2 {
		yamlPath = os.Args[1]
	} else if len(os.Args) == 4 {
		onlyPrint = false

		yamlPath = os.Args[1]
		headerPath = os.Args[2]
		bodyPath = os.Args[3]
	} else {
		return fmt.Errorf("Usage: gochart <PATH> [<HEADER_PATH> <BODY_PATH>]")
	}

	scdata, err := readFrontend(yamlPath)
	if err != nil {
		return fmt.Errorf("reading frontend: %w", err)
	}

	sc, err := ir.ProcessStatechartData(scdata)
	if err != nil {
		return fmt.Errorf("processing statechart data: %w", err)
	}

	fmt.Println("Trigger Count:", len(sc.Triggers))
	for name := range sc.Triggers {
		fmt.Println("Trigger:", name)
	}

	backend := cpp.NewCppGochartBackend(func(o *cpp.BackendOptions) {
		// For now we just assume the include is in the same directory.
		if headerPath != "" {
			o.HeaderInclude = filepath.Base(headerPath)
		}
	})
	headerData, bodyData, err := backend.Generate(sc)
	if err != nil {
		return fmt.Errorf("generating backend: %w", err)
	}

	if onlyPrint {
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
	} else {
		if err := writeToFile(headerPath, headerData); err != nil {
			return fmt.Errorf("writing header: %w", err)
		}
		fmt.Printf("Wrote header to %s\n", headerPath)

		if err := writeToFile(bodyPath, bodyData); err != nil {
			return fmt.Errorf("writing body: %w", err)
		}
		fmt.Printf("Wrote body to %s\n", bodyPath)
	}

	return nil
}

func main() {
	if err := internalMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
