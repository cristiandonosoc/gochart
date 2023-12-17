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

func (ft *frontendTranslator) translateStatechart(nsc *ASTNodeStatechart) *common.StatechartData {
	// Collect triggers.
	triggers := make([]*common.TriggerData, 0, len(nsc.Triggers))
	for _, nt := range nsc.Triggers {
		triggers = append(triggers, ft.translateTrigger(nt))
	}

	// Collect states.
	var states []*common.StateData
	var transitions []*common.TransitionData
	for _, ns := range nsc.States {
		states, transitions = ft.translateState(ns, states, transitions)
	}

	return &common.StatechartData{
		Name:                nsc.Name,
		Triggers:            triggers,
		States:              states,
		Transitions:         transitions,
		Frontend_KeywordPos: tokenToPosition(ft.lf, nsc.keywordToken),
		Frontend_NamePos:    tokenToPosition(ft.lf, nsc.nameToken),
	}
}

func (ft *frontendTranslator) translateTrigger(nt *ASTNodeTrigger) *common.TriggerData {
	return &common.TriggerData{
		Name:                nt.Name,
		ArgsString:          nt.ArgsString,
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
		Frontend_KeywordPos: tokenToPosition(ft.lf, ns.keywordToken),
		Frontend_NamePos:    tokenToPosition(ft.lf, ns.nameToken),
	}

	for _, er := range ns.EnterReactions {
		state.EnterReactions = append(state.EnterReactions, &common.ReactionData{
			Trigger:             er.Trigger,
			Frontend_KeywordPos: tokenToPosition(ft.lf, er.keywordToken),
			Frontend_TriggerPos: tokenToPosition(ft.lf, er.triggerToken),
		})
	}

	for _, er := range ns.ExitReactions {
		state.ExitReactions = append(state.ExitReactions, &common.ReactionData{
			Trigger:             er.Trigger,
			Frontend_KeywordPos: tokenToPosition(ft.lf, er.keywordToken),
			Frontend_TriggerPos: tokenToPosition(ft.lf, er.triggerToken),
		})
	}
	states = append(states, state)

	for _, c := range ns.Captures {
		state.Captures = append(state.Captures, &common.CaptureData{
			Trigger:             c.Trigger,
			Condition:           c.Condition,
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

func tokenToPosition(lf *files.LoadedFile, token *Token) *files.LoadedFilePosition {
	if token == nil {
		return nil
	}

	return newLoadedFilePosition(lf, token.line, token.char)
}

func newLoadedFilePosition(lf *files.LoadedFile, line, char int) *files.LoadedFilePosition {
	return &files.LoadedFilePosition{
		File: lf,
		Line: line,
		Char: char,
	}
}
