// Package test_detection is a minor package that helps to detect whether the program is running
// under a test or not.
package test_detection

import (
	"os"
	"sync"
	"testing"
)

var bazelCheckOnce sync.Once
var gRunningAsBazelTest bool

// RunningAsTest checks to see if this program is running as a test. This will return true both for
// the `go test` case that the `bazel test` case.
func RunningAsTest() bool {
	if RunningAsBazelTest() {
		return true
	}

	if testing.Testing() {
		return true
	}

	return false
}

// RunningAsBazelTest is similar to |RunningAsTest|, though only returning true if invoked via
// `bazel test`.
func RunningAsBazelTest() bool {
	bazelCheckOnce.Do(func() {
		// Sadly this is not send on windows, so we rely on one of the other environment variables
		// sent by Bazel as a proxy.
		// See https://github.com/bazelbuild/bazel/issues/21420
		if os.Getenv("BAZEL_TEST") == "1" {
			gRunningAsBazelTest = true
			return
		}

		// We use a couple of the "Bazel envs". This should give use "some" level of confidence.
		envs := []string{
			"TEST_TARGET",
			"TEST_WORKSPACE",
			"TEST_TMPDIR",
		}

		for _, env := range envs {
			// As soon of any of this is empty, we assume we're not running on Bazel.
			if os.Getenv(env) == "" {
				return
			}
		}

		gRunningAsBazelTest = true
	})

	return gRunningAsBazelTest
}
