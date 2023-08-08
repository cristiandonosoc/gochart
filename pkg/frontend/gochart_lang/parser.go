package gochart_lang

import (
	"fmt"
	"strconv"
)

// Parser is an object capable of taking tokens of the gochart_lang language and parse it against
// the grammer, returning an Abstract Syntax Tree (AST).
type Parser struct {
	tokens  []*Token
	current int
}

func NewParser() *Parser {
	return &Parser{}
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
		errors = append(errors, fmt.Errorf("expected \"statechart\" expression, got %q", p.peek()))
	}

	root.statecharts = statecharts
	return root, errors
}

// Statechart --------------------------------------------------------------------------------------
//
// Statechart -> 'statechart' NAME ( '(' ')' )? '{' '}'

type ASTNodeStatechart struct {
	Name     string
	Triggers []*ASTNodeTrigger
	States   []*ASTNodeState
}

var _ ASTNode = (*ASTNodeStatechart)(nil)

func (p *Parser) parseStatechart() (*ASTNodeStatechart, error) {
	// statechart keyword.
	if _, err := p.consumeToken(Token_KeywordStatechart); err != nil {
		return nil, err
	}

	// The name.
	name, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, fmt.Errorf("expect name after \"statechart\" token")
	}

	// Optional ( ).
	if p.peekToken(Token_LeftParen) {
		if _, err := p.consumeToken(Token_LeftParen); err != nil {
			return nil, err
		}

		if _, err := p.consumeToken(Token_RightParen); err != nil {
			return nil, err
		}
	}

	// Statechart body.
	var triggers []*ASTNodeTrigger
	var states []*ASTNodeState
	{
		// {
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, err
		}

		// We start a loop searching for things within the body.
		// It closes when we find a right paren.
		for !p.peekToken(Token_RightBrace) {
			// Check if there is a trigger.
			if trigger, ok, err := p.parseTrigger(); err != nil {
				return nil, fmt.Errorf("statechart %q: parsing trigger: %w", name, err)
			} else if ok {
				triggers = append(triggers, trigger)
			}

			// Check if there is a state.
			if state, ok, err := p.parseState(nil); err != nil {
				return nil, fmt.Errorf("statechart %q: parsing state: %w", name, err)
			} else if ok {
				states = append(states, state)
			}

		}

		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, err
		}
	}

	return &ASTNodeStatechart{
		Name:     name,
		Triggers: triggers,
		States:   states,
	}, nil
}

// Trigger -----------------------------------------------------------------------------------------
//
// Trigger -> 'trigger' NAME '(' ARG_TYPE+ ARG_NAME (',' ARG_TYPE+ ARG_NAME)* ')'

type ASTNodeTrigger struct {
	Name       string
	ArgsString string
}

var _ ASTNode = (*ASTNodeTrigger)(nil)

func (p *Parser) parseTrigger() (*ASTNodeTrigger, bool, error) {
	// If we found the trigger keyword, we found a trigger.
	if _, ok := p.match(Token_KeywordTrigger); !ok {
		return nil, false, nil
	}

	// The name.
	name, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, false, fmt.Errorf("expect name after \"trigger\" token")
	}

	// First paren.
	if _, err := p.consumeToken(Token_LeftParen); err != nil {
		return nil, false, err
	}

	// Optionally we could have arguments.
	var args string
	if p.peek().id == Token_TriggerArgs {
		argsStr, err := p.consumeTokenLiteral(Token_TriggerArgs)
		if err != nil {
			return nil, false, fmt.Errorf("consuming args token: %w", err)
		}
		args = argsStr
	}

	// Last paren.
	if _, err := p.consumeToken(Token_RightParen); err != nil {
		return nil, false, err
	}

	return &ASTNodeTrigger{
		Name:       name,
		ArgsString: args,
	}, true, nil
}

// State -------------------------------------------------------------------------------------------
//
// State -> 'state' NAME ( '(' ( StateOption ( ',' StateOption )* ')' ) StateBody
// State -> 'state' NAME 'StateOptions? StateBody

type ASTNodeState struct {
	Name           string
	ParentName     string
	Options        []*ASTNodeStateOption
	EnterReactions []*ASTNodeStateReaction
	ExitReactions  []*ASTNodeStateReaction
	Transitions    []*ASTNodeStateTransition
	Children       []*ASTNodeState
}

var _ ASTNode = (*ASTNodeState)(nil)

