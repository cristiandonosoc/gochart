// Package test_support adds some testing utilities, especially for detecting when running on Bazel.
package test_support

import (
	"fmt"
	"os"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

func toUnix(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// RunfilePath tries to find a testdata file in a build system agnostic way, working for both Bazel
// environments and default Go ones.
func RunfilePath(path string) (string, error) {
	// We attempt to query Bazel to see if it can find runfiles.
	runfiles, err := bazel.ListRunfiles()
	if err != nil {
		return "", fmt.Errorf("listing runfiles: %w", err)
	}

	// If we get runfiles, then we are running in a Bazel mode, so we return the files from there.
	if len(runfiles) > 0 {
		// Bazel returns the runfiles as relative from the workspace root, so for now we do a simple
		// suffix match. This might need some more thought in the future.
		path = toUnix(path)
		for _, rf := range runfiles {
			if strings.HasSuffix(toUnix(rf.ShortPath), path) {
				return rf.Path, nil
			}
		}

		return "", fmt.Errorf("cannot find %q", path)
	}

	// If not, we attempt to find the file normally, as that would work for normal Go invocations.
	return path, nil
}

// ReadRunfile tries to read a file using the loading rules of |RunfilePath|.
func ReadRunfile(path string) ([]byte, error) {
	rp, err := RunfilePath(path)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(rp)
}
