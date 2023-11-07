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

	return

	t.Log("RUNFILES:", runfiles)
	for _, rf := range runfiles {

		t.Run(filepath.Base(rf), func(t *testing.T) {
			// Read the file.
			t.Log("READING:", rf)
			data, err := os.ReadFile(rf)
			require.NoError(t, err)

			wantErrorText := extractError(t, bytes.NewReader(data))
			_, gotErr := gf.Process(bytes.NewReader(data))

			if wantErrorText == "" {
				require.NoError(t, gotErr)
				return
			}

			assert.Equal(t, wantErrorText, gotErr.Error())
		})
	}
}

const (
	errorTagBegin = "// ERROR BEGIN:"
	errorTagEnd   = "// ERROR END:"
)

func extractError(t *testing.T, r io.Reader) string {
	scanner := bufio.NewScanner(r)

	var lines []string
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

		if collectingError {
			require.True(t, strings.HasPrefix(line, "//"))

			line = strings.TrimPrefix(line, "//")
			lines = append(lines, line)
		}
	}

	require.False(t, collectingError)
	require.NoError(t, scanner.Err())

	return strings.Join(lines, "\n")
}
