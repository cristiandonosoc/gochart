package gochart_lang

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/cristiandonosoc/gochart/internal/errors"
	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

// Parser is an object capable of taking tokens of the gochart_lang language and parse it against
// the grammer, returning an Abstract Syntax Tree (AST).
type Parser struct {
	tokens  []*Token
	current int

	currentStatechart string
	currentState      string

	errors []error

	options *Options
	lf      *files.LoadedFile
}

func NewParser(lf *files.LoadedFile, options *Options) *Parser {
	return &Parser{
		options: options,
		lf:      lf,
	}
}

// Parse receives a slice of tokens as given by the |Scanner| and tries to match it against the
// gochart_lang grammar.
func (p *Parser) Parse(tokens []*Token) (*ASTNodeRoot, []error) {
	p.tokens = tokens
	p.current = 0
	p.errors = nil

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
	p.debuglog("parsing root node")

	root := &ASTNodeRoot{}

	var statecharts []*ASTNodeStatechart

	sc, err := p.parseStatechart()
	if err != nil {
		if len(p.errors) == 0 {
			panic(fmt.Sprintf("unaccounted parser error on err: %v", err.Error()))
		}
		return nil, p.errors
	}
	statecharts = append(statecharts, sc)

	// See if there are more statechart nodes.
	// TODO(cdc): See how we can make the parser recover from an error.
	for {
		if p.peek().id == Token_KeywordStatechart {
			sc, err := p.parseStatechart()
			if err != nil {
				if len(p.errors) == 0 {
					panic(fmt.Sprintf("unaccounted parser error on err: %v", err.Error()))
				}
				return nil, p.errors
			}
			statecharts = append(statecharts, sc)
		}

		break
	}

	if !p.atEnd() {
		p.pushErrorf(0, "expected \"statechart\" keyword, got %q", p.peek())
	}

	root.statecharts = statecharts
	return root, p.errors
}

// Statechart --------------------------------------------------------------------------------------
//
// Statechart -> 'statechart' NAME ( '(' ')' )? '{' StatechartBodyStmt+ '}'

type ASTNodeStatechart struct {
	Name     string
	Triggers []*ASTNodeTrigger
	States   []*ASTNodeState

	token *Token
}

var _ ASTNode = (*ASTNodeStatechart)(nil)

func (p *Parser) parseStatechart() (*ASTNodeStatechart, error) {
	p.debuglog("parsing statechart node")

	// statechart keyword.
	baseToken, err := p.consumeToken(Token_KeywordStatechart)
	if err != nil {
		return nil, p.pushErrorf(0, err.Error())
	}

	// The name.
	name, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"statechart\" keyword: %v", err.Error())
	}

	// Debug tracking.
	{
		prevStatechart := p.currentStatechart
		p.currentStatechart = name
		defer func() {
			p.currentStatechart = prevStatechart
		}()
	}

	// Optional ( ).
	if p.peekToken(Token_LeftParen) {
		if _, err := p.consumeToken(Token_LeftParen); err != nil {
			return nil, p.pushErrorf(0, err.Error())
		}

		if _, err := p.consumeToken(Token_RightParen); err != nil {
			return nil, p.pushErrorf(0, err.Error())
		}
	}

	sc := &ASTNodeStatechart{
		Name:  name,
		token: baseToken,
	}

	{
		// {
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, p.pushErrorf(0, err.Error())
		}

		bodyStmtsFound := 0

		for !p.atEnd() {
			ok, err := p.parseStatechartBodyStmt(sc)
			if err != nil {
				return nil, fmt.Errorf("parsing statechart body statement: %w", err)
			}

			if ok {
				bodyStmtsFound++
				continue

			}

			break
		}

		if bodyStmtsFound == 0 {
			return nil, p.pushErrorf(0, "empty statechart found, need at least one statement")
		}

		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, p.pushErrorf(0, err.Error())
		}
	}

	return sc, nil
}

// Options ------------------------------------------------------------------------------------
//
// Options -> ( '(' Option ( ',' Option )* ')' )?

