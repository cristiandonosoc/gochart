package gochart_lang

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatherTokens(t *testing.T) {
	input := `
// This is a comment.
(( )){} // Grouping stuff.
[["this is a string literal"]]
"literal"`
	r := strings.NewReader(input)

	want := []*Token{
		{id: Token_LeftParen},
		{id: Token_LeftParen},
		{id: Token_RightParen},
		{id: Token_RightParen},
		{id: Token_LeftBrace},
		{id: Token_RightBrace},

		{id: Token_LeftBracket},
		{id: Token_LeftBracket},
		{id: Token_StringLiteral, literal: "this is a string literal"},
		{id: Token_RightBracket},
		{id: Token_RightBracket},

		{id: Token_StringLiteral, literal: "literal"},
	}

	s := NewScanner()
	got, errors := s.gatherTokens(r)
	assert.Empty(t, errors)

	compareTokens(t, want, got)
}

func compareTokens(t *testing.T, want, got []*Token) {
	assert.Equal(t, len(got), len(want))

	for i := 0; i < len(got); i++ {
		got := got[i]
		want := want[i]

		assert.Equal(t, want.id, got.id, "token %d", i)
		assert.Equal(t, want.literal, got.literal, "token %d", i)
	}
}
