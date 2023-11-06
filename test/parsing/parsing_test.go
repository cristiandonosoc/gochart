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
			// Read the file.
			t.Log("READING:", rf)
			data, err := os.ReadFile(rf)
			require.NoError(t, err)

			errorLines, err := extractExpectedErrorLines(bytes.NewReader(data))
			require.NoError(t, err)
			assert.Empty(t, errorLines)


			// Compile the statecharts.
			t.Log("STATECHART:", string(data))
			_, err = gf.Process(bytes.NewReader(data))
			t.Log("ERROR:", err.Error())
			assert.Error(t, err)

		})
	}
}

const (
	errorTag = "// EXPECTED ERROR: "
)

func extractExpectedErrorLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, errorTag) {
			lines = append(lines, strings.TrimPrefix(line, errorTag))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
