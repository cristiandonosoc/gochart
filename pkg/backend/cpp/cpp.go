// cpp is a Gochart backend meant to generate C++ code for statechart.
package cpp

import (
	"fmt"
	"io"
	"os"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

type cppGochartBackend struct {
}

func NewCppGochartBackend() *cppGochartBackend {
	return &cppGochartBackend{}
}

func (cpp *cppGochartBackend) GenerateToFiles(sc *ir.Statechart, headerPath, bodyPath string) error {
	// if ok, err := ensureDirExists(filepath.Dir(headerPath)); err != nil {
	// 	return fmt.Errorf("ensuring %q owning directory exists: %w", headerPath, err)
	// } else if !ok {
	// 	return fmt.Errorf("parent path for %q is not a directory", headerPath)
	// }

	// if ok, err := ensureDirExists(filepath.Dir(bodyPath)); err != nil {
	// 	return fmt.Errorf("ensuring %q owning directory exists: %w", bodyPath, err)
	// } else if !ok {
	// 	return fmt.Errorf("parent path for %q is not a directory", bodyPath)
	// }

	panic("IMPLEMENT ME")
}

func (cpp *cppGochartBackend) Generate(sc *ir.Statechart) (header io.Reader, err error) {
	return generateHeader(sc)
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