func (p *Parser) parseOptions() ([]*ASTNodeOption, bool, error) {
	p.debuglog("parsing state options node")

	// If we have no left paren, then there is no options.
	if _, ok := p.match(Token_LeftParen); !ok {
		p.debuglog("-> -> no options %q found. Skipping.", Token_LeftParen)
		return nil, false, nil
	}

	var options []*ASTNodeOption

	// Parse the first option.
	option, err := p.parseOption()
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
		option, err := p.parseOption()
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

// Option -> NAME ( '=' ( IDENTIFIER | STRING | NUMBER | true | false ) )?

type ASTNodeOption struct {
	Name      string
	ValueKind output.OptionValueKind
	Value     string

	token *Token
}

var _ ASTNode = (*ASTNodeOption)(nil)

func (p *Parser) parseOption() (*ASTNodeOption, error) {
	p.debuglog("-> parsing state option")

	// The name.
	nameToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"trigger\" keyword")
	}

	// See if there is an optional value.
	// If not, by default this is a "true" boolean option.
	valueKind := output.OptionValueKind_Boolean
	value := "true"
	if _, ok := p.match(Token_Equal); ok {
		if token, ok := p.match(Token_Identifier); ok {
			valueKind = output.OptionValueKind_String
			value = token.literal
		} else if token, ok := p.match(Token_String); ok {
			valueKind = output.OptionValueKind_String
			value = token.literal
		} else if token, ok := p.match(Token_Integer); ok {
			_, err := strconv.Atoi(token.literal)
			if err != nil {
				return nil, fmt.Errorf("could not parse option value %q to int: %w", token.literal, err)
			}
			valueKind = output.OptionValueKind_Integer
			value = token.literal
		} else if token, ok := p.match(Token_Boolean); ok {
			_, err := strconv.ParseBool(token.literal)
			if err != nil {
				return nil, fmt.Errorf("could not parse option value %q to boolean: %w", token.literal, err)
			}
			valueKind = output.OptionValueKind_Boolean
			value = token.literal
		} else {
			return nil, p.pushErrorf(0, "unexpected token on state option. Expected Identifier, String, Number or Boolean, got: %s", token.String())
		}
	}

	option := &ASTNodeOption{
		Name:      nameToken.literal,
		ValueKind: valueKind,
		Value:     value,
		token:     nameToken,
	}

	p.debuglog("-> -> parsed option: %+v", option)

	return option, nil
}

// StatechartBodyStmt ------------------------------------------------------------------------------
//
// StatechartBodyStmt -> Trigger
//                    -> State

// When parsing the statechart body, there are a set of keywords that are allowed and indicate what
// it is next (trigger, state, etc.).
// These are the only things allowed in as a statechart body statement.
var expectedStatechartBodyStmtKeywords = []TokenIdentifier{
	Token_KeywordTrigger,
	Token_KeywordState,
}

func (p *Parser) parseStatechartBodyStmt(sc *ASTNodeStatechart) (bool, error) {
	// If we're closing the brace, we simply exit the statement.
	if peek := p.peek(); peek.id == Token_RightBrace {
		return false, nil
	}

	stmtToken := p.advance()

	if !slices.Contains(expectedStatechartBodyStmtKeywords, stmtToken.id) {
		return false, p.pushErrorf(-1, "expected token to be in %v, got %q", expectedStatechartBodyStmtKeywords, stmtToken.id)
	}

	switch stmtToken.id {
	case Token_KeywordTrigger:
		trigger, err := p.parseTrigger(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing trigger: %w", err)
		}
		sc.Triggers = append(sc.Triggers, trigger)
		return true, nil

	case Token_KeywordState:
		p.debuglog("parsing root state")
		state, err := p.parseState(nil, stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing root state: %w", err)
		}
		sc.States = append(sc.States, state)
		return true, nil
	}

	// TODO(cdc): Mark this is parser internal error.
	return false, p.pushInternalErrorf(0, "expected token to be in %v, got %q", expectedStatechartBodyStmtKeywords, stmtToken.id)
}

// Trigger -----------------------------------------------------------------------------------------
//
// Trigger -> 'trigger' NAME '(' ARG_TYPE+ ARG_NAME (',' ARG_TYPE+ ARG_NAME)* ')'

type ASTNodeTrigger struct {
	Name       string
	ArgsString string

	token *Token
}

var _ ASTNode = (*ASTNodeTrigger)(nil)

