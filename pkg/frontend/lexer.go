package frontend

import ()

// TokenIdentifier represents a single token of our parser.
type TokenIdentifier int64

const (
	// Single character tokens.
	Token_LeftParen TokenIdentifier = iota
	Token_RightParen
	Token_LeftBrace
	Token_RightBrace
	Token_LeftBracket
	Token_RightBracket

	// literals.
	Token_String
	Token_Number

	// Keywords
	Token_Statechart
	Token_State
	Token_Transition
	Token_Trigger

	Token_EOF
)
