package gochart_lang

import (
	"bytes"
	"os"
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

func TestReadSimple(t *testing.T) {
	data, err := os.ReadFile("testdata/example.gochart")
	require.NoError(t, err)

	scanner := NewScanner()
	tokens, errors := scanner.gatherTokens(bytes.NewReader(data))
	require.Empty(t, errors)

	parser := NewParser()
	got, errors := parser.Parse(tokens)
	require.Empty(t, errors)

	want := &ASTNodeRoot{
		statecharts: []*ASTNodeStatechart{
			{
				Name: "Example",
				Triggers: []*ASTNodeTrigger{
					{
						Name: "Execute",
					},
					{
						Name:       "Update",
						ArgsString: "float dt",
					},
					{
						Name:       "SomeMessage",
						ArgsString: "const hvn::Message* message, int other, const std::string& another",
					},
				},
			},
		},
	}

	compareASTNodeRoot(t, want, got)
}

// Helpers -----------------------------------------------------------------------------------------

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

	if assert.Equal(t, len(want.Triggers), len(got.Triggers)) {
		for i := 0; i < len(want.Triggers); i++ {
			want := want.Triggers[i]
			got := got.Triggers[i]

			compareASTNodeTrigger(t, want, got)
		}
	}
}

func compareASTNodeTrigger(t *testing.T, want, got *ASTNodeTrigger) {
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.ArgsString, got.ArgsString)
}
