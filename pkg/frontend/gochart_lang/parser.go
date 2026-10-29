package gochart_lang

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/cristiandonosoc/gochart/internal/errors"
	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

// When parsing the state body, there are a set of "keywords" that prompt what the next thing will
// be (a transition, another state, etc.). If we don't find one of those, we would rightfully expect
// the state body to be over.
var (
	expectedStateBodyStmtKeywords = []TokenIdentifier{
		Token_KeywordEnter,
		Token_KeywordExit,
		Token_KeywordTransition,
		Token_KeywordCapture,
		Token_KeywordState,
		Token_Comment,
	}

	expectedGochartBodyStmtKeywords = []TokenIdentifier{
		Token_KeywordTrigger,
		// Plus all the state ones.
		Token_KeywordEnter,
		Token_KeywordExit,
		Token_KeywordTransition,
		Token_KeywordCapture,
		Token_KeywordState,
		Token_Comment,
	}
)

// Parser is an object capable of taking tokens of the gochart_lang language and parse it against
// the grammar, returning an Abstract Syntax Tree (AST).
type Parser struct {
	tokens  []*Token
	current int

	currentGochart string
	currentState   string

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
	// AppendNode is a convenient function to be able to append nodes into a node and have correct
	// tracking of the internal indices.
	AppendNode(node ASTNode)
}

// RULES *******************************************************************************************

// Root --------------------------------------------------------------------------------------------
//
// Root -> Gochart+

type ASTNodeRoot struct {
	gocharts []*ASTNodeGochart
}

var _ ASTNode = (*ASTNodeRoot)(nil)

func (n *ASTNodeRoot) AppendNode(node ASTNode) {
	panic("this node cannot append!")
}

// root -> gochart+
func (p *Parser) parseRoot() (*ASTNodeRoot, []error) {
	p.debuglog("parsing root node")

	root := &ASTNodeRoot{}

	var gocharts []*ASTNodeGochart

	gc, err := p.parseGochart()
	if err != nil {
		if len(p.errors) == 0 {
			p.pushInternalErrorf(0, "unaccounted parser error on err: %v", err.Error())
		}
		return nil, p.errors
	}
	gocharts = append(gocharts, gc)

	// See if there are more gochart nodes.
	// TODO(cdc): See how we can make the parser recover from an error.
	for {
		if p.peek().id == Token_KeywordGochart {
			gc, err := p.parseGochart()
			if err != nil {
				if len(p.errors) == 0 {
					p.pushInternalErrorf(0, fmt.Sprintf("unaccounted parser error on err: %v", err.Error()))
				}
				return nil, p.errors
			}
			gocharts = append(gocharts, gc)
		}

		break
	}

	if !p.atEnd() {
		p.pushErrorf(0, "expected \"gochart\" keyword, got %q", p.peek())
	}

	root.gocharts = gocharts
	return root, p.errors
}

// Gochart --------------------------------------------------------------------------------------
//
// Gochart -> 'gochart' NAME ( '(' ')' )? '{' GochartBodyStmt+ '}'

type ASTNodeGochart struct {
	Name string

	Options []*ASTNodeOption

	Triggers []*ASTNodeTrigger
	Root     *ASTNodeState
	States   []*ASTNodeState

	// These comments will be outside the gochart, as inside comments will be inside of the root
	// state comment tracking.
	Comments []*ASTNodeComment

	keywordToken *Token
	nameToken    *Token

	entryCount int
}

var _ ASTNode = (*ASTNodeGochart)(nil)

func (n *ASTNodeGochart) AppendNode(node ASTNode) {
	switch v := node.(type) {
	case *ASTNodeTrigger:
		// We use the indices of the root state.
		v.entryIndex = n.Root.entryCount
		n.Root.entryCount++
		n.Triggers = append(n.Triggers, v)
	case *ASTNodeComment:
		v.entryIndex = n.entryCount
		n.entryCount++
		n.Comments = append(n.Comments, v)
	default:
		panic(fmt.Sprintf("unsupported type being appended %q", reflect.TypeOf(v).Name()))
	}
}

func (n *ASTNodeGochart) NonCommentEntryCount() int {
	return len(n.Triggers) + n.Root.NonCommentEntryCount()
}

