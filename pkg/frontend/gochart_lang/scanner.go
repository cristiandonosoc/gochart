// Package gochart_lang holds the logic to partse a statechart from a file/text into our intermediate
// representation, which can then be used by a backend.
package gochart_lang

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/bradenaw/juniper/xslices"
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
		nextTokens, err := s.nextTokens()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		tokens = append(tokens, nextTokens...)
	}

	return tokens, errors
}

// nextTokens attemps to parse the next set of tokens in the stream. In most cases, it will return
// just one token, but it could be that some tokens come in groups, such as the trigger with the
// arguments.
func (s *Scanner) nextTokens() ([]*Token, error) {
	var tokens []*Token

	// We do this in a loop because there are certain operators that discard a lot of input before
	// detecting where the "next" token should be (eg. comments). So we use this token to return to the
	// scan for the "next token" event.
NEXT_TOKEN_LOOP:
	for {
		// If we're at the end, we simply return a nice EOF.
		if s.atEnd() {
			tokens = append(tokens, s.newToken(Token_EOF, ""))
			break NEXT_TOKEN_LOOP
		}

		// We ignore any whitespace that might come.
		if err := s.eatWhitespace(); err != nil {
			return nil, fmt.Errorf("eating whitespace: %w", err)
		}

		peek, err := s.peek()
		if err != nil {
			return nil, fmt.Errorf("peeking rune: %w", err)
		}

		if token, ok, err := s.handleBoolean(); err != nil {
			return nil, fmt.Errorf("handling boolean: %w", err)
		} else if ok {
			tokens = append(tokens, token)
			break NEXT_TOKEN_LOOP
		}

		if token, ok, err := s.handleInteger(); err != nil {
			return nil, fmt.Errorf("handling integer: %w", err)
		} else if ok {
			tokens = append(tokens, token)
			break NEXT_TOKEN_LOOP
		}

		// We check for comments. Comments are one lines with //.
		// We consume the tokens until the end of line and then return to the processing of the next
		// token as if we had never found the comment.
		if peek == '/' {
			if err := s.handleComment(); err != nil {
				return nil, fmt.Errorf("handling comment: %w", err)
			}
			// Comment consumed. We go to scan for the next token.
			continue NEXT_TOKEN_LOOP
		}

		// New lines add to our current line. Also restart our current character counter.
		if peek == '\n' {
			if err := s.newLineFound(); err != nil {
				return nil, fmt.Errorf("new line handling: %w", err)
			}
			continue NEXT_TOKEN_LOOP
		}

		// Identifier start with a letter, and could be keywords.
		if unicode.IsLetter(peek) {
			if tks, err := s.handleIdentifier(); err != nil {
				return nil, fmt.Errorf("handling identifier: %w", err)
			} else {
				tokens = tks
			}
			break NEXT_TOKEN_LOOP
		}

		// String literals are between '"' characters.
		if peek == '"' {
			if tk, err := s.handleString(); err != nil {
				return nil, fmt.Errorf("handling string: %w", err)
			} else {
				tokens = append(tokens, tk)
			}
			break NEXT_TOKEN_LOOP
		}

		// Single rune tokens. They get translated directly.
		r, err := s.pop()
		if err != nil {
			return nil, fmt.Errorf("popping rune: %w", err)
		}

		switch r {
		case '(':
			tokens = append(tokens, s.newToken(Token_LeftParen, string(r)))
			break NEXT_TOKEN_LOOP
		case ')':
			tokens = append(tokens, s.newToken(Token_RightParen, string(r)))
			break NEXT_TOKEN_LOOP
		case '{':
			tokens = append(tokens, s.newToken(Token_LeftBrace, string(r)))
			break NEXT_TOKEN_LOOP
		case '}':
			tokens = append(tokens, s.newToken(Token_RightBrace, string(r)))
			break NEXT_TOKEN_LOOP
		case '[':
			tokens = append(tokens, s.newToken(Token_LeftBracket, string(r)))
			break NEXT_TOKEN_LOOP
		case ']':
			tokens = append(tokens, s.newToken(Token_RightBracket, string(r)))
			break NEXT_TOKEN_LOOP
		case '=':
			tokens = append(tokens, s.newToken(Token_Equal, string(r)))
			break NEXT_TOKEN_LOOP
		case ',':
			tokens = append(tokens, s.newToken(Token_Comma, string(r)))
			break NEXT_TOKEN_LOOP
		}

		return nil, fmt.Errorf("unrecognized rune %q", r)

	}

	// Verify that the given tokens were correct.
	for _, token := range tokens {
		if token == nil || token.id == Token_Invalid {
			return nil, fmt.Errorf("invalid token parsing. Likely a bug")
		}
	}

	return tokens, nil
}

