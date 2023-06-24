// Package gochart_lang holds the logic to partse a statechart from a file/text into our intermediate
// representation, which can then be used by a backend.
package gochart_lang

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

// ScanError is a custom error associated with scanning.
type ScanError struct {
	ErrorToken Token
}

func (se *ScanError) Error() string {
	return fmt.Sprintf("invalid token %q at line %d, char %d",
		se.ErrorToken.literal, se.ErrorToken.line, se.ErrorToken.char)
}

// Scanner is an object capable of taking in a reader and scanning over to create a valid
// representation of our statechart grammar.
type Scanner struct {
	source string

	// We count our characters in runes, which could be more than one byte, because we support utf8,
	// mostly because we're cool.
	// That means that most of our tokenizing logic is in runes, but we still have to keep track of
	// where we are in the byte array because go stores the strings as bytes and we use the |utf8|
	// package to decode the runes on the fly.

	start                  int
	line                   int
	currentByteCount       int
	currentRuneCount       int
	currentRuneCountInLine int
	totalRunes             int

	keywords map[string]TokenIdentifier
}

// NewScanner returns a scanner ready to process input.
// Here all the keywords are defined.
func NewScanner() *Scanner {
	return &Scanner{
		keywords: map[string]TokenIdentifier{
			"statechart": Token_KeywordStatechart,
			"state":      Token_KeywordState,
			"transition": Token_KeywordTransition,
			"trigger":    Token_KeywordTrigger,
		},
	}
}

func (s *Scanner) reset() {
	s.start = 0
	s.currentRuneCount = 0
	s.totalRunes = 0
	s.line = 1
}

func (s *Scanner) Scan(r io.Reader) (*frontend.StatechartData, []error) {
	// Get all the tokens in this input.
	_, errors := s.gatherTokens(r)
	if errors != nil {
		return nil, errors
	}

	return &frontend.StatechartData{}, nil
}

func (s *Scanner) gatherTokens(r io.Reader) ([]*Token, []error) {
	// Forget any previous state before scanning.
	s.reset()

	// Read all the input into the source string.
	if b, err := io.ReadAll(r); err != nil {
		return nil, []error{fmt.Errorf("reading input reader: %w", err)}
	} else {
		s.source = string(b)
	}
	s.totalRunes = utf8.RuneCountInString(s.source)

	var tokens []*Token
	var errors []error

	for !s.atEnd() {
		token, err := s.nextToken()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		tokens = append(tokens, token)
	}

	return tokens, errors
}

func (s *Scanner) nextToken() (*Token, error) {
	token := &Token{}

	// We do this in a loop because there are certain operators that discard a lot of input before
	// detecting where the "next" token should be (eg. comments). So we use this token to return to the
	// scan for the "next token" event.
NEXT_TOKEN_LOOP:
	for {
		// If we're at the end, we simply return a nice EOF.
		if s.atEnd() {
			token.id = Token_EOF
			break
		}

		r, err := s.pop()
		if err != nil {
			return nil, fmt.Errorf("popping rune: %w", err)
		}

		// We check for comments. Comments are one lines with //.
		// We consume the tokens until the end of line and then return to the processing of the next
		// token as if we had never found the comment.
		if r == '/' {
			if err := s.handleComment(); err != nil {
				return nil, fmt.Errorf("handling comment: %w", err)
			}
			// Comment consumed. We go to scan for the next token.
			continue NEXT_TOKEN_LOOP
		}

		// Ignore any whitespace. We consume and go back to processing the node.
		if r == ' ' || r == '\t' || r == '\r' {
			continue NEXT_TOKEN_LOOP
		}

		// New lines add to our current line. Also restart our current character counter.
		if r == '\n' {
			s.newLineFound()
			continue NEXT_TOKEN_LOOP
		}

		// Identifier start with a letter, and could be keywords.
		if unicode.IsLetter(r) {
			if tk, err := s.handleIdentifier(r); err != nil {
				return nil, fmt.Errorf("handling identifier: %w", err)
			} else {
				token = tk
			}
			break
		}

		// String literals are between '"' characters.
		if r == '"' {
			if tk, err := s.handleString(); err != nil {
				return nil, fmt.Errorf("handling string: %w", err)
			} else {
				token = tk
			}
			break
		}

		// Single rune tokens. They get translated
		switch r {
		case '(':
			token.id = Token_LeftParen
			break NEXT_TOKEN_LOOP
		case ')':
			token.id = Token_RightParen
			break NEXT_TOKEN_LOOP
		case '{':
			token.id = Token_LeftBrace
			break NEXT_TOKEN_LOOP
		case '}':
			token.id = Token_RightBrace
			break NEXT_TOKEN_LOOP
		case '[':
			token.id = Token_LeftBracket
			break NEXT_TOKEN_LOOP
		case ']':
			token.id = Token_RightBracket
			break NEXT_TOKEN_LOOP
		}

		return nil, fmt.Errorf("unsupported rune %q", r)
	}

	if token == nil || token.id == Token_Invalid {
		return nil, fmt.Errorf("invalid token parsing. Likely a bug")
	}

	// Fill in the last details of the token if needed, as some tokens come already filled, like the
	// case of string literals.
	if token.line == 0 {
		token.line = s.line
	}
	if token.char == 0 {
		token.char = s.currentRuneCountInLine
	}

	return token, nil
}

