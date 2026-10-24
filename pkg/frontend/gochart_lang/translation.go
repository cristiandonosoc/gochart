package gochart_lang

import (
	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

type frontendTranslator struct {
	lf *files.LoadedFile
}

func newFrontendTranslator(lf *files.LoadedFile) *frontendTranslator {
	return &frontendTranslator{
		lf: lf,
	}
}

func (ft *frontendTranslator) translateGochart(ngc *ASTNodeGochart) *common.GochartData {
	// Collect triggers.
	triggers := make([]*common.TriggerData, 0, len(ngc.Triggers))
	for _, nt := range ngc.Triggers {
		triggers = append(triggers, ft.translateTrigger(nt))
	}

	// Collect states.
	var states []*common.StateData
	var transitions []*common.TransitionData

	states, transitions = ft.translateState(ngc.Root, states, transitions)
	if states[0].Name != common.RootStateName {
		panic("root state is not the first one!")
	}

	return &common.GochartData{
		Name:                ngc.Name,
		Options:             ft.translateOptions(ngc.Options),
		Triggers:            triggers,
		Root:                states[0],
		States:              states,
		Transitions:         transitions,
		Comments:            ft.translateComments(ngc.Comments),
		Frontend_KeywordPos: tokenToPosition(ft.lf, ngc.keywordToken),
		Frontend_NamePos:    tokenToPosition(ft.lf, ngc.nameToken),
	}
}

func (ft *frontendTranslator) translateTrigger(nt *ASTNodeTrigger) *common.TriggerData {
	return &common.TriggerData{
		Name:                nt.Name,
		ArgsString:          nt.ArgsString,
		Frontend_Index:      nt.entryIndex,
		Frontend_KeywordPos: tokenToPosition(ft.lf, nt.keywordToken),
		Frontend_NamePos:    tokenToPosition(ft.lf, nt.nameToken),
		Frontend_ArgsPos:    tokenToPosition(ft.lf, nt.argsToken),
	}
}

func (ft *frontendTranslator) translateState(ns *ASTNodeState, states []*common.StateData, transitions []*common.TransitionData) ([]*common.StateData, []*common.TransitionData) {
	state := &common.StateData{
		Name:                ns.Name,
		Parent:              ns.ParentName,
		Options:             ft.translateOptions(ns.Options),
		Frontend_Index:      ns.entryIndex,
		Frontend_KeywordPos: tokenToPosition(ft.lf, ns.keywordToken),
		Frontend_NamePos:    tokenToPosition(ft.lf, ns.nameToken),
	}

	for _, er := range ns.EnterReactions {
		state.EnterReactions = append(state.EnterReactions, &common.ReactionData{
			Trigger:             er.Trigger,
			Frontend_Index:      er.entryIndex,
			Frontend_KeywordPos: tokenToPosition(ft.lf, er.keywordToken),
			Frontend_TriggerPos: tokenToPosition(ft.lf, er.triggerToken),
		})
	}

	for _, er := range ns.ExitReactions {
		state.ExitReactions = append(state.ExitReactions, &common.ReactionData{
			Trigger:             er.Trigger,
			Frontend_Index:      er.entryIndex,
			Frontend_KeywordPos: tokenToPosition(ft.lf, er.keywordToken),
			Frontend_TriggerPos: tokenToPosition(ft.lf, er.triggerToken),
		})
	}
	states = append(states, state)

	for _, c := range ns.Captures {
		state.Captures = append(state.Captures, &common.CaptureData{
			Trigger:             c.Trigger,
			Condition:           c.Condition,
			Frontend_Index:      c.entryIndex,
			Frontend_KeywordPos: tokenToPosition(ft.lf, c.keywordToken),
			Frontend_TriggerPos: tokenToPosition(ft.lf, c.triggerToken),
		})
	}

	// Collect the transitions.
	for _, t := range ns.Transitions {
		transition := &common.TransitionData{
			From:                ns.Name,
			To:                  t.TargetState,
			Trigger:             t.Trigger,
			Condition:           t.Condition,
			Action:              t.Action,
			Frontend_Index:      t.entryIndex,
			Frontend_KeywordPos: tokenToPosition(ft.lf, t.keywordToken),
			Frontend_TriggerPos: tokenToPosition(ft.lf, t.triggerToken),
			Frontend_TargetPos:  tokenToPosition(ft.lf, t.targetToken),
		}
		transitions = append(transitions, transition)
	}

	// Collect the children.
	for _, child := range ns.Children {
		states, transitions = ft.translateState(child, states, transitions)
	}

	state.Comments = ft.translateComments(ns.Comments)

	return states, transitions
}

func (ft *frontendTranslator) translateOptions(nos []*ASTNodeOption) []*common.OptionData {
	options := make([]*common.OptionData, 0, len(nos))
	for _, no := range nos {
		options = append(options, &common.OptionData{
			Name:             no.Name,
			Kind:             no.ValueKind,
			Value:            no.Value,
			Frontend_NamePos: tokenToPosition(ft.lf, no.token),
		})
	}

	return options
}

func (ft *frontendTranslator) translateComments(nodes []*ASTNodeComment) []*common.CommentData {
	comments := make([]*common.CommentData, 0, len(nodes))
	for _, node := range nodes {
		comments = append(comments, &common.CommentData{
			Comment:           node.comment,
			Frontend_Index:    node.entryIndex,
			Frontend_Position: tokenToPosition(ft.lf, node.token),
		})
	}

	return comments
}

func tokenToPosition(lf *files.LoadedFile, token *Token) *files.LoadedFilePosition {
	if token == nil {
		return nil
	}

	return files.NewLoadedFilePosition(lf, token.line, token.char)
}
