package gochart_lang

import ()

// TokenIdentifier represents a single token of our parser.
type TokenIdentifier int64

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
	Token_StringLiteral // "content"
	Token_Number        // 123456

	// Keywords
	Token_Statechart // statechart
	Token_State      // state
	Token_Transition // transition
	Token_Trigger    // trigger

	Token_EOF
)

type Token struct {
	// Id is what kind of token this is.
	id TokenIdentifier

	line int
	char int

	literal string
}

func NewToken(id TokenIdentifier) Token {
	return Token{
		id: id,
	}
}

func (t *Token) valid() bool {
	return t.id != Token_Invalid
}
