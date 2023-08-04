package gochart_lang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/assert"
)

func TestGatherTokens(t *testing.T) {
	input := `
// This is a comment.
(( )){} // Grouping stuff.
[["this is a string literal"]]
"literal"
   "multiline
  string""after_line"
  statechart state transition trigger aabb a0123456789b
	std::string<std::vector*>&, std::unique_ptr<Foo**&>
`
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
		{id: Token_String, literal: "this is a string literal", line: 4, char: 3},
		{id: Token_RightBracket, line: 4, char: 29},
		{id: Token_RightBracket, line: 4, char: 30},

		{id: Token_String, literal: "literal", line: 5, char: 1},
		{id: Token_String, literal: `multiline
  string`, line: 6, char: 4},
		{id: Token_String, literal: "after_line", line: 7, char: 10},

		{id: Token_KeywordStatechart, literal: "statechart", line: 8, char: 3},
		{id: Token_KeywordState, literal: "state", line: 8, char: 14},
		{id: Token_KeywordTransition, literal: "transition", line: 8, char: 20},
		{id: Token_KeywordTrigger, literal: "trigger", line: 8, char: 31},

		{id: Token_Identifier, literal: "aabb", line: 8, char: 39},
		{id: Token_Identifier, literal: "a0123456789b", line: 8, char: 44},

		{id: Token_Identifier, literal: "std", line: 9, char: 2},
		{id: Token_Other, literal: ":", line: 9, char: 5},
		{id: Token_Other, literal: ":", line: 9, char: 6},
		{id: Token_Identifier, literal: "string", line: 9, char: 7},
		{id: Token_Other, literal: "<", line: 9, char: 13},
		{id: Token_Identifier, literal: "std", line: 9, char: 14},
		{id: Token_Other, literal: ":", line: 9, char: 17},
		{id: Token_Other, literal: ":", line: 9, char: 18},
		{id: Token_Identifier, literal: "vector", line: 9, char: 19},
		{id: Token_Other, literal: "*", line: 9, char: 25},
		{id: Token_Other, literal: ">", line: 9, char: 26},
		{id: Token_Other, literal: "&", line: 9, char: 27},
		{id: Token_Other, literal: ",", line: 9, char: 28},
		{id: Token_Identifier, literal: "std", line: 9, char: 30},
		{id: Token_Other, literal: ":", line: 9, char: 33},
		{id: Token_Other, literal: ":", line: 9, char: 34},
		{id: Token_Identifier, literal: "unique_ptr", line: 9, char: 35},
		{id: Token_Other, literal: "<", line: 9, char: 45},
		{id: Token_Identifier, literal: "Foo", line: 9, char: 46},
		{id: Token_Other, literal: "*", line: 9, char: 49},
		{id: Token_Other, literal: "*", line: 9, char: 50},
		{id: Token_Other, literal: "&", line: 9, char: 51},
		{id: Token_Other, literal: ">", line: 9, char: 52},

		{id: Token_EOF, line: 10, char: 0},
	}

	s := NewScanner()
	got, errors := s.gatherTokens(r)
	assert.Empty(t, errors)
	if errors == nil {
		compareTokens(t, want, got)
	}
}

func compareTokens(t *testing.T, want, got []*Token) {
	if !assert.Equal(t, len(want), len(got)) {
		printTokenErrors(t, want, got)
	} else {
		var errorLines []int

		// We compare each line and see if there are errors.
		for i := 0; i < len(want); i++ {
			want := want[i]
			got := got[i]

			isError := !assert.Equal(t, want, got, "Token %d, Want: %q, Got: %q", i, want.id, got.id)
			if isError {
				errorLines = append(errorLines, i)
			}
		}

		if len(errorLines) > 0 {
			printTokenErrors(t, want, got, errorLines...)
		}
	}
}

func printTokenErrors(t *testing.T, want, got []*Token, errorLines ...int) {
	t.Log("\nError Summary:\n")

	for i := 0; i < len(want) || i < len(got); i++ {
		isErrorLine := (xslices.Count(errorLines, i) > 0)

		wantStr := "Want: <empty>"
		if i < len(want) {
			wantStr = fmt.Sprintf("Want: %s", want[i])
		} else {
			isErrorLine = true
		}

		gotStr := "Got: <empty>"
		if i < len(got) {
			gotStr = fmt.Sprintf("Got: %s", got[i])
		} else {
			isErrorLine = true
		}

		if isErrorLine {
			t.Logf("[ERROR] Token %2d - %s, %s", i, wantStr, gotStr)
		} else {
			t.Logf("[   OK] Token %2d - %s, %s", i, wantStr, gotStr)
		}
	}
}