func (s *Scanner) newToken(id TokenIdentifier, literal string) *Token {
	return &Token{
		id:      id,
		literal: literal,
		line:    s.line,
		char:    s.currentRuneCountInLine,
	}
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
func (s *Scanner) newLineFound() error {
	// We assume the caller called us with a '\n' peek.
	if r, err := s.pop(); err != nil {
		return fmt.Errorf("popping rune: %w", err)
	} else if r != '\n' {
		return fmt.Errorf("new line handling expected a '\n' first rune, got %q", r)
	}

	s.line++
	s.currentRuneCountInLine = 0
	return nil
}

// peek looks at the current rune pointed by the scanner.
func (s *Scanner) peek() (rune, error) {
	if s.atEnd() {
		return utf8.RuneError, fmt.Errorf("already at input's end")
	}

	// We decode the rune to where we're current pointing in the string byte array.
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

// longMatch works like match, expect it looks up more than one rune.
// Either the next runes match the tokens exactly (and will be consumed), or they will not be
// touched at all.
func (s *Scanner) longMatch(want string) (bool, error) {
	// First check that we have that amount of runes to look forward into.
	runeCount := utf8.RuneCountInString(want)
	if s.currentRuneCount+runeCount >= s.totalRunes {
		return false, nil
	}

	offset := s.currentByteCount
	for _, wantRune := range want {
		gotRune, _ := utf8.DecodeRuneInString(s.source[offset:])

		if wantRune != gotRune {
			return false, nil
		}

		offset += utf8.RuneLen(gotRune)
	}

	// Now that we matched all the runes, we need to advance out stream.
	for i := 0; i < runeCount; i++ {
		if _, err := s.pop(); err != nil {
			return false, fmt.Errorf("popping rune: %w", err)
		}
	}

	return true, nil
}

func (s *Scanner) atEnd() bool {
	return s.currentRuneCount >= s.totalRunes
}

// INDIVIDUAL TOKEN CASES --------------------------------------------------------------------------

func (s *Scanner) handleIdentifier() ([]*Token, error) {
	var runes []rune

	// We assume the caller called us with a unicode letter.
	if r, err := s.pop(); err != nil {
		return nil, fmt.Errorf("popping rune: %w", err)
	} else if !unicode.IsLetter(r) {
		return nil, fmt.Errorf("comment handling expected a unicode first rune, got %q", r)
	} else {
		runes = append(runes, r)
	}

	var tokens []*Token

	// We cache the start of the token to then returning when creating the token.
	startLine := s.line
	startChar := s.currentRuneCountInLine

	// We consume the literal as much as we can. We later see if it matches any keyword.
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

	tokens = append(tokens, token)

	// We check for special cases, where we should search for more tokens.
	if token.id == Token_KeywordTrigger {
		triggerTokens, err := s.handleTriggerIdentifier()
		if err != nil {
			return nil, fmt.Errorf("handling trigger identifier: %w", err)
		}
		tokens = append(tokens, triggerTokens...)
	}

	return tokens, nil
}

func (s *Scanner) handleTriggerIdentifier() ([]*Token, error) {
	var tokens []*Token

	if err := s.eatWhitespace(); err != nil {
		return nil, fmt.Errorf("eating whitespace: %w", err)
	}

	// We expect an identifier after the "trigger" keyword.
	ids, err := s.handleIdentifier()
	if err != nil {
		return nil, fmt.Errorf("getting trigger name identifier: %w", err)
	} else if len(ids) != 1 {
		return nil, fmt.Errorf("expected one identifier token after \"trigger\", got %v", ids)
	}
	triggerNameToken := ids[0]
	if triggerNameToken.id != Token_Identifier {
		return nil, fmt.Errorf("expected trigger name token to be Token_Identifier, was %q", triggerNameToken.id)
	}
	tokens = append(tokens, triggerNameToken)

	if err := s.eatWhitespace(); err != nil {
		return nil, fmt.Errorf("eating whitespace: %w", err)
	}

	// Now we need an left paren.
	r, err := s.pop()
	if err != nil {
		return nil, fmt.Errorf("popping rune: %w", err)
	}

	if r != '(' {
		return nil, fmt.Errorf("expected ( after \"update\" keyword, got %q", r)
	}
	tokens = append(tokens, s.newToken(Token_LeftParen, string(r)))

	// Now we can start globbing all strings until we find the matching right paren.
	// We have to be careful to account for any interior parenthesis though.
	startLine := s.line
	startChar := s.currentRuneCountInLine
	matched := false
	openParenCount := 0
	var runes []rune
	for !s.atEnd() {
		r, err := s.pop()
		if err != nil {
			return nil, fmt.Errorf("popping rune: %w", err)
		}

		// If we find a '(', we have to count.
		if r == '(' {
			openParenCount++
			runes = append(runes, r)
			continue
		}

		// If we find a ')', we either decrease the count or we're done.
		if r == ')' {
			if openParenCount > 0 {
				openParenCount--
				runes = append(runes, r)
				continue
			}

			// Here we found the end of the trigger parenthesis.
			matched = true
			break
		}

		// Special case for now. Because it is annoying, we don't allow weird whitespace characters in
		// the trigger arguments. This could change in the future if there is a need for it.
		if r == '\r' || r == '\t' {
			return nil, fmt.Errorf("unallowed rune %q in trigger arguments. This is a concious decision of the parser. If you feel this is correct, please file a bug", r)
		}

		// We also do this for new lines.
		if r == '\n' {
			return nil, fmt.Errorf("unallowed newline in trigger arguments. This is a concious decision of the parser. If you feel this is correct, please file a bug")
		}

		// Normal case, we just glob the literal.
		runes = append(runes, r)
	}

	// If we are out of the loop and didn't match the parenthesis, we have a syntax error.
	if !matched {
		return nil, fmt.Errorf("unmatched trigger arguments parenthesis")
	}

	// We create the tokens.
	tokens = append(tokens, &Token{
		id:      Token_TriggerArgs,
		literal: string(runes),
		line:    startLine,
		char:    startChar,
	})

	tokens = append(tokens, s.newToken(Token_RightParen, ")"))

	return tokens, nil
}

// handleString gets called when a '"' character is found. Will create the string token and add it
// to the list.
func (s *Scanner) handleString() (*Token, error) {
	// We assume the caller called us with a '"' rune.
	if r, err := s.pop(); err != nil {
		return nil, fmt.Errorf("popping rune: %w", err)
	} else if r != '"' {
		return nil, fmt.Errorf("string handling expected a '\"' first rune, got %q", r)
	}

	// We cache the start of the token to then returning when creating the token.
	startLine := s.line
	startChar := s.currentRuneCountInLine

	// We consume until we find another '"' token.
	matched := false
	var runes []rune
	for !s.atEnd() {
		peek, err := s.peek()
		if err != nil {
			return nil, fmt.Errorf("peeking rune: %w", err)
		}

		// If we found the the other '"' character, we stop searching.
		if peek == '"' {
			matched = true
			// Consume the rune.
			if _, err := s.pop(); err != nil {
				return nil, fmt.Errorf("popping rune: %w", err)
			}
			break
		}

		// We consume the character and advance.
		runes = append(runes, peek)

		// Special case: if we find a new line, we need to update the current line tracking.
		if peek == '\n' {
			s.newLineFound()
			continue
		}

		// Otherwise we just pop this rune.
		if _, err := s.pop(); err != nil {
			return nil, fmt.Errorf("popping rune: %w", err)
		}
	}

	if !matched {
		return nil, fmt.Errorf("unterminated string literal")
	}

	return &Token{
		id:      Token_String,
		literal: string(runes),
		line:    startLine,
		char:    startChar,
	}, nil
}

func (s *Scanner) handleBoolean() (*Token, bool, error) {
	if ok, err := s.longMatch("true"); err != nil {
		return nil, false, fmt.Errorf("long matching \"true\": %w", err)
	} else if ok {
		return s.newToken(Token_Boolean, "true"), true, nil
	}

	if ok, err := s.longMatch("false"); err != nil {
		return nil, false, fmt.Errorf("long matching \"false\": %w", err)
	} else if ok {
		return s.newToken(Token_Boolean, "false"), true, nil
	}

	// If we didn't match any boolean value, we simply ignore.
	return nil, false, nil
}

func (s *Scanner) handleInteger() (*Token, bool, error) {
	var runes []rune

	// We go along searching for numbers.
	for !s.atEnd() {
		peek, err := s.peek()
		if err != nil {
			return nil, false, fmt.Errorf("peeking rune: %w", err)
		}

		// We stop iterating if we found a number.
		if !unicode.IsDigit(peek) {
			break
		}

		// Otherwise we advance and continue scanning.
		r, err := s.pop()
		if err != nil {
			return nil, false, fmt.Errorf("popping rune: %w", err)
		}
		runes = append(runes, r)
	}

	// If we didn't find any digits, we simply say we didn't found a number.
	if len(runes) == 0 {
		return nil, false, nil
	}

	// We must be careful, as numbers should be separated by a certain amount of characters only.
	peek, err := s.peek()
	if err != nil {
		return nil, false, fmt.Errorf("peeking rune: %w", err)
	}

	acceptable := []rune{' ', '\n', ','}
	if xslices.Count(acceptable, peek) == 0 {
		return nil, false, fmt.Errorf("numbers must be separated by %v, got %q", acceptable, peek)
	}

	return s.newToken(Token_Integer, string(runes)), true, nil
}

func (s *Scanner) handleComment() error {
	// We assume the caller called us with a '/' peek.
	if r, err := s.pop(); err != nil {
		return fmt.Errorf("popping rune: %w", err)
	} else if r != '/' {
		return fmt.Errorf("comment handling expected a '/' first rune, got %q", r)
	}

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

// eatWhitespace is a helper function meant to simply ignore whitespace characters.
func (s *Scanner) eatWhitespace() error {
	for !s.atEnd() {
		peek, err := s.peek()
		if err != nil {
			return fmt.Errorf("peeking rune: %w", err)
		}

		// Ignore any whitespace. We simply consume.
		if isRuneWhitespace(peek) {
			if _, err := s.pop(); err != nil {
				return fmt.Errorf("popping rune: %w", err)
			}

			continue
		}

		break
	}

	return nil
}

func isRuneWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}
