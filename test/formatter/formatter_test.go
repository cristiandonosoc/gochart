package formatter

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/backend/formatter"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/ir"

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
			wantLF, err := files.LoadFileFromPath(rf)
			require.NoError(t, err)

			wantScData, err := gf.Process(wantLF)
			require.NoError(t, err)

			wantSc, err := ir.ProcessStatechartData(wantScData)
			require.NoError(t, err)

			outputs, err := ff.Generate(context.Background(), wantSc)
			require.NoError(t, err)
			require.Equal(t, len(outputs), 1)

			output := outputs[0]
			require.NotNil(t, output.Content)
			content, err := io.ReadAll(output.Content)
			require.NoError(t, err)

			gotLF, err := files.LoadFileFromData(content)
			require.NoError(t, err)

			gotScData, err := gf.Process(gotLF)
			require.NoError(t, err)

			gotSc, err := ir.ProcessStatechartData(gotScData)
			require.NoError(t, err)

			ir.CompareStatechart(t, wantSc, gotSc)
		})
	}
}