func (p *Parser) parseGochart() (*ASTNodeGochart, error) {
	p.debuglog("parsing gochart node")

	// Create the tentative gochart to be able to track comments.
	// We create the root state immediately as we use that for the index source.
	gc := &ASTNodeGochart{
		Root: &ASTNodeState{},
	}

	// See if we get any out of initial comments.
	for {
		commentToken, ok := p.match(Token_Comment)
		if !ok {
			break
		}

		gc.AppendNode(&ASTNodeComment{
			comment: commentToken.literal,
			token:   commentToken,
		})
	}

	// gochart keyword.
	baseToken, err := p.consumeToken(Token_KeywordGochart)
	if err != nil {
		return nil, p.pushErrorf(0, err.Error())
	}

	// The name.
	nameToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"gochart\" keyword: %v", err.Error())
	}
	name := nameToken.literal

	// Debug tracking.
	{
		prevGochart := p.currentGochart
		p.currentGochart = name
		defer func() {
			p.currentGochart = prevGochart
		}()
	}

	// The options.
	var options []*ASTNodeOption
	if opts, ok, err := p.parseOptions(); err != nil {
		return nil, fmt.Errorf("gochart %q: parsing options: %w", name, err)
	} else if ok {
		options = opts
	}

	// Fill in the gochart now that we have the info.
	gc.Name = name
	gc.Options = options
	gc.keywordToken = baseToken
	gc.nameToken = nameToken
	// The root state has the position of the gochart.
	gc.Root.Name = common.RootStateName
	gc.Root.keywordToken = baseToken
	gc.Root.nameToken = nameToken
	gc.Root.entryIndex = gc.entryCount
	gc.entryCount++ // To account for the root state.

	{
		// {
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, p.pushErrorf(0, err.Error())
		}

		for !p.atEnd() {
			ok, err := p.parseGochartBodyStmt(gc)
			if err != nil {
				return nil, fmt.Errorf("parsing gochart body statement: %w", err)
			}

			if ok {
				continue
			}

			break
		}

		if gc.NonCommentEntryCount() == 0 {
			return nil, p.pushErrorf(0, "empty gochart found, need at least one statement")
		}

		if _, err := p.consumeToken(Token_RightBrace); err != nil {
			return nil, p.pushErrorf(0, err.Error())
		}
	}

	// See if we get any comments after the gochart.
	for {
		commentToken, ok := p.match(Token_Comment)
		if !ok {
			break
		}

		gc.AppendNode(&ASTNodeComment{
			comment: commentToken.literal,
			token:   commentToken,
		})
	}

	return gc, nil
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
	ValueKind common.OptionValueKind
	Value     string

	token *Token
}

var _ ASTNode = (*ASTNodeOption)(nil)

func (n *ASTNodeOption) AppendNode(node ASTNode) {
	panic("this node cannot append!")
}

