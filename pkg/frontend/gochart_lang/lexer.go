package gochart_lang

import (
	"fmt"
)

// TokenIdentifier represents a single token of our parser.
//
//go:generate stringer -type=TokenIdentifier
type TokenIdentifier int

const (
	Token_Invalid TokenIdentifier = iota

	// Single character tokens.
	Token_LeftParen         // '('
	Token_RightParen        // ')'
	Token_LeftBrace         // '{'
	Token_RightBrace        // '}'
	Token_LeftBracket       // '['
	Token_RightBracket      // ']'
	Token_LeftAngleBracket  // '<'
	Token_RightAngleBracket // '>'
	Token_Equal             // '='
	Token_Comma             // ','

	Token_Boolean // 'true' | 'false'

	// literals.
	Token_String  // "content"
	Token_Integer // 123456

	// Keywords
	Token_KeywordStatechart // 'statechart'
	Token_KeywordState      // 'state'
	Token_KeywordEnter      // 'enter'
	Token_KeywordExit       // 'exit'
	Token_KeywordTransition // 'transition'
	Token_KeywordTrigger    // 'trigger'
	Token_KeywordNull       // 'null'

	// Represents the argument string of a trigger.
	Token_TriggerArgs

	Token_Identifier // Any string that is not quoted.

	Token_EOF
)

// Token represents a single token that the scanner outputs to the parser, which holds associated
// information such as the associated string literal and line/char information.
type Token struct {
	// Id is what kind of token this is.
	id TokenIdentifier

	line int
	char int

	literal string
}

func (t *Token) String() string {
	return fmt.Sprintf("%s %q (%d:%d)", t.id, t.literal, t.line, t.char)
}

func NewToken(id TokenIdentifier) Token {
	return Token{
		id: id,
	}
}

func (t *Token) IsValid() bool {
	return t.id != Token_Invalid
}
