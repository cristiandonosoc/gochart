package gochart_lang

import (
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

type frontendTranslator struct {
	triggerIndex    int
	stateIndex      int
	transitionIndex int
}

func newFrontendTranslator() *frontendTranslator {
	return &frontendTranslator{}
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
		Name:        nsc.Name,
		Triggers:    triggers,
		States:      states,
		Transitions: transitions,
	}
}

func (ft *frontendTranslator) translateTrigger(nt *ASTNodeTrigger) *output.TriggerData {
	index := ft.triggerIndex
	ft.triggerIndex += 1
	return &output.TriggerData{
		Name:       nt.Name,
		ArgsString: nt.ArgsString,
		Index:      index,
	}
}

func (ft *frontendTranslator) translateState(ns *ASTNodeState, states []*output.StateData, transitions []*output.TransitionData) ([]*output.StateData, []*output.TransitionData) {
	index := ft.stateIndex
	ft.stateIndex++
	state := &output.StateData{
		Name:   ns.Name,
		Parent: ns.ParentName,
		Index:  index,
	}

	for _, er := range ns.EnterReactions {
		state.EnterReactions = append(state.EnterReactions, er.Trigger)
	}

	for _, er := range ns.ExitReactions {
		state.ExitReactions = append(state.ExitReactions, er.Trigger)
	}
	states = append(states, state)

	// Collect the transitions.
	for _, t := range ns.Transitions {
		index := ft.transitionIndex
		ft.transitionIndex++
		transition := &output.TransitionData{
			From:    ns.Name,
			To:      t.TargetState,
			Trigger: t.Trigger,
			Index:   index,
		}
		transitions = append(transitions, transition)
	}

	// Collect the children.
	for _, child := range ns.Children {
		states, transitions = ft.translateState(child, states, transitions)
	}

	return states, transitions
}