func (p *Parser) parseOption() (*ASTNodeOption, error) {
	p.debuglog("-> parsing state option")

	// The name.
	nameToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"trigger\" keyword")
	}

	// See if there is an optional value.
	// If not, by default this is a "true" (empty) boolean option.
	valueKind := common.OptionValueKind_Boolean
	value := ""
	if _, ok := p.match(Token_Equal); ok {
		if token, ok := p.match(Token_Identifier); ok {
			valueKind = common.OptionValueKind_String
			value = token.literal
		} else if token, ok := p.match(Token_String); ok {
			valueKind = common.OptionValueKind_String
			value = token.literal
		} else if token, ok := p.match(Token_Integer); ok {
			_, err := strconv.Atoi(token.literal)
			if err != nil {
				return nil, fmt.Errorf("could not parse option value %q to int: %w", token.literal, err)
			}
			valueKind = common.OptionValueKind_Integer
			value = token.literal
		} else if token, ok := p.match(Token_Boolean); ok {
			_, err := strconv.ParseBool(token.literal)
			if err != nil {
				return nil, fmt.Errorf("could not parse option value %q to boolean: %w", token.literal, err)
			}
			valueKind = common.OptionValueKind_Boolean
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

// GochartBodyStmt ------------------------------------------------------------------------------
//
// GochartBodyStmt -> Trigger
//                 -> State

func (p *Parser) parseGochartBodyStmt(gc *ASTNodeGochart) (bool, error) {
	// If we're closing the brace, we simply exit the statement.
	if peek := p.peek(); peek.id == Token_RightBrace {
		return false, nil
	}

	peek := p.peek()

	if !slices.Contains(expectedGochartBodyStmtKeywords, peek.id) {
		return false, p.pushErrorf(0, "expected token to be in %v, got %q", expectedGochartBodyStmtKeywords, peek.id)
	}

	// If we found a trigger, we hook it as such.
	if peek.id == Token_KeywordTrigger {
		trigger, err := p.parseTrigger(p.advance())
		if err != nil {
			return false, fmt.Errorf("parsing trigger: %w", err)
		}
		gc.AppendNode(trigger)
		return true, nil
	}

	// Otherwise all statements go into our root state.
	if ok, err := p.parseStateBodyStmt(gc.Root); err != nil {
		return false, fmt.Errorf("parsing root state body statement: %w", err)
	} else if ok {
		return true, nil
	}

	return false, p.pushInternalErrorf(0, "expected token to be in %v, got %q", expectedGochartBodyStmtKeywords, peek.id)
}

// Trigger -----------------------------------------------------------------------------------------
//
// Trigger -> 'trigger' NAME '(' ARG_TYPE+ ARG_NAME (',' ARG_TYPE+ ARG_NAME)* ')'

type ASTNodeTrigger struct {
	Name       string
	ArgsString string

	keywordToken *Token
	nameToken    *Token
	argsToken    *Token
	entryIndex   int
}

var _ ASTNode = (*ASTNodeTrigger)(nil)

func (n *ASTNodeTrigger) AppendNode(node ASTNode) {
	panic("this node cannot append!")
}

func (p *Parser) parseTrigger(baseToken *Token) (*ASTNodeTrigger, error) {
	p.debuglog("parsing trigger node")

	if baseToken.id != Token_KeywordTrigger {
		return nil, p.pushInternalErrorf(0, "expected 'trigger' keyword, got: %q", baseToken.id)
	}

	// The name.
	nameToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"trigger\" keyword: %v", err.Error())
	}
	name := nameToken.literal

	// First paren.
	if _, err := p.consumeToken(Token_LeftParen); err != nil {
		return nil, p.pushErrorf(0, err.Error())
	}

	// Optionally we could have arguments.
	var args string
	var argsToken *Token
	if p.peek().id == Token_TriggerArgs {
		at, err := p.consumeToken(Token_TriggerArgs)
		if err != nil {
			return nil, fmt.Errorf("consuming args token: %w", err)
		}
		argsToken = at
		args = at.literal
	}

	// Last paren.
	if _, err := p.consumeToken(Token_RightParen); err != nil {
		return nil, p.pushErrorf(0, err.Error())
	}

	trigger := &ASTNodeTrigger{
		Name:         name,
		ArgsString:   args,
		keywordToken: baseToken,
		nameToken:    nameToken,
		argsToken:    argsToken,
	}

	p.debuglog("-> parsed trigger: %+v", trigger)

	return trigger, nil
}

// State -------------------------------------------------------------------------------------------
//
// State -> 'state' NAME ( '(' ( Option ( ',' Option )* ')' ) '{' StateBody+'}'

type ASTNodeState struct {
	Name       string
	ParentName string
	Default    bool
	Options    []*ASTNodeOption

	EnterReactions []*ASTNodeStateReaction
	ExitReactions  []*ASTNodeStateReaction
	Transitions    []*ASTNodeStateTransition
	Captures       []*ASTNodeStateCapture
	Children       []*ASTNodeState

	Comments []*ASTNodeComment

	keywordToken *Token
	nameToken    *Token

	entryIndex int
	entryCount int
}

var _ ASTNode = (*ASTNodeState)(nil)

func (n *ASTNodeState) AppendNode(node ASTNode) {
	switch v := node.(type) {
	case *ASTNodeStateReaction:
		v.entryIndex = n.entryCount
		if v.IsEnterReaction() {
			n.EnterReactions = append(n.EnterReactions, v)
		}
		if v.IsExitReaction() {
			n.ExitReactions = append(n.ExitReactions, v)
		}
	case *ASTNodeStateTransition:
		v.entryIndex = n.entryCount
		n.Transitions = append(n.Transitions, v)
	case *ASTNodeStateCapture:
		v.entryIndex = n.entryCount
		n.Captures = append(n.Captures, v)
	case *ASTNodeState:
		v.entryIndex = n.entryCount
		n.Children = append(n.Children, v)
	case *ASTNodeComment:
		v.entryIndex = n.entryCount
		n.Comments = append(n.Comments, v)
	default:
		panic(fmt.Sprintf("unsupported type being appended %q", reflect.TypeOf(v).Name()))
	}

	n.entryCount++
}

