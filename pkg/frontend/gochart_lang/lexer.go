package gochart_lang

import (
	"fmt"
)

// TokenIdentifier represents a single token of our parser.
//go:generate stringer -type=TokenIdentifier
type TokenIdentifier int

const (
	Token_Invalid TokenIdentifier = iota

	// Single character tokens.
	Token_LeftParen    // (
	Token_RightParen   // )
	Token_LeftBrace    // {
	Token_RightBrace   // }
	Token_LeftBracket  // [
	Token_RightBracket // ]
	Token_Slash        // /

	// literals.
	Token_String // "content"
	Token_Number // 123456

	// Keywords
	Token_KeywordStatechart // statechart
	Token_KeywordState      // state
	Token_KeywordTransition // transition
	Token_KeywordTrigger    // trigger

	Token_Identifier // Any string that is not quoted.

	// This is a grab-bag, generally for use when we need to gobble up tokens without really caring on
	// what they mean. This is normally because the frontend doesn't really care about them and the
	// meaning is given in the IR.
	//
	// The contents of triggers arguments is an example of this.
	Token_Other

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