func (p *Parser) parseState(parent *ASTNodeState) (*ASTNodeState, bool, error) {
	// If we found  the 'state' keyword, we parse it.
	if _, ok := p.match(Token_KeywordState); !ok {
		return nil, false, nil
	}

	// The name.
	name, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, false, fmt.Errorf("expect name after \"trigger\" token")
	}

	// The options.
	var options []*ASTNodeStateOption
	if opts, ok, err := p.parseStateOptions(); err != nil {
		return nil, false, fmt.Errorf("state %q: parsing options: %w", name, err)
	} else if ok {
		options = opts
	}

	// Create the state to be passed to the children.
	var parentName string
	if parent != nil {
		parentName = parent.Name
	}

	state := &ASTNodeState{
		Name:       name,
		Options:    options,
		ParentName: parentName,
	}

	// The body.
	var enterReactions []*ASTNodeStateReaction
	var exitReactions []*ASTNodeStateReaction
	var transitions []*ASTNodeStateTransition
	var children []*ASTNodeState
	{
		// The body opening '{'
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, false, fmt.Errorf("state %q: parsing state body: %w", name, err)
		}

		// We loop searching for things that can be in the body.
		for !p.atEnd() {
			// If we peek a right brace, we know the body is closed.
			if p.peekToken(Token_RightBrace) {
				break
			}

			// See if we found an enter reaction.
			if reaction, ok, err := p.parseStateEnterReaction(); err != nil {
				return nil, false, fmt.Errorf("state %q: parsing enter reaction: %w", name, err)
			} else if ok {
				enterReactions = append(enterReactions, reaction)
			}

			// See if we found an exit reaction.
			if reaction, ok, err := p.parseStateExitReaction(); err != nil {
				return nil, false, fmt.Errorf("state %q: parsing exit reaction: %w", name, err)
			} else if ok {
				exitReactions = append(exitReactions, reaction)
			}

			// See if we found a transition.
			if transition, ok, err := p.parseStateTransition(); err != nil {
				return nil, false, fmt.Errorf("state %q: parsing transition: %w", name, err)
			} else if ok {
				transitions = append(transitions, transition)
			}

			// See if we found a child state.
			if child, ok, err := p.parseState(state); err != nil {
				return nil, false, fmt.Errorf("state %q: parse child state: %w", name, err)
			} else if ok {
				children = append(children, child)
			}
		}

		// The body closing '}'
		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, false, fmt.Errorf("state %q: closing state body: %w", name, err)
		}
	}

	// Fill in the rest of the state.
	state.EnterReactions = enterReactions
	state.ExitReactions = exitReactions
	state.Transitions = transitions
	state.Children = children

	return state, true, nil
}

// StateOptions ------------------------------------------------------------------------------------
//
// StateOptions -> ( '(' StateOption ( ',' StateOption )* ')' )?

func (p *Parser) parseStateOptions() ([]*ASTNodeStateOption, bool, error) {
	// If we have no left paren, then there is no options.
	if _, ok := p.match(Token_LeftParen); !ok {
		return nil, false, nil
	}

	var options []*ASTNodeStateOption

	// Parse the first option.
	option, err := p.parseStateOption()
	if err != nil {
		return nil, false, fmt.Errorf("parsing first option: %w", err)
	}
	options = append(options, option)

	// See if we have more options.
	for !p.atEnd() {
		// A ',' indicates that we are going to have another option.
		if _, ok := p.match(Token_Comma); !ok {
			break
		}

		// Parse the next option.
		option, err := p.parseStateOption()
		if err != nil {
			return nil, false, fmt.Errorf("parsing option: %w", err)
		}
		options = append(options, option)
	}

	// Close the options ')'
	if _, err := p.consumeToken(Token_RightParen); err != nil {
		return nil, false, fmt.Errorf("unclosed state options parenthesis: %w", err)
	}

	return options, true, nil
}

// StateOption -> NAME ( '=' ( IDENTIFIER | STRING | NUMBER | true | false ) )?

type StateOptionValueType int

const (
	ValueType_None StateOptionValueType = iota
	ValueType_String
	ValueType_Integer
	ValueType_Boolean
)

type ASTNodeStateOption struct {
	Name      string
	ValueType StateOptionValueType
	Value     string
}

var _ ASTNode = (*ASTNodeStateOption)(nil)

func (p *Parser) parseStateOption() (*ASTNodeStateOption, error) {
	// The name.
	name, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, fmt.Errorf("expect name after \"trigger\" token")
	}

	// See if there is an optional value.
	valueType := ValueType_None
	var value string
	if _, ok := p.match(Token_Equal); ok {
		if token, ok := p.match(Token_Identifier); ok {
			valueType = ValueType_String
			value = token.literal
		} else if token, ok := p.match(Token_String); ok {
			valueType = ValueType_String
			value = token.literal
		} else if token, ok := p.match(Token_Integer); ok {
			_, err := strconv.Atoi(token.literal)
			if err != nil {
				return nil, fmt.Errorf("could not parse option value %q to int: %w", token.literal, err)
			}
			valueType = ValueType_Integer
			value = token.literal
		} else if token, ok := p.match(Token_Boolean); ok {
			_, err := strconv.ParseBool(token.literal)
			if err != nil {
				return nil, fmt.Errorf("could not parse option value %q to boolean: %w", token.literal, err)
			}
			valueType = ValueType_Boolean
			value = token.literal
		} else {
			return nil, fmt.Errorf("unexpected token on state option. Expected Identifier, String, Number or Boolean, got: %s", token.String())
		}
	}

	return &ASTNodeStateOption{
		Name:      name,
		ValueType: valueType,
		Value:     value,
	}, nil
}