// pop reads the current rune and advances the current index
func (s *Scanner) pop() (rune, error) {
	r, err := s.peek()
	if err != nil {
		return utf8.RuneError, fmt.Errorf("peeking: %w", err)
	}

	// Update the indices.
	s.advance(r)

	return r, nil
}

// advances moves the stream forward by a given byte width, which is represents a single rune
// character in the stream.
func (s *Scanner) advance(r rune) {
	s.currentByteCount += utf8.RuneLen(r)
	s.currentRuneCount += 1
	s.currentRuneCountInLine += 1
}

// newLineFound updates the scanner to correctly account for new lines being found.
func (s *Scanner) newLineFound() {
	s.line++
	s.currentRuneCountInLine = 0
}

// peek looks at the current rune pointed by the scanner.
func (s *Scanner) peek() (rune, error) {
	if s.atEnd() {
		return utf8.RuneError, fmt.Errorf("already at input's end")
	}

	// We decode the rune to where we're current pointing in the sring byte array.
	r, _ := utf8.DecodeRuneInString(s.source[s.currentByteCount:])
	if r == utf8.RuneError {
		return utf8.RuneError, fmt.Errorf("invalid rune")
	}

	return r, nil
}

// match returns whether the current rune looked by the stream matches a particular expected rune.
// If the match is successful, it also consumes the rune and advances the stream.
func (s *Scanner) match(expected rune) (bool, error) {
	r, err := s.peek()
	if err != nil {
		return false, fmt.Errorf("peeking: %w", err)
	}

	if r != expected {
		return false, nil
	}

	// We matched, so we advance the stream.
	s.advance(r)
	return true, nil
}

func (s *Scanner) atEnd() bool {
	return s.currentRuneCount >= s.totalRunes
}

// INDIVIDUAL TOKEN CASES --------------------------------------------------------------------------

func (s *Scanner) handleIdentifier(firstRune rune) (*Token, error) {
	// We cache the start of the token to then returning when creating the token.
	startLine := s.line
	startChar := s.currentRuneCountInLine

	// We consume the literal as much as we can. We later see if it matches any keyword.
	runes := []rune{firstRune}
	for !s.atEnd() {
		peek, err := s.peek()
		if err != nil {
			return nil, fmt.Errorf("peeking: %w", err)
		}

		// If it's alphanumerical or _, we consider it part of an identifier. Otherwise, we consider
		// this token terminated.
		isIdentifierChar := unicode.IsLetter(peek) || unicode.IsNumber(peek) || peek == '_'
		if !isIdentifierChar {
			break
		}

		// We pop the peeked character and add it to the current identifier.
		popped, err := s.pop()
		if err != nil {
			return nil, fmt.Errorf("popping: %w", err)
		}
		runes = append(runes, popped)
	}

	identifier := string(runes)
	token := &Token{
		literal: identifier,
		line:    startLine,
		char:    startChar,
	}

	// We check if the identifier is a keyword.
	if id, ok := s.keywords[identifier]; ok {
		token.id = id
	} else {
		token.id = Token_Identifier
	}

	return token, nil
}

// handleString gets called when a '"' character is found. Will create the string token and add it
// to the list.
func (s *Scanner) handleString() (*Token, error) {
	// We cache the start of the token to then returning when creating the token.
	startLine := s.line
	startChar := s.currentRuneCountInLine

	// We consume until we find another '"' token.
	matched := false
	var literal []rune
	for !s.atEnd() {
		peek, err := s.peek()
		if err != nil {
			return nil, fmt.Errorf("peeking: %w", err)
		}

		// In any case, we consume the literal.
		s.advance(peek)

		// If we found the the other '"' character, we stop searching.
		if peek == '"' {
			matched = true
			// Consume the
			break
		}

		// We consume the character and advance.
		literal = append(literal, peek)

		// Special case: if we find a new line, we need to update the current line tracking.
		if peek == '\n' {
			s.newLineFound()
		}
	}

	if !matched {
		return nil, fmt.Errorf("unterminated string literal")
	}

	return &Token{
		id:      Token_StringLiteral,
		literal: string(literal),
		line:    startLine,
		char:    startChar,
	}, nil
}

func (s *Scanner) handleComment() error {
	if isSlash, err := s.match('/'); err != nil {
		return fmt.Errorf("matching '/': %w", err)
	} else if isSlash {
		// We found a comment! It goes until the end of line.
		for !s.atEnd() {
			peek, err := s.peek()
			if err != nil {
				return fmt.Errorf("peeking: %w", err)
			}

			if peek != '\n' && !s.atEnd() {
				s.advance(peek)
				continue
			}

			break
		}
	} else {
		// Single slash doesn't mean anything in our language.
		return fmt.Errorf("single '/' token found")
	}

	return nil
}
