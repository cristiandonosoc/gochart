package parsing

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsing(t *testing.T) {
	runfiles, err := test_support.Runfiles("testdata")
	require.NoError(t, err)

	gf, err := gochart_lang.NewGochartLangFrontend()
	assert.NoError(t, err)

	t.Log("RUNFILES:", runfiles)
	for _, rf := range runfiles {

		t.Run(filepath.Base(rf), func(t *testing.T) {
			data, err := os.ReadFile(rf)
			require.NoError(t, err)

			normalText, wantErrorText, err := extractError(t, bytes.NewReader(data))
			require.NoError(t, err)

			_, gotErr := gf.Process(strings.NewReader(normalText))

			if wantErrorText == "" {
				require.NoError(t, gotErr)
				return
			}

			if wantErrorText != gotErr.Error() {
				t.Log("WANT:\n", wantErrorText)
				t.Log("---------------------------")
				t.Log("GOT:\n", gotErr.Error())
				assert.Equal(t, wantErrorText, gotErr.Error())
			}
		})
	}
}

const (
	errorTagBegin = "// --- ERROR BEGIN"
	errorTagEnd   = "// --- ERROR END"
)

func extractError(t *testing.T, r io.Reader) (_normalLines, _errorLines string, _err error) {
	scanner := bufio.NewScanner(r)

	var errorLines []string
	var normalLines []string
	collectingError := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, errorTagBegin) {
			require.False(t, collectingError)
			collectingError = true
			continue
		}

		if strings.HasPrefix(line, errorTagEnd) {
			require.True(t, collectingError)
			collectingError = false
			continue
		}

		// If we're collecting the error, we add it to our array of error lines.
		// Otherwise, we add it to the normal lines.
		if collectingError {
			require.True(t, strings.HasPrefix(line, "// "))

			line = strings.TrimPrefix(line, "// ")
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
