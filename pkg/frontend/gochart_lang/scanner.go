// Package gochart_lang holds the logic to partse a statechart from a file/text into our intermediate
// representation, which can then be used by a backend.
package gochart_lang

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/cristiandonosoc/gochart/pkg/ir"
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
}

// NewScanner returns a scanner ready to process input.
func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) reset() {
	s.start = 0
	s.currentRuneCount = 0
	s.totalRunes = 0
	s.line = 1
}

func (s *Scanner) Scan(r io.Reader) (*ir.Statechart, []error) {
	// Forget any previous state before scanning.
	s.reset()

	// Read all the input into the source string.
	if b, err := io.ReadAll(r); err != nil {
		return nil, []error{fmt.Errorf("reading input reader: %w", err)}
	} else {
		s.source = string(b)
	}
	s.totalRunes = utf8.RuneCountInString(s.source)

	// Get all the tokens in this input.
	_, errors := s.gatherTokens()
	if errors != nil {
		return nil, errors
	}

	return &ir.Statechart{}, nil
}

func (s *Scanner) gatherTokens() ([]*Token, []error) {
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
	id := Token_Invalid

	// We do this in a loop because there are certain operators that discard a lot of input before
	// detecting where the "next" token should be (eg. comments). So we use this token to return to the
	// scan for the "next token" event.
NEXT_TOKEN_LOOP:
	for {
		// If we're at the end, we simply return a nice EOF.
		if s.atEnd() {
			id = Token_EOF
			break
		}

		r, err := s.pop()
		if err != nil {
			return nil, fmt.Errorf("popping rune: %w", err)
		}

		// Go over one-rune tokens.
		switch r {
		case '(':
			id = Token_LeftParen
			break
		case ')':
			id = Token_RightParen
			break
		case '{':
			id = Token_LeftBrace
			break
		case '}':
			id = Token_RightBrace
			break
		case '[':
			id = Token_LeftBracket
			break
		case ']':
			id = Token_RightBracket
			break
		// We ignore any whitespace.
		case ' ', '\t', '\r':
			continue NEXT_TOKEN_LOOP
			// New lines add to our current line. Also restart our current character counter.
		case '\n':
			s.line++
			s.currentRuneCountInLine = 0
			continue NEXT_TOKEN_LOOP
		// We check for comments.
		case '/':
			if isSlash, err := s.match('/'); err != nil {
				return nil, fmt.Errorf("matchingn '/': %w", err)
			} else if isSlash {
				// We found a comment! It goes until the end of line.
				for !s.atEnd() {
					peek, err := s.peek()
					if err != nil {
						return nil, fmt.Errorf("peeking: %w", err)
					}

					if peek != '\n' && !s.atEnd() {
						s.advance(peek)
						continue
					}

					// We go to scan for the next token.
					continue NEXT_TOKEN_LOOP
				}
			} else {
				// Single slash doesn't mean anything in our language.
				return nil, fmt.Errorf("single '/' token found")
			}
		}

		if id == Token_Invalid {
			return nil, fmt.Errorf("unknown rune %q", r)
		}

		// We found a valid token, so we stop searching.
		break
	}

	return &Token{
		id:   id,
		line: s.line,
		char: s.currentRuneCountInLine,
	}, nil
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