func (n *ASTNodeState) NonCommentEntryCount() int {
	count := len(n.EnterReactions)
	count += len(n.ExitReactions)
	count += len(n.Transitions)
	count += len(n.Captures)
	count += len(n.Children)

	return count
}

func (p *Parser) parseState(parent *ASTNodeState, baseToken *Token) (*ASTNodeState, error) {
	p.debuglog("parsing state node")

	if baseToken.id != Token_KeywordState {
		return nil, p.pushInternalErrorf(0, "expected \"state\" keyword, got: %q", baseToken.id)
	}

	// The name.
	nameToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "after \"state\" keyword: %v", err.Error())
	}
	name := nameToken.literal

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

		keywordToken: baseToken,
		nameToken:    nameToken,
	}

	// The body.
	{
		// The body opening '{'
		if _, err := p.consumeToken(Token_LeftBrace); err != nil {
			return nil, fmt.Errorf("state %q: parsing state body: %w", name, err)
		}

		for !p.atEnd() {
			ok, err := p.parseStateBodyStmt(state)
			if err != nil {
				return nil, fmt.Errorf("state %q: parsing state body statement: %w", state.Name, err)
			}

			if ok {
				continue
			}

			break
		}

		if state.NonCommentEntryCount() == 0 {
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
//               -> Comment

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
		state.AppendNode(reaction)
		return true, nil
	case Token_KeywordExit:
		p.debuglog("parsing state exit reaction node")
		reaction, err := p.parseStateReaction(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing exit reaction: %w", err)
		}
		state.AppendNode(reaction)
		return true, nil
	case Token_KeywordTransition:
		transition, err := p.parseStateTransition(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing transition: %w", err)
		}
		state.AppendNode(transition)
		return true, nil
	case Token_KeywordCapture:
		p.debuglog("parsing state capture reaction")
		capture, err := p.parseStateCapture(stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing capture: %w", err)
		}
		state.AppendNode(capture)
		return true, nil
	case Token_KeywordState:
		p.debuglog("parsing child state node")
		child, err := p.parseState(state, stmtToken)
		if err != nil {
			return false, fmt.Errorf("parsing child state: %w", err)
		}
		state.AppendNode(child)
		return true, nil
	case Token_Comment:
		p.debuglog("parsing comment node")
		state.AppendNode(&ASTNodeComment{
			comment: stmtToken.literal,
			token:   stmtToken,
		})
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
	entryIndex   int
}

var _ ASTNode = (*ASTNodeStateReaction)(nil)

func (n *ASTNodeStateReaction) AppendNode(node ASTNode) {
	panic("this node cannot append!")
}

func (n *ASTNodeStateReaction) IsEnterReaction() bool {
	return n.keywordToken.id == Token_KeywordEnter
}

func (n *ASTNodeStateReaction) IsExitReaction() bool {
	return n.keywordToken.id == Token_KeywordExit
}

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

	// Named callbacks ("if some_condition", "do some_action"), as opposed to inlined code.
	ConditionCallback string
	ActionCallback    string

	keywordToken *Token
	triggerToken *Token
	targetToken  *Token
	entryIndex   int
}

var _ ASTNode = (*ASTNodeStateTransition)(nil)

func (n *ASTNodeStateTransition) AppendNode(node ASTNode) {
	panic("this node cannot append!")
}

func (p *Parser) parseStateTransition(baseToken *Token) (*ASTNodeStateTransition, error) {
	p.debuglog("parsing state transition node")

	if baseToken.id != Token_KeywordTransition {
		return nil, p.pushInternalErrorf(0, "unexpected token: expected \"transition\", got: %q", baseToken.id)
	}

	// Check if it's a null transition. If not, we expected an identifier.
	var triggerName string
	var triggerToken *Token
	if nullToken, ok := p.match(Token_KeywordNull); ok {
		triggerToken = nullToken
	} else {
		tt, err := p.consumeToken(Token_Identifier)
		if err != nil {
			return nil, p.pushErrorf(0, "expected null or identifier token: %w", err)
		}
		triggerToken = tt
		triggerName = tt.literal
	}

	// The right angle bracket.
	if _, err := p.consumeToken(Token_RightAngleBracket); err != nil {
		return nil, p.pushErrorf(0, "after transition trigger identifier: %w", err)
	}

	// We now expected the trigger identifier.
	stateToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		// return nil, p.pushErrorf(0, "expected identifier after >: %w", err)
		return nil, p.pushErrorf(0, "expected identifier after >: %w", err)
	}

	// Check if there is a condition.
	var condition, conditionCallback string
	if _, ok := p.match(Token_KeywordIf); ok {
		code, callback, err := p.parseCodeOrCallback("if")
		if err != nil {
			return nil, err
		}
		condition, conditionCallback = code, callback
	}

	// Check if there is an action.
	var action, actionCallback string
	if _, ok := p.match(Token_KeywordDo); ok {
		code, callback, err := p.parseCodeOrCallback("do")
		if err != nil {
			return nil, err
		}
		action, actionCallback = code, callback
	}

	transition := &ASTNodeStateTransition{
		Trigger:           triggerName,
		TargetState:       stateToken.literal,
		Condition:         condition,
		Action:            action,
		ConditionCallback: conditionCallback,
		ActionCallback:    actionCallback,
		keywordToken:      baseToken,
		triggerToken:      triggerToken,
		targetToken:       stateToken,
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

	// Named condition callback ("if some_condition"), as opposed to inlined code.
	ConditionCallback string

	keywordToken *Token
	triggerToken *Token
	entryIndex   int
}