func (p *Parser) parseTrigger(baseToken *Token) (*ASTNodeTrigger, error) {
	p.debuglog("parsing trigger node")

	if baseToken.id != Token_KeywordTrigger {
		return nil, p.pushInternalErrorf(0, "expected 'trigger' keyword, got: %q", baseToken.id)
	}

	// The name.
	name, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"trigger\" keyword: %v", err.Error())
	}

	// First paren.
	if _, err := p.consumeToken(Token_LeftParen); err != nil {
		return nil, p.pushErrorf(0, err.Error())
	}

	// Optionally we could have arguments.
	var args string
	if p.peek().id == Token_TriggerArgs {
		argsStr, err := p.consumeTokenLiteral(Token_TriggerArgs)
		if err != nil {
			return nil, fmt.Errorf("consuming args token: %w", err)
		}
		args = argsStr
	}

	// Last paren.
	if _, err := p.consumeToken(Token_RightParen); err != nil {
		return nil, p.pushErrorf(0, err.Error())
	}

	trigger := &ASTNodeTrigger{
		Name:       name,
		ArgsString: args,
		token:      baseToken,
	}

	p.debuglog("-> parsed trigger: %+v", trigger)

	return trigger, nil
}

// State -------------------------------------------------------------------------------------------
//
// State -> 'state' NAME ( '(' ( Option ( ',' Option )* ')' ) '{' StateBody+'}'

type ASTNodeState struct {
	Name           string
	ParentName     string
	Default        bool
	Options        []*ASTNodeOption
	EnterReactions []*ASTNodeStateReaction
	ExitReactions  []*ASTNodeStateReaction
	Transitions    []*ASTNodeStateTransition
	Captures       []*ASTNodeStateCapture
	Children       []*ASTNodeState

	token *Token
}

var _ ASTNode = (*ASTNodeState)(nil)

func (p *Parser) parseState(parent *ASTNodeState, baseToken *Token) (*ASTNodeState, error) {
	p.debuglog("parsing state node")

	if baseToken.id != Token_KeywordState {
		return nil, p.pushInternalErrorf(0, "expected \"state\" keyword, got: %q", baseToken.id)
	}

	// The name.
	name, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"state\" keyword: %v", err.Error())
	}

	// Debug tracking.
	{
		prevState := p.currentState
		p.currentState = name
		defer func() {
			p.currentState = prevState
		}()
	}

	// The options.
	var options []*ASTNodeOption
	if opts, ok, err := p.parseOptions(); err != nil {
		return nil, fmt.Errorf("state %q: parsing options: %w", name, err)
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
		token:      baseToken,
	}

	// The body.
	{
		// The body opening '{'
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, fmt.Errorf("state %q: parsing state body: %w", name, err)
		}

		bodyStmtsFound := 0

		for !p.atEnd() {
			ok, err := p.parseStateBodyStmt(state)
			if err != nil {
				return nil, fmt.Errorf("state %q: parsing state body statement: %w", state.Name, err)
			}

			if ok {
				bodyStmtsFound++
				continue
			}

			break
		}

		if bodyStmtsFound == 0 {
			return nil, p.pushErrorf(0, "state %q: empty state found, need at least one statement", state.Name)
		}

		// The body closing '}'
		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, p.pushErrorf(0, "state %q: closing state body: %w", name, err)
		}
	}

	return state, nil
}

// StateBody ---------------------------------------------------------------------------------------
//
// StateBodyStmt -> StateEnterReaction
//               -> StateExitReaction
//               -> StateTransition
//               -> State

// When parsing the state body, there are a set of "keywords" that prompt what the next thing will
// be (a transition, another state, etc.). If we don't find one of those, we would rightfully expect
// the state body to be over.
var expectedStateBodyStmtKeywords = []TokenIdentifier{
	Token_KeywordEnter,
	Token_KeywordExit,
	Token_KeywordTransition,
	Token_KeywordCapture,
	Token_KeywordState,
}

