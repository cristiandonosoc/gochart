package gochart_lang

import (
	"fmt"
)

// Parser is an object capable of taking tokens of the gochart_lang language and parse it against
// the grammer, returning an Abstract Syntax Tree (AST).
type Parser struct {
	tokens  []*Token
	current int
}

// Parse receives a slice of tokens as given by the |Scanner| and tries to match it against the
// gochart_lang grammar.
func (p *Parser) Parse(tokens []*Token) error {
	p.tokens = tokens
	p.current = 0

	return fmt.Errorf("NOT IMPLEMENTED")
}

// RULES -------------------------------------------------------------------------------------------

type ASTNode interface {
	Print()
}

var _ ASTNode = (*ASTNodeRoot)(nil)

type ASTNodeRoot struct {
	statecharts []*ASTNodeStatechart
}

// root -> statechart+
func (p *Parser) parseRoot() (*ASTNodeRoot, error) {
	root := &ASTNodeRoot{}

	for {
		sc, ok, err := p.parseStatechart()
		if err != nil {
			return nil, fmt.Errorf("parsing statechart node: %w", err)
		}

		// If there are no more statecharts to parse, we're done.
		if !ok {
			break
		}

		// Otherwise we add it to the list and try to parse the next one.
		root.statecharts = append(root.statecharts, sc)
	}

	return root, nil
}

var _ ASTNode = (*ASTNodeStatechart)(nil)

type ASTNodeStatechart struct {
}

// statechart -> STATECHART LEFT_BRACE RIGHT_BRACE
func (p *Parser) parseStatechart() (*ASTNodeStatechart, bool, error) {
	return nil, false, fmt.Errorf("NOT IMPLEMENTED")
}

// HELPERS -----------------------------------------------------------------------------------------

// match returns whether the current token matches any of the particular tokens being asked for.
// If the match is successful, it also consumes the token and advances the stream.
func (p *Parser) match(ids ...TokenIdentifier) bool {
	if p.atEnd() {
		return false
	}

	peek := p.peek()
	for _, id := range ids {
		// If the peeked token matches the type, we advance the stream forward.
		if peek.id == id {
			p.advance()
			return true
		}
	}

	return false
}

// peek looks at the current token pointed by the stream.
func (p *Parser) peek() *Token {
	if p.atEnd() {
		panic("already at input's end")
	}

	return p.tokens[p.current]
}

// prev returns the token behind the current (behind peek).
// Returns the token just advances over.
func (p *Parser) prev() *Token {
	if p.current == 0 {
		panic("cannot ask for prev of the first token")
	}

	return p.tokens[p.current-1]
}

// advance moves the stream one token forward.
func (p *Parser) advance() *Token {
	if !p.atEnd() {
		p.current += 1
	}
	return p.prev()
}

// atEnd returns whether the parser has already consumed all the tokens in the stream and is
// currently at the last token: EOF.
func (p *Parser) atEnd() bool {
	return p.peek().id == Token_EOF
}

// Printing ----------------------------------------------------------------------------------------

func (n *ASTNodeRoot) Print() {
	for _, sc := range n.statecharts {
		sc.Print()
	}
}

func (n *ASTNodeStatechart) Print() {
	fmt.Print("statechart { }")
}