var _ ASTNode = (*ASTNodeStateCapture)(nil)

func (n *ASTNodeStateCapture) AppendNode(node ASTNode) {
	panic("this node cannot append!")
}

func (p *Parser) parseStateCapture(baseToken *Token) (*ASTNodeStateCapture, error) {
	p.debuglog("parsing state capture node")

	if baseToken.id != Token_KeywordCapture {
		return nil, p.pushInternalErrorf(0, "unexpected token: expected \"capture\", got: %q", baseToken.id)
	}

	// The trigger name.
	triggerToken, err := p.consumeToken(Token_Identifier)
	if err != nil {
		return nil, p.pushErrorf(0, "expected identifier after (: %w", err)
	}
	triggerName := triggerToken.literal

	// Check if there is a condition.
	var condition, conditionCallback string
	if _, ok := p.match(Token_KeywordIf); ok {
		code, callback, err := p.parseCodeOrCallback("if")
		if err != nil {
			return nil, err
		}
		condition, conditionCallback = code, callback
	}

	capture := &ASTNodeStateCapture{
		Trigger:           triggerName,
		Condition:         condition,
		ConditionCallback: conditionCallback,
		keywordToken:      baseToken,
		triggerToken:      triggerToken,
	}

	p.debuglog("-> parsed capture: %+v", capture)

	return capture, nil
}

// parseCodeOrCallback parses what comes after an "if" or "do" keyword: either an inlined code block
// ('{' CODE '}') or the name of a callback (IDENTIFIER). Returns one of the two.
func (p *Parser) parseCodeOrCallback(keyword string) (string, string, error) {
	if token, ok := p.match(Token_Identifier); ok {
		return "", token.literal, nil
	}

	// We should get a left brace.
	if _, err := p.consumeToken(Token_LeftBrace); err != nil {
		return "", "", p.pushErrorf(0, "after %q: %w", keyword, err)
	}
	// After we should get a code token.
	token, err := p.consumeToken(Token_Code)
	if err != nil {
		return "", "", p.pushErrorf(0, "expected a code token: %w", err)
	}
	// We need a right brace.
	if _, err := p.consumeToken(Token_RightBrace); err != nil {
		return "", "", p.pushErrorf(0, "expected closing brace: %w", err)
	}

	return token.literal, "", nil
}

// Comment -----------------------------------------------------------------------------------------
//
// // <COMMENT>\n

type ASTNodeComment struct {
	comment string
	token   *Token

	entryIndex int
}

var _ ASTNode = (*ASTNodeComment)(nil)

func (n *ASTNodeComment) AppendNode(node ASTNode) {
	panic("this node cannot append!")
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

// consumeTokenLiteral is an even more revised convenience variation that tries to consume a token
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
	block := `
**********************************************************************************
***** THIS IS MOST LIKELY AN INTERNAL ERROR. PLEASE FILE A BUG WITH A REPRO. *****
**********************************************************************************`
	finalText := fmt.Sprintf("%s\n%s\n", originalText, block)
	return p.pushErrorf(tokenOffset, finalText)
}

// Debug -------------------------------------------------------------------------------------------

func (p *Parser) debuglog(format string, args ...interface{}) {
	if !p.options.Verbose {
		return
	}

	fmt.Printf("GOCHART PARSER: STATECHART %q, STATE: %q: %s\n", p.currentGochart, p.currentState, fmt.Sprintf(format, args...))
}
