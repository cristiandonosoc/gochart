package gochart_lang

import (
	"fmt"
	"slices"
	"testing"

	"github.com/cristiandonosoc/golib/pkg/files"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherTokens(t *testing.T) {
	input := `
// This is a comment.
(( )){} // Grouping stuff.
<<[["this is a string literal"]]>>
"literal"
   "multiline true
  string""after_line enter exit gochart state transition trigger null"
  gochart state transition aabb a0123456789b null
	trigger SomeTrigger(std::string<std::vector*>&, std::unique_ptr<void(Foo**&)>),=,
	,= true false true
123456789 22
333, 4444)
enter exit
capture capture if {
	if (nested) {
		if something
	}
} do {
	do whatever
}
`

	want := []*Token{
		{id: Token_Comment, literal: "This is a comment.", line: 2, char: 1},

		{id: Token_LeftParen, literal: "(", line: 3, char: 1},
		{id: Token_LeftParen, literal: "(", line: 3, char: 2},
		{id: Token_RightParen, literal: ")", line: 3, char: 4},
		{id: Token_RightParen, literal: ")", line: 3, char: 5},
		{id: Token_LeftBrace, literal: "{", line: 3, char: 6},
		{id: Token_RightBrace, literal: "}", line: 3, char: 7},
		{id: Token_Comment, literal: "Grouping stuff.", line: 3, char: 9},

		{id: Token_LeftAngleBracket, literal: "<", line: 4, char: 1},
		{id: Token_LeftAngleBracket, literal: "<", line: 4, char: 2},
		{id: Token_LeftBracket, literal: "[", line: 4, char: 3},
		{id: Token_LeftBracket, literal: "[", line: 4, char: 4},
		{id: Token_String, literal: "this is a string literal", line: 4, char: 5},
		{id: Token_RightBracket, literal: "]", line: 4, char: 31},
		{id: Token_RightBracket, literal: "]", line: 4, char: 32},
		{id: Token_RightAngleBracket, literal: ">", line: 4, char: 33},
		{id: Token_RightAngleBracket, literal: ">", line: 4, char: 34},

		{id: Token_String, literal: "literal", line: 5, char: 1},
		{id: Token_String, literal: `multiline true
  string`, line: 6, char: 4},
		{id: Token_String, literal: "after_line enter exit gochart state transition trigger null", line: 7, char: 10},

		{id: Token_KeywordGochart, literal: "gochart", line: 8, char: 3},
		{id: Token_KeywordState, literal: "state", line: 8, char: 11},
		{id: Token_KeywordTransition, literal: "transition", line: 8, char: 17},

		{id: Token_Identifier, literal: "aabb", line: 8, char: 28},
		{id: Token_Identifier, literal: "a0123456789b", line: 8, char: 33},
		{id: Token_KeywordNull, literal: "null", line: 8, char: 46},

		{id: Token_KeywordTrigger, literal: "trigger", line: 9, char: 2},
		{id: Token_Identifier, literal: "SomeTrigger", line: 9, char: 10},
		{id: Token_LeftParen, literal: "(", line: 9, char: 20},
		{id: Token_TriggerArgs, literal: "std::string<std::vector*>&, std::unique_ptr<void(Foo**&)>", line: 9, char: 21},
		{id: Token_RightParen, literal: ")", line: 9, char: 79},
		{id: Token_Comma, literal: ",", line: 9, char: 80},
		{id: Token_Equal, literal: "=", line: 9, char: 81},
		{id: Token_Comma, literal: ",", line: 9, char: 82},

		{id: Token_Comma, literal: ",", line: 10, char: 2},
		{id: Token_Equal, literal: "=", line: 10, char: 3},
		{id: Token_Boolean, literal: "true", line: 10, char: 8},
		{id: Token_Boolean, literal: "false", line: 10, char: 14},
		{id: Token_Boolean, literal: "true", line: 10, char: 19},

		{id: Token_Integer, literal: "123456789", line: 11, char: 1},
		{id: Token_Integer, literal: "22", line: 11, char: 11},

		{id: Token_Integer, literal: "333", line: 12, char: 1},
		{id: Token_Comma, literal: ",", line: 12, char: 4},
		{id: Token_Integer, literal: "4444", line: 12, char: 6},
		{id: Token_RightParen, literal: ")", line: 12, char: 10},

		{id: Token_KeywordEnter, literal: "enter", line: 13, char: 1},
		{id: Token_KeywordExit, literal: "exit", line: 13, char: 7},

		{id: Token_KeywordCapture, literal: "capture", line: 14, char: 1},
		{id: Token_KeywordCapture, literal: "capture", line: 14, char: 9},
		{id: Token_KeywordIf, literal: "if", line: 14, char: 18},
		{id: Token_LeftBrace, literal: "{", line: 14, char: 19},
		{
			id: Token_Code,
			literal: `
	if (nested) {
		if something
	}
`,
			line: 14,
			char: 20,
		},
		{id: Token_RightBrace, literal: "}", line: 18, char: 1},

		{id: Token_KeywordDo, literal: "do", line: 18, char: 4},
		{id: Token_LeftBrace, literal: "{", line: 18, char: 5},
		{
			id: Token_Code,
			literal: `
	do whatever
`,
			line: 18,
			char: 6,
		},
		{id: Token_RightBrace, literal: "}", line: 20, char: 1},

		{id: Token_EOF, line: 21, char: 0},
	}

	lf, err := files.NewFromData("test_gather_tokens", []byte(input), true)
	require.NoError(t, err)

	s := NewScanner()
	got, errors := s.gatherTokens(lf)
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
		isErrorLine := slices.Contains(errorLines, i)

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
