package syntactic

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/files"
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

	for _, rf := range runfiles {

		t.Run(filepath.Base(rf), func(t *testing.T) {
			lf, err := files.LoadFileFromPath(rf)
			require.NoError(t, err)

			_, wantErrorText, err := test_support.ExtractExpectedError(t, bytes.NewReader(lf.Data))
			require.NoError(t, err)

			_, gotErr := gf.Process(lf)

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
