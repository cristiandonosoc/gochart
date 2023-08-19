// Package files provides minor support for dealing with files.
package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

func WriteToFile(path string, r io.Reader) error {
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

var cpps = []string{
	".cc",
	".cpp",
}

// SwapHeaderCpp swaps between .cc/.cpp into .h
func CppToHeader(path string) (string, error) {
	ext := filepath.Ext(path)

	if !slices.Contains(cpps, ext) {
		return "", fmt.Errorf("path %q does not have C++ extension (%v). Is it a C++ file?", path, cpps)
	}

	root := path[0 : len(path)-len(ext)]
	return root + ".h", nil
}
