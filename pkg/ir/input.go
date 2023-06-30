package ir

import (
	"fmt"
	"sort"

	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"github.com/bradenaw/juniper/xslices"
)

// inputHandler is a helper struct to keep running state while we process the frontend input.
type inputHandler struct {
	scdata *frontend.StatechartData

	triggers   []*Trigger
	rootStates []*State

	triggerMap map[string]*Trigger
	stateMap   map[string]*State
}

func ProcessStatechartData(scdata *frontend.StatechartData) (*Statechart, error) {
	ih := inputHandler{
		scdata:     scdata,
		triggerMap: make(map[string]*Trigger),
		stateMap:   make(map[string]*State),
	}

	if err := ih.collectTriggers(); err != nil {
		return nil, fmt.Errorf("collecting triggers: %w", err)
	}

	if err := ih.collectStates(); err != nil {
		return nil, fmt.Errorf("collecting states: %w", err)
	}

	if err := ih.collectTransitions(); err != nil {
		return nil, fmt.Errorf("collecting transitions: %w", err)
	}

	// Collect the states as defined in the order of the frontend.
	states := make([]*State, 0, len(ih.stateMap))
	for _, state := range ih.stateMap {
		states = append(states, state)
	}

	sort.Slice(states, func(i, j int) bool {
		return states[i].frontendData.Index < states[j].frontendData.Index
	})

	return &Statechart{
		Name:         scdata.Name,
		Roots:        ih.rootStates,
		Triggers:     ih.triggers,
		States:       states,
		TriggerMap:   ih.triggerMap,
		StateMap:     ih.stateMap,
		frontendData: scdata,
	}, nil
}

func (ih *inputHandler) collectTriggers() error {
	triggers := make([]*Trigger, 0, len(ih.scdata.Triggers))

	for _, tdata := range ih.scdata.Triggers {
		trigger, err := ih.createTrigger(tdata)
		if err != nil {
			return fmt.Errorf("creating trigger %q: %w", tdata.Name, err)
		}

		triggers = append(triggers, trigger)
		ih.triggerMap[trigger.Name] = trigger
	}

	ih.triggers = triggers
	return nil
}

func (ih *inputHandler) createTrigger(tdata *frontend.TriggerData) (*Trigger, error) {
	// We make sure that the trigger doesn't exist already.
	if _, ok := ih.triggerMap[tdata.Name]; ok {
		return nil, fmt.Errorf("trigger %q defined twice", tdata.Name)
	}

	// Parse the arguments.
	// For now we only support C++, but we could support other languages as well if needed.
	args, err := ParseCppArguments(tdata.ArgumentsString)
	if err != nil {
		return nil, fmt.Errorf("parsing arguments for trigger %q: %w", tdata.Name, err)
	}

	return &Trigger{
		Name:         tdata.Name,
		Args:         args,
		frontendData: tdata,
	}, nil
}

func (ih *inputHandler) collectStates() error {
	// We first create all the states and track its associated data.
	stateMap := make(map[string]*State)
	for _, statedata := range ih.scdata.States {
		// The state should not exist.
		if _, ok := stateMap[statedata.Name]; ok {
			return fmt.Errorf("state %q already exists", statedata.Name)
		}

		// For now, we simply create the state. Parenthood will be set on a second pass.
		state := &State{
			Name:         statedata.Name,
			frontendData: statedata,
		}
		stateMap[statedata.Name] = state
	}

	// Now we check for parenthood.
	var roots []*State
	for name, state := range stateMap {
		// If the parent name is null, it means that this is a root state.
		if state.frontendData.Parent == "" {
			roots = append(roots, state)
			continue
		}

		// Search for the parent and mark it as a child of the other.
		parent, ok := stateMap[state.frontendData.Parent]
		if !ok {
			return fmt.Errorf("state %q has unexistent parent state %q", name, state.frontendData.Parent)
		}

		// The parent should not have have this state already.
		// We mark the parent <=> child relationship.
		if parent.IsParentOf(state) {
			return fmt.Errorf("state %q already has state %q as child", parent.Name, state.Name)
		}
		parent.States = append(parent.States, state)
		state.Parent = parent
	}

	// We collect the transition reactions.
	for name, state := range stateMap {
		// Collect the enter reactions.
		state.DefaultEnter = state.frontendData.DefaultEnter
		enters, err := ih.collectReactions(state.frontendData.EnterReactionTriggers)
		if err != nil {
			return fmt.Errorf("state %q: collecting enter reactions: %w", name, err)
		}
		state.EnterReactions = enters

		// Collect the exit reactions.
		state.DefaultExit = state.frontendData.DefaultExit
		exits, err := ih.collectReactions(state.frontendData.ExitReactionTriggers)
		if err != nil {
			return fmt.Errorf("state %q: collecting exit reactions: %w", name, err)
		}
		state.ExitReactions = exits
	}

	ih.rootStates = roots
	ih.stateMap = stateMap

	return nil
}

func (ih *inputHandler) collectReactions(triggerNames []string) ([]*TransitionReaction, error) {
	var triggers []*Trigger
	for _, triggerName := range triggerNames {
		trigger, ok := ih.triggerMap[triggerName]
		if !ok {
			return nil, fmt.Errorf("cannot find trigger %q", triggerName)
		}
		triggers = append(triggers, trigger)
	}

	return xslices.Map(triggers, func(t *Trigger) *TransitionReaction {
		return &TransitionReaction{
			Trigger: t,
		}
	}), nil
}

func (ih *inputHandler) collectTransitions() error {
	// We go over all the transitions and generate the actual mapping.
	for _, tdata := range ih.scdata.Transitions {
		transition, fromState, err := ih.createTransition(tdata)
		if err != nil {
			return fmt.Errorf("creating transition %q: %w", tdata.String(), err)
		}

		fromState.Transitions = append(fromState.Transitions, transition)
	}

	return nil
}

// createTransition returns a new transition, as well as the state it stems from.
func (ih *inputHandler) createTransition(tdata *frontend.TransitionData) (*Transition, *State, error) {
	from, ok := ih.stateMap[tdata.From]
	if !ok {
		return nil, nil, fmt.Errorf("cannot find from state %q", tdata.From)
	}

	to, ok := ih.stateMap[tdata.To]
	if !ok {
		return nil, nil, fmt.Errorf("cannot find to state %q", tdata.To)
	}

	// See if there is a trigger available (if not, it's a null transition).
	var trigger *Trigger
	if tdata.Trigger != "" {
		t, ok := ih.triggerMap[tdata.Trigger]
		if !ok {
			return nil, nil, fmt.Errorf("cannot find trigger %q", tdata.Trigger)
		}
		trigger = t
	}

	transition := &Transition{
		From:         from,
		To:           to,
		Trigger:      trigger,
		frontendData: tdata,
	}

	return transition, from, nil
}