// StateReaction -----------------------------------------------------------------------------------
//
// StateReaction -> 'enter' ( '(' IDENTIFIER ')' )?
//               ->  'exit' ( '(' IDENTIFIER ')' )?

type ASTNodeStateReaction struct {
	Trigger string
}

var _ ASTNode = (*ASTNodeStateReaction)(nil)

func (p *Parser) parseStateEnterReaction() (*ASTNodeStateReaction, bool, error) {
	return p.parseStateReaction(Token_KeywordEnter)
}

func (p *Parser) parseStateExitReaction() (*ASTNodeStateReaction, bool, error) {
	return p.parseStateReaction(Token_KeywordExit)
}

func (p *Parser) parseStateReaction(keywordToken TokenIdentifier) (*ASTNodeStateReaction, bool, error) {
	// Check if we found the keyword we were expecting.
	if _, ok := p.match(keywordToken); !ok {
		return nil, false, nil
	}

	// We check to see if the user indicated a trigger associated with the reaction.
	var triggerName string
	if triggerToken, ok := p.match(Token_Identifier); ok {
		triggerName = triggerToken.literal
	}

	return &ASTNodeStateReaction{
		Trigger: triggerName,
	}, true, nil
}

// StateTransition ---------------------------------------------------------------------------------
//
// StateTransition -> 'transition' ( 'null' | IDENTIFIER ) > IDENTIFIER

type ASTNodeStateTransition struct {
	Trigger     string
	TargetState string
}

var _ ASTNode = (*ASTNodeState)(nil)

func (p *Parser) parseStateTransition() (*ASTNodeStateTransition, bool, error) {
	// Check if there is a 'transition' keyword.
	if _, ok := p.match(Token_KeywordTransition); !ok {
		return nil, false, nil
	}

	// Check if it's a null transition. If not, we expect an identifier.
	var triggerName string
	if _, ok := p.match(Token_KeywordNull); !ok {
		triggerToken, err := p.consumeToken(Token_Identifier)
		if err != nil {
			return nil, false, fmt.Errorf("expected null or identifier token: %w", err)
		}
		triggerName = triggerToken.literal
	}

	// The right angle bracket.
	if _, err := p.consumeToken(Token_RightAngleBracket); err != nil {
		return nil, false, fmt.Errorf("after transition trigger identifier: %w", err)
	}

	// We now expect the trigger identifier.
	stateToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, false, fmt.Errorf("expect identifier after >: %w", err)
	}

	return &ASTNodeStateTransition{
		Trigger:     triggerName,
		TargetState: stateToken.literal,
	}, true, nil
}

// HELPERS -----------------------------------------------------------------------------------------

// consume is a mix of match and advance, that tries to consume the given list of tokens and gives
// and error if it cannot. Returns the list of consumed tokens.
func (p *Parser) consumeTokens(ids ...TokenIdentifier) ([]*Token, error) {
	var tokens []*Token

	for _, id := range ids {
		if !p.peekToken(id) {
			return nil, fmt.Errorf("expected %q, but got %q", id, p.peek().id)
		}
		tokens = append(tokens, p.advance())
	}

	return tokens, nil
}

// consumeToken is a variation of consume that only returns one token for convenience.
func (p *Parser) consumeToken(id TokenIdentifier) (*Token, error) {
	tokens, err := p.consumeTokens(id)
	if err != nil {
		return nil, err
	}

	if len(tokens) != 1 {
		return nil, fmt.Errorf("expecting to consume %q to return only one token, got %v", id, tokens)
	}

	return tokens[0], nil
}

// consumeTokenLiteral is an even more refived convenience variation that tries to consume a token
// and returns the literal instead of the token. This is useful for identifier nodes that we really
// care the literal.
func (p *Parser) consumeTokenLiteral(id TokenIdentifier) (string, error) {
	token, err := p.consumeToken(id)
	if err != nil {
		return "", err
	}

	return token.literal, nil
}

// match returns whether the current token matches any of the particular tokens being asked for.
// If the match is successful, it also consumes the token and advances the stream.
func (p *Parser) match(id TokenIdentifier) (*Token, bool) {
	if p.atEnd() {
		return nil, false
	}

	// If the peeked token matches the type, we advance the stream forward.
	if p.peekToken(id) {
		token := p.advance()
		return token, true
	}

	return nil, false
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
	if p.current == len(p.tokens)-1 {
		return p.tokens[p.current].id == Token_EOF
	}
	return p.current >= len(p.tokens)
}
