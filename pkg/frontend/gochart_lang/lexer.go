package gochart_lang

import (
	"fmt"
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

// TokenIdentifier represents a single token of our parser.
// IMPORTANT: If you add another token, remember to add it to the String() function at the end
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
	Token_KeywordIf         // 'if'
	Token_KeywordDo         // 'do'
	Token_KeywordStatechart // 'statechart'
	Token_KeywordState      // 'state'
	Token_KeywordEnter      // 'enter'
	Token_KeywordExit       // 'exit'
	Token_KeywordTransition // 'transition'
	Token_KeywordTrigger    // 'trigger'
	Token_KeywordCapture    // 'capture'
	Token_KeywordNull       // 'null'

	// Represents the argument string of a trigger.
	Token_TriggerArgs

	Token_Identifier // Any string that is not quoted.

	Token_EOF
)

func (t TokenIdentifier) String() string {
	switch id := t; id {
	case Token_Invalid:
		return "<invalid>"
	case Token_LeftParen:
		return "("
	case Token_RightParen:
		return ")"
	case Token_LeftBrace:
		return "{"
	case Token_RightBrace:
		return "}"
	case Token_LeftBracket:
		return "["
	case Token_RightBracket:
		return "]"
	case Token_LeftAngleBracket:
		return "<"
	case Token_RightAngleBracket:
		return ">"
	case Token_Equal:
		return "="
	case Token_Comma:
		return ","
	case Token_Boolean:
		return "<boolean>"
	case Token_String:
		return "<string>"
	case Token_Integer:
		return "<integer>"
	case Token_KeywordIf:
		return "if"
	case Token_KeywordDo:
		return "do"
	case Token_KeywordStatechart:
		return "statechart"
	case Token_KeywordState:
		return "state"
	case Token_KeywordEnter:
		return "enter"
	case Token_KeywordExit:
		return "exit"
	case Token_KeywordTransition:
		return "transition"
	case Token_KeywordTrigger:
		return "trigger"
	case Token_KeywordCapture:
		return "capture"
	case Token_KeywordNull:
		return "null"
	case Token_TriggerArgs:
		return "<args>"
	case Token_Identifier:
		return "<identifier>"
	case Token_EOF:
		return "<EOF>"
	default:
		panic(fmt.Sprintf("Unsupported token id \"%v\". Most likely someone forgot to update the switch when adding a new token.", id))
	}

}
