// Package test_support holds common testing utilities.
package test_support

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/cristiandonosoc/golib/pkg/files"
	"github.com/cristiandonosoc/golib/pkg/test_detection"
)

// TestTmpDir returns a valid base string to be fed to os.MkdirTemp.
func TestTmpBase() string {
	if test_detection.RunningAsBazelTest() {
		os.Getenv("TEST_TMPDIR")
	}

	// Fallback to the system default.
	return ""
}

// Runfiles returns a list of all the runfiles associated with this test that contains |dir|.
// Typical use is Runfiles("testdata")
func Runfiles(dir string) ([]string, error) {
	if !test_detection.RunningAsTest() {
		return nil, fmt.Errorf("should only be called for tests")
	}

	var candidates []string
	if test_detection.RunningAsBazelTest() {
		bazelCandidates, err := bazelCandidatesRunfiles(dir)
		if err != nil {
			return nil, fmt.Errorf("reading bazel runfiles: %w", err)
		}
		candidates = bazelCandidates
	} else {
		// Otherwise we open the dir and list it.
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("opening dir %q: %w", dir, err)
		}

		for _, entry := range entries {
			candidates = append(candidates, filepath.Join(dir, entry.Name()))
		}
	}

	// For now we just query single level. This could be extended for recursive files.
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		stat, err := os.Stat(candidate)
		if err != nil {
			return nil, fmt.Errorf("os stat %q: %w", candidate, err)
		}

		if stat.IsDir() {
			continue

		}

		result = append(result, candidate)
	}

	return result, nil
}

func bazelCandidatesRunfiles(dir string) ([]string, error) {
	// We attempt to query Bazel to see if it can find runfiles.
	runfiles, err := bazel.ListRunfiles()
	if err != nil {
		return nil, fmt.Errorf("listing runfiles: %w", err)
	}

	var candidates []string
	if len(runfiles) > 0 {
		for _, rf := range runfiles {
			if !strings.Contains(files.ToUnixPath(rf.ShortPath), dir) {
				continue
			}

			candidates = append(candidates, rf.Path)
		}
	}

	return candidates, nil
}

// RunfilePath tries to find a testdata file in a build system agnostic way, working for both Bazel
// environments and default Go ones.
func RunfilePath(path string) (string, error) {
	if !test_detection.RunningAsTest() {
		return "", fmt.Errorf("should only be called for tests")
	}

	if test_detection.RunningAsBazelTest() {
		// We attempt to query Bazel to see if it can find runfiles.
		runfiles, err := bazel.ListRunfiles()
		if err != nil {
			return "", fmt.Errorf("listing runfiles: %w", err)
		}

		// If we get runfiles, then we are running in a Bazel mode, so we return the files from there.
		if len(runfiles) > 0 {
			// Bazel returns the runfiles as relative from the workspace root, so for now we do a simple
			// suffix match. This might need some more thought in the future.
			path = files.ToUnixPath(path)
			for _, rf := range runfiles {
				if strings.HasSuffix(files.ToUnixPath(rf.Path), path) {
					return rf.Path, nil
				}
			}

		}

		return "", fmt.Errorf("cannot find %q", path)
	}

	// If not, we attempt to find the file normally, as that would work for normal Go invocations.
	if _, found, err := files.StatFile(path); err != nil || !found {
		return "", fmt.Errorf("stat %q: %w (found: %t)", path, err, found)
	}

	return path, nil
}

// LoadRunfile tries to read a file using the loading rules of |RunfilePath|.
func LoadRunfile(path string) (*files.LoadedFile, error) {
	if !test_detection.RunningAsTest() {
		return nil, fmt.Errorf("should only be called for tests")
	}

	rp, err := RunfilePath(path)
	if err != nil {
		return nil, err
	}

	return files.LoadFileFromPath(rp)
}
