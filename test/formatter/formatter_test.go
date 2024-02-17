package formatter

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/backend/formatter"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"
	"github.com/cristiandonosoc/golib/pkg/files"

	"github.com/stretchr/testify/require"
)

func TestEnsureFormatterDoesNotChangeAST(t *testing.T) {
	runfiles, err := test_support.Runfiles("testdata")
	require.NoError(t, err)

	gf, err := gochart_lang.NewGochartLangFrontend()
	require.NoError(t, err)

	ff, err := formatter.NewFormatterGochartBackend()
	require.NoError(t, err)

	for _, rf := range runfiles {
		t.Run(filepath.Base(rf), func(t *testing.T) {
			base := filepath.Base(rf)
			wantLF, err := files.LoadFileFromPathWithKey(base+"-want", rf, true)
			require.NoError(t, err)

			wantGcData, err := gf.Process(wantLF)
			require.NoError(t, err)

			wantGc, err := ir.ProcessGochartData(wantGcData)
			require.NoError(t, err)

			outputs, err := ff.Generate(context.Background(), wantGc)
			require.NoError(t, err)
			require.Equal(t, len(outputs), 1)

			output := outputs[0]
			require.NotNil(t, output.Content)
			content, err := io.ReadAll(output.Content)
			require.NoError(t, err)

			gotLF, err := files.NewFromData(base+"-got", content, true)
			require.NoError(t, err)

			gotGcData, err := gf.Process(gotLF)
			require.NoError(t, err)

			gotGc, err := ir.ProcessGochartData(gotGcData)
			require.NoError(t, err)

			ir.CompareGochart(t, wantGc, gotGc)
		})
	}
}

func TestThatFormatterIsIdempotent(t *testing.T) {
	runfiles, err := test_support.Runfiles("testdata")
	require.NoError(t, err)

	gf, err := gochart_lang.NewGochartLangFrontend()
	require.NoError(t, err)

	ff, err := formatter.NewFormatterGochartBackend()
	require.NoError(t, err)

	for _, rf := range runfiles {
		t.Run(filepath.Base(rf), func(t *testing.T) {
			base := filepath.Base(rf)
			// We load the file for the first time.
			lf, err := files.LoadFileFromPathWithKey(base, rf, true)
			require.NoError(t, err)

			gcdata, err := gf.Process(lf)
			require.NoError(t, err)

			wantGc, err := ir.ProcessGochartData(gcdata)
			require.NoError(t, err)

			// Now that we have the formatted lf, we generate it over and over a couple of times and
			// ensure the output is the same byte to byte.
			for i := 0; i < 5; i++ {
				wantOutputs, err := ff.Generate(context.Background(), wantGc)
				require.NoError(t, err)
				require.Equal(t, len(wantOutputs), 1)

				wantOutput := wantOutputs[0]
				require.NotNil(t, wantOutput.Content)
				wantContent, err := io.ReadAll(wantOutput.Content)
				require.NoError(t, err)

				wantLF, err := files.NewFromData(base+"-want", wantContent, true)
				require.NoError(t, err)

				wantGcData, err := gf.Process(wantLF)
				require.NoError(t, err)

				wantGc, err := ir.ProcessGochartData(wantGcData)
				require.NoError(t, err)

				// We generate the output from this.
				gotOutputs, err := ff.Generate(context.Background(), wantGc)
				require.NoError(t, err)

				require.Equal(t, len(gotOutputs), 1)

				gotOutput := gotOutputs[0]
				require.NotNil(t, gotOutput.Content)
				gotContent, err := io.ReadAll(gotOutput.Content)
				require.NoError(t, err)

				require.Equal(t, wantContent, gotContent)

				gotLF, err := files.NewFromData(base+"-got", gotContent, true)
				require.NoError(t, err)

				gotGcData, err := gf.Process(gotLF)
				require.NoError(t, err)

				gotGc, err := ir.ProcessGochartData(gotGcData)
				require.NoError(t, err)

				ir.CompareGochart(t, wantGc, gotGc)

				// Now we set the current gochart as the next one and do the loop again.
				wantGc = gotGc
			}
		})
	}
}
