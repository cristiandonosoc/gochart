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

func NewParser() *Parser {
	return &Parser{
	}
}

// Parse receives a slice of tokens as given by the |Scanner| and tries to match it against the
// gochart_lang grammar.
func (p *Parser) Parse(tokens []*Token) (*ASTNodeRoot, []error) {
	p.tokens = tokens
	p.current = 0

	return p.parseRoot()
}

// ASTNode is the common interface for all the nodes within our Abstract Syntax Tree.
type ASTNode interface {
}

// RULES *******************************************************************************************

// Root --------------------------------------------------------------------------------------------
//
// Root -> Statechart+

type ASTNodeRoot struct {
	statecharts []*ASTNodeStatechart
}

var _ ASTNode = (*ASTNodeRoot)(nil)

// root -> statechart+
func (p *Parser) parseRoot() (*ASTNodeRoot, []error) {
	root := &ASTNodeRoot{}

	var statecharts []*ASTNodeStatechart
	var errors []error

	sc, err := p.parseStatechart()
	if err != nil {
		errors = append(errors, fmt.Errorf("parsing statechart node: %w", err))
		return nil, errors
	}
	statecharts = append(statecharts, sc)

	// See if there are more statechart nodes.
	for {
		if p.peek().id == Token_KeywordStatechart {
			sc, err := p.parseStatechart()
			if err != nil {
				errors = append(errors, fmt.Errorf("parsing statechart node: %w", err))
				return nil, errors
			}
			statecharts = append(statecharts, sc)
		}

		break
	}

	if !p.atEnd() {
		errors = append(errors, fmt.Errorf("expected \"statechart\" expression"))
	}

	root.statecharts = statecharts
	return root, errors
}

// Statechart --------------------------------------------------------------------------------------
//
// Statechart -> "statechart" NAME ( "(" ")" )? "{" "}"

type ASTNodeStatechart struct {
	Name string
}

var _ ASTNode = (*ASTNodeStatechart)(nil)

func (p *Parser) parseStatechart() (*ASTNodeStatechart, error) {
	// statechart keyword.
	if _, err := p.consume(Token_KeywordStatechart); err != nil {
		return nil, err
	}

	// The name.
	nameToken, err := p.consume(Token_Identifier)
	if err != nil {
		return nil, fmt.Errorf("expect name after \"statechart\" token")
	}

	// Optional ( ).
	if p.peekToken(Token_LeftParen) {
		if _, err := p.consume(Token_LeftParen); err != nil {
			return nil, err
		}

		if _, err := p.consume(Token_RightParen); err != nil {
			return nil, err
		}
	}

	// { }
	{
		if _, err := p.consume(Token_LeftBrace); err != nil {
			return nil, err
		}

		if _, err := p.consume(Token_RightBrace); err != nil {
			return nil, err
		}
	}

	return &ASTNodeStatechart{
		Name: nameToken[0].literal,
	}, nil
}

// HELPERS -----------------------------------------------------------------------------------------

// consume is a mix of match and advance, that tries to consume the given list of tokens and gives
// and error if it cannot. Returns the list of consumed tokens.
func (p *Parser) consume(ids ...TokenIdentifier) ([]*Token, error) {
	var tokens []*Token

	for _, id := range ids {
		if p.peek().id != id {
			return nil, fmt.Errorf("expected %q, but got %q", p.peek().id, id)
		}
		tokens = append(tokens, p.advance())
	}

	return tokens, nil
}

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
		return &Token{
			id: Token_EOF,
		}
	}

	return p.tokens[p.current]
}

func (p *Parser) peekToken(id TokenIdentifier) bool {
	return p.peek().id == id
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
	return p.current >= len(p.tokens)
}
