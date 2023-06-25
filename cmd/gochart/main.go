package main

import (
	"fmt"
	"os"
	"io"

	"github.com/cristiandonosoc/gochart/pkg/backend/cpp"
)

func internalMain() error {

	backend := cpp.NewCppGochartBackend()
	headerData, err := backend.Generate(nil)
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
