package gochart_lang

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserSimple(t *testing.T) {
	input := "statechart Simple {}"
	r := strings.NewReader(input)

	want := &ASTNodeRoot{
		statecharts: []*ASTNodeStatechart{
			{
				Name: "Simple",
			},
		},
	}

	scanner := NewScanner()
	tokens, errors := scanner.gatherTokens(r)
	require.Empty(t, errors)

	parser := NewParser()
	got, errors := parser.Parse(tokens)
	require.Empty(t, errors)

	compareASTNodeRoot(t, want, got)
}

func compareASTNodeRoot(t *testing.T, want, got *ASTNodeRoot) {
	if assert.Equal(t, len(want.statecharts), len(got.statecharts)) {
		for i := 0; i < len(want.statecharts); i++ {
			want := want.statecharts[i]
			got := got.statecharts[i]

			compareASTNodeStatechart(t, want, got)
		}
	}
}

func compareASTNodeStatechart(t *testing.T, want, got *ASTNodeStatechart) {
	assert.Equal(t, want.Name, got.Name)
}
