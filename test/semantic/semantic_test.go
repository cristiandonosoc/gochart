package semantic

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"
	"github.com/cristiandonosoc/golib/pkg/files"

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
			base := filepath.Base(rf)
			lf, err := files.LoadFileFromPathWithKey(base, rf, true)
			require.NoError(t, err)

			_, wantErrorText, err := test_support.ExtractExpectedError(t, bytes.NewReader(lf.Data))
			require.NoError(t, err)

			gcdata, err := gf.Process(lf)
			require.NoError(t, err)

			_, gotErr := ir.ProcessGochartData(gcdata)
			require.NoError(t, err)

			if wantErrorText == "" {
				require.NoError(t, gotErr)
				return
			}
			require.Error(t, gotErr)

			if wantErrorText != gotErr.Error() {
				t.Log("WANT:\n", wantErrorText)
				t.Log("---------------------------")
				t.Log("GOT:\n", gotErr.Error())
				assert.Equal(t, wantErrorText, gotErr.Error())
			}
		})
	}
}
