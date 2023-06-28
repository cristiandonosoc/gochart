// cpp is a Gochart backend meant to generate C++ code for statechart.
package cpp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

type cppGochartBackend struct {
}

func NewCppGochartBackend() *cppGochartBackend {
	return &cppGochartBackend{}
}

func (cpp *cppGochartBackend) GenerateToFiles(sc *ir.Statechart, headerPath, bodyPath string) error {
	if err := ensureDirectoriesExists(headerPath, bodyPath); err != nil {
		return fmt.Errorf("ensuring dir exists: %w", err)
	}

	// We generate the contents of the files.
	header, body, err := cpp.Generate(sc)
	if err != nil {
		return fmt.Errorf("generating file contents: %w", err)
	}

	// Output the header.
	if err := writeToFile(headerPath, header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Output the body.
	if err := writeToFile(bodyPath, body); err != nil {
		return fmt.Errorf("writing body: %w", err)
	}

	return nil
}

func (cpp *cppGochartBackend) Generate(sc *ir.Statechart) (_header, _body io.Reader, _err error) {
	common := &commonContext{
		Version: "DEVELOPMENT",
		Time:    time.Now(),
	}

	tm, err := newTemplateManager(sc, common)
	if err != nil {
		return nil, nil, fmt.Errorf("building new template manager: %w", err)
	}

	header, err := tm.generateHeader()
	if err != nil {
		return nil, nil, fmt.Errorf("generating header: %w", err)
	}

	body, err := tm.generateBody()
	if err != nil {
		return nil, nil, fmt.Errorf("generating body: %w", err)
	}

	return header, body, nil
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