func (p *Parser) parseStateBodyStmt(state *ASTNodeState) (bool, error) {
	stmtToken, ok := p.matchAny(expectedStateBodyStmtKeywords...)
	if !ok {
		return false, nil
	}

	switch stmtToken.id {
	case Token_KeywordEnter:
		p.debuglog("parsing state enter reaction node")
		reaction, err := p.parseStateReaction(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing enter reaction: %w", err)
		}
		state.EnterReactions = append(state.EnterReactions, reaction)
		return true, nil
	case Token_KeywordExit:
		p.debuglog("parsing state exit reaction node")
		reaction, err := p.parseStateReaction(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing exit reaction: %w", err)
		}
		state.ExitReactions = append(state.ExitReactions, reaction)
		return true, nil
	case Token_KeywordTransition:
		transition, err := p.parseStateTransition(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing transition: %w", err)
		}
		state.Transitions = append(state.Transitions, transition)
		return true, nil
	case Token_KeywordCapture:
		p.debuglog("parsing state capture reaction")
		capture, err := p.parseStateCapture(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing capture: %w", err)
		}
		state.Captures = append(state.Captures, capture)
		return true, nil
	case Token_KeywordState:
		p.debuglog("parsing child state node")
		child, err := p.parseState(state, stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing child state: %w", err)
		}
		state.Children = append(state.Children, child)
		return true, nil
	}

	return false, p.pushInternalErrorf(0, "expected token to be in %v, got %q", expectedStateBodyStmtKeywords, stmtToken.id)
}

// StateReaction -----------------------------------------------------------------------------------
//
// StateReaction -> 'enter' ( IDENTIFIER )?
//               ->  'exit' ( IDENTIFIER )?

type ASTNodeStateReaction struct {
	Trigger string

	keywordToken *Token
	triggerToken *Token
}

var _ ASTNode = (*ASTNodeStateReaction)(nil)

func (p *Parser) parseStateReaction(baseToken *Token) (*ASTNodeStateReaction, error) {
	p.debuglog("parsing state reaction node")

	if baseToken.id != Token_KeywordEnter && baseToken.id != Token_KeywordExit {
		return nil, p.pushInternalErrorf(0, "unexpected reaction token, expected 'enter', 'exit', got: %q", baseToken.id)
	}

	// We check to see if the user indicated a trigger associated with the reaction.
	var triggerName string
	triggerToken, ok := p.match(Token_Identifier)
	if ok {
		triggerName = triggerToken.literal
	}

	reaction := &ASTNodeStateReaction{
		Trigger:      triggerName,
		keywordToken: baseToken,
		triggerToken: triggerToken,
	}

	p.debuglog("-> parsed reaction: %+v", reaction)
	return reaction, nil
}

// StateTransition ---------------------------------------------------------------------------------
//
// StateTransition -> 'transition' ( 'null' | IDENTIFIER ) > IDENTIFIER

type ASTNodeStateTransition struct {
	Trigger     string
	TargetState string
	Condition   string
	Action      string

	token *Token
}

var _ ASTNode = (*ASTNodeState)(nil)

func (p *Parser) parseStateTransition(baseToken *Token) (*ASTNodeStateTransition, error) {
	p.debuglog("parsing state transition node")

	if baseToken.id != Token_KeywordTransition {
		return nil, p.pushInternalErrorf(0, "unexpected token: expected \"transition\", got: %q", baseToken.id)
	}

	// Check if it's a null transition. If not, we expected an identifier.
	var triggerName string
	if _, ok := p.match(Token_KeywordNull); !ok {
		triggerToken, err := p.consumeToken(Token_Identifier)
		if err != nil {
			return nil, fmt.Errorf("expected null or identifier token: %w", err)
		}
		triggerName = triggerToken.literal
	}

	// The right angle bracket.
	if _, err := p.consumeToken(Token_RightAngleBracket); err != nil {
		return nil, fmt.Errorf("after transition trigger identifier: %w", err)
	}

	// We now expected the trigger identifier.
	stateToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, fmt.Errorf("expected identifier after >: %w", err)
	}

	// Check if there is a condition.
	var condition string
	if _, ok := p.match(Token_KeywordIf); ok {
		// We should get a right brace.
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, p.pushErrorf(0, "after \"if\": %w", err)
		}
		// After we should get a code token.
		if token, err := p.consumeToken(Token_Code); err != nil {
			return nil, p.pushErrorf(0, "expected a code token: %w", err)
		} else {
			condition = token.literal
		}
		// We need a right brace.
		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, p.pushErrorf(0, "expected closing brace: %w", err)
		}
	}

	// Check if there is an action.
	var action string
	if _, ok := p.match(Token_KeywordDo); ok {
		// We should get a right brace.
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, p.pushErrorf(0, "after \"if\": %w", err)
		}
		// After we should get a code token.
		if token, err := p.consumeToken(Token_Code); err != nil {
			return nil, p.pushErrorf(0, "expected a code token: %w", err)
		} else {
			action = token.literal
		}
		// We need a right brace.
		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, p.pushErrorf(0, "expected closing brace: %w", err)
		}
	}

	transition := &ASTNodeStateTransition{
		Trigger:     triggerName,
		TargetState: stateToken.literal,
		Condition:   condition,
		Action:      action,
		token:       baseToken,
	}

	p.debuglog("-> parsed transition: %+v", transition)
	return transition, nil
}

