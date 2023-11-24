package logical

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogicalErrors(t *testing.T) {
	runfiles, err := test_support.Runfiles("testdata")
	require.NoError(t, err)

	gf, err := gochart_lang.NewGochartLangFrontend()
	assert.NoError(t, err)

	for _, rf := range runfiles {

		t.Run(filepath.Base(rf), func(t *testing.T) {
			data, err := os.ReadFile(rf)
			require.NoError(t, err)

			normalText, wantErrorText, err := test_support.ExtractExpectedError(t, bytes.NewReader(data))
			require.NoError(t, err)

			scdata, err := gf.Process(strings.NewReader(normalText))
			require.NoError(t, err)

			_, gotErr := ir.ProcessStatechartData(scdata)
			require.NoError(t, err)

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
