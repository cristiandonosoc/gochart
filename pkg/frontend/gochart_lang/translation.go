package gochart_lang

import (
	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

type frontendTranslator struct {
	lf *files.LoadedFile
}

func newFrontendTranslator(lf *files.LoadedFile) *frontendTranslator {
	return &frontendTranslator{
		lf: lf,
	}
}

func (ft *frontendTranslator) translateStatechart(nsc *ASTNodeStatechart) *output.StatechartData {
	// Collect triggers.
	triggers := make([]*output.TriggerData, 0, len(nsc.Triggers))
	for _, nt := range nsc.Triggers {
		triggers = append(triggers, ft.translateTrigger(nt))
	}

	// Collect states.
	var states []*output.StateData
	var transitions []*output.TransitionData
	for _, ns := range nsc.States {
		states, transitions = ft.translateState(ns, states, transitions)
	}

	return &output.StatechartData{
		Name:             nsc.Name,
		Triggers:         triggers,
		States:           states,
		Transitions:      transitions,
		FrontendPosition: tokenToPosition(ft.lf, nsc.token),
	}
}

func (ft *frontendTranslator) translateTrigger(nt *ASTNodeTrigger) *output.TriggerData {
	return &output.TriggerData{
		Name:             nt.Name,
		ArgsString:       nt.ArgsString,
		FrontendPosition: tokenToPosition(ft.lf, nt.token),
	}
}

func (ft *frontendTranslator) translateState(ns *ASTNodeState, states []*output.StateData, transitions []*output.TransitionData) ([]*output.StateData, []*output.TransitionData) {
	state := &output.StateData{
		Name:             ns.Name,
		Parent:           ns.ParentName,
		Options:          ft.translateOptions(ns.Options),
		FrontendPosition: tokenToPosition(ft.lf, ns.token),
	}

	for _, er := range ns.EnterReactions {
		state.EnterReactions = append(state.EnterReactions, &output.ReactionData{
			Trigger:          er.Trigger,
			FrontendPosition: tokenToPosition(ft.lf, er.token),
		})
	}

	for _, er := range ns.ExitReactions {
		state.ExitReactions = append(state.ExitReactions, &output.ReactionData{
			Trigger:          er.Trigger,
			FrontendPosition: tokenToPosition(ft.lf, er.token),
		})
	}
	states = append(states, state)

	for _, c := range ns.Captures {
		state.Captures = append(state.Captures, &output.CaptureData{
			Trigger:          c.Trigger,
			Condition:        c.Condition,
			FrontendPosition: tokenToPosition(ft.lf, c.token),
		})
	}

	// Collect the transitions.
	for _, t := range ns.Transitions {
		transition := &output.TransitionData{
			From:             ns.Name,
			To:               t.TargetState,
			Trigger:          t.Trigger,
			Condition:        t.Condition,
			Action:           t.Action,
			FrontendPosition: tokenToPosition(ft.lf, t.token),
		}
		transitions = append(transitions, transition)
	}

	// Collect the children.
	for _, child := range ns.Children {
		states, transitions = ft.translateState(child, states, transitions)
	}

	return states, transitions
}

func (ft *frontendTranslator) translateOptions(nos []*ASTNodeOption) []*output.OptionData {
	options := make([]*output.OptionData, 0, len(nos))
	for _, no := range nos {
		options = append(options, &output.OptionData{
			Name:             no.Name,
			Kind:             no.ValueKind,
			Value:            no.Value,
			FrontendPosition: tokenToPosition(ft.lf, no.token),
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
