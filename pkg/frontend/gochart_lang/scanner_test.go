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
"literal"
   "multiline
  string""after_line"`
	r := strings.NewReader(input)

	want := []*Token{
		{id: Token_LeftParen, line: 3, char: 1},
		{id: Token_LeftParen, line: 3, char: 2},
		{id: Token_RightParen, line: 3, char: 4},
		{id: Token_RightParen, line: 3, char: 5},
		{id: Token_LeftBrace, line: 3, char: 6},
		{id: Token_RightBrace, line: 3, char: 7},

		{id: Token_LeftBracket, line: 4, char: 1},
		{id: Token_LeftBracket, line: 4, char: 2},
		{id: Token_StringLiteral, literal: "this is a string literal", line: 4, char: 3},
		{id: Token_RightBracket, line: 4, char: 29},
		{id: Token_RightBracket, line: 4, char: 30},

		{id: Token_StringLiteral, literal: "literal", line: 5, char: 1},
		{id: Token_StringLiteral, literal: `multiline
  string`, line: 6, char: 4},
		{id: Token_StringLiteral, literal: "after_line", line: 7, char: 10},
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
		assert.Equal(t, want.line, got.line, "token %d", i)
		assert.Equal(t, want.char, got.char, "token %d", i)
	}
}
