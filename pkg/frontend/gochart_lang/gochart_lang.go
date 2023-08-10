// Package gochart_lang is the frontend for what we can "gochart_lang", which is a custom spec
// language to describe statecharts.
package gochart_lang

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

type gochartLangFrontend struct {
}

var _ frontend.GochartFrontend = (*gochartLangFrontend)(nil)

// NewFrontend returns a frontend capable of parting the gochart language.
func NewGochartLangFrontend() *gochartLangFrontend {
	return &gochartLangFrontend{}
}

func (glf *gochartLangFrontend) Process(r io.Reader) (*frontend.StatechartData, error) {
	// data, err := io.ReadAll(r)
	// if err != nil {
	// 	return nil, fmt.Errorf("reading input reader: %w", err)
	// }

	// var scdata frontend.StatechartData
	return nil, fmt.Errorf("IMPLEMENT ME")
}

// AST TO OUTPUT TRANSLATION -----------------------------------------------------------------------
//
// This section translated from the AST nodes into the format the output expects.

type frontendTranslator struct {
	triggerIndex    int
	stateIndex      int
	transitionIndex int
}

func newFrontendTranslator() *frontendTranslator {
	return &frontendTranslator{}
}

func (ft *frontendTranslator) translateStatechart(nsc *ASTNodeStatechart) *frontend.StatechartData {
	// Collect triggers.
	triggers := make([]*frontend.TriggerData, 0, len(nsc.Triggers))
	for _, nt := range nsc.Triggers {
		triggers = append(triggers, ft.translateTrigger(nt))
	}

	// Collect states.
	var states []*frontend.StateData
	var transitions []*frontend.TransitionData
	for _, ns := range nsc.States {
		states, transitions = ft.translateState(ns, states, transitions)
	}

	return &frontend.StatechartData{
		Name:        nsc.Name,
		Triggers:    triggers,
		States:      states,
		Transitions: transitions,
	}
}

func (ft *frontendTranslator) translateTrigger(nt *ASTNodeTrigger) *frontend.TriggerData {
	index := ft.triggerIndex
	ft.triggerIndex += 1
	return &frontend.TriggerData{
		Name:       nt.Name,
		ArgsString: nt.ArgsString,
		Index:      index,
	}
}

func (ft *frontendTranslator) translateState(ns *ASTNodeState, states []*frontend.StateData, transitions []*frontend.TransitionData) ([]*frontend.StateData, []*frontend.TransitionData) {
	index := ft.stateIndex
	ft.stateIndex++
	state := &frontend.StateData{
		Name:   ns.Name,
		Parent: ns.ParentName,
		Index:  index,
	}

	for _, er := range ns.EnterReactions {
		state.EnterReactionTriggers = append(state.EnterReactionTriggers, er.Trigger)
	}

	for _, er := range ns.ExitReactions {
		state.ExitReactionTriggers = append(state.ExitReactionTriggers, er.Trigger)
	}
	states = append(states, state)

	// Collect the transitions.
	for _, t := range ns.Transitions {
		index := ft.transitionIndex
		ft.transitionIndex++
		transition := &frontend.TransitionData{
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