// StateCapture ------------------------------------------------------------------------------------
//
// StateCapture -> 'capture' IDENTIFIER ( 'if' '{' CODE '}' )?

type ASTNodeStateCapture struct {
	Trigger   string
	Condition string

	token *Token
}

var _ ASTNode = (*ASTNodeState)(nil)

func (p *Parser) parseStateCapture(baseToken *Token) (*ASTNodeStateCapture, error) {
	p.debuglog("parsing state capture node")

	if baseToken.id != Token_KeywordCapture {
		return nil, p.pushInternalErrorf(0, "unexpected token: expected \"capture\", got: %q", baseToken.id)
	}

	// The trigger name.
	triggerName, err := p.consumeTokenLiteral(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "expected identifier after (: %w", err)
	}

	// Check if there is a condition.
	var condition string
	if _, ok := p.match(Token_KeywordIf); ok {
		// We should get a right brace.
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, p.pushErrorf(0, "after \"if\": %w", err)
		}
		// After we should get a code token.
		if token, err := p.consumeToken(Token_Code); err != nil {
			return nil, p.pushErrorf(0, "expected a code token: %w", err)
		} else {
			condition = token.literal
		}
		// We need a right brace.
		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, p.pushErrorf(0, "expected closing brace: %w", err)
		}
	}

	capture := &ASTNodeStateCapture{
		Trigger:   triggerName,
		Condition: condition,
		token:     baseToken,
	}

	p.debuglog("-> parsed capture: %+v", capture)

	return capture, nil
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

// matchAny see if the next token matches any of the identifiers given.
// If not, it doesn't advance the stream. If it does, it advances the stream and returns the token.
func (p *Parser) matchAny(ids ...TokenIdentifier) (*Token, bool) {
	if p.atEnd() {
		return nil, false
	}

	peek := p.peek()
	if slices.Contains(ids, peek.id) {
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

func (p *Parser) currentToken() *Token {
	return p.tokenAt(p.current)
}

func (p *Parser) tokenAt(index int) *Token {
	index = max(0, min(index, len(p.tokens)-1))
	return p.tokens[index]
}

// Error -------------------------------------------------------------------------------------------

func (p *Parser) pushErrorf(tokenOffset int, format string, args ...interface{}) error {
	originalErr := fmt.Errorf(format, args...)

	token := p.tokenAt(p.current + tokenOffset)
	fe := errors.Errorf(tokenToPosition(p.lf, token), "Parser", "%s", originalErr)
	p.errors = append(p.errors, fe)

	return originalErr
}

func (p *Parser) pushInternalErrorf(tokenOffset int, format string, args ...interface{}) error {
	originalText := fmt.Sprintf(format, args...)
	finalText := fmt.Sprintf("%s\n*** This is most likely an internal error. Please file a bug, hopefully with a repro. ***", originalText)
	return p.pushErrorf(tokenOffset, finalText)
}

// Debug -------------------------------------------------------------------------------------------

func (p *Parser) debuglog(format string, args ...interface{}) {
	if !p.options.Verbose {
		return
	}

	fmt.Printf("GOCHART PARSER: STATECHART %q, STATE: %q: %s\n", p.currentStatechart, p.currentState, fmt.Sprintf(format, args...))
}
