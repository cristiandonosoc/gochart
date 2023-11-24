package test_support

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	ErrorTagBegin = "// --- ERROR BEGIN"
	ErrorTagEnd   = "// --- ERROR END"
)

// ExtractExpectedError reads an input files and extracts the errors between the marked tags so that
// they can be compared to the output given by the code.
func ExtractExpectedError(t *testing.T, r io.Reader) (_normalLines, _errorLines string, _err error) {
	scanner := bufio.NewScanner(r)

	var errorLines []string
	var normalLines []string
	collectingError := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, ErrorTagBegin) {
			require.False(t, collectingError)
			collectingError = true
			continue
		}

		if strings.HasPrefix(line, ErrorTagEnd) {
			require.True(t, collectingError)
			collectingError = false
			continue
		}

		// If we're collecting the error, we add it to our array of error lines.
		// Otherwise, we add it to the normal lines.
		if collectingError {
			require.True(t, strings.HasPrefix(line, "//"))

			line = strings.TrimPrefix(line, "//")
			errorLines = append(errorLines, strings.TrimRight(line, " \r\t\n"))
		} else {
			normalLines = append(normalLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	require.False(t, collectingError)
	require.NoError(t, scanner.Err())

	return strings.Join(normalLines, "\n"), strings.Join(errorLines, "\n"), nil
}
