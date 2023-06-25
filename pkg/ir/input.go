package ir

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

// inputHandler is a helper struct to keep running state while we process the frontend input.
type inputHandler struct {
	scdata *frontend.StatechartData

	rootStates []*State
	triggers   map[string]*Trigger

	allStates map[string]*State
}

func InputStatechartData(scdata *frontend.StatechartData) (*Statechart, error) {
	ih := inputHandler{
		scdata: scdata,
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

	return &Statechart{
		Name:         scdata.Name,
		Roots:        ih.rootStates,
		stateMap:     ih.allStates,
		frontendData: scdata,
	}, nil
}

func (ih *inputHandler) collectTriggers() error {
	for _, tdata := range ih.scdata.Triggers {
		trigger, err := ih.createTrigger(tdata)
		if err != nil {
			return fmt.Errorf("creating trigger %q: %w", tdata.Name, err)
		}

		ih.triggers[trigger.Name] = trigger
	}

	return nil
}

func (ih *inputHandler) createTrigger(tdata *frontend.TriggerData) (*Trigger, error) {
	// We make sure that the trigger doesn't exist already.
	if _, ok := ih.triggers[tdata.Name]; ok {
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
	var stateMap map[string]*State
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
		if parent.Contains(state) {
			return fmt.Errorf("state %q already has state %q as child", parent.Name, state.Name)
		}
		parent.States = append(parent.States, state)
		state.parent = parent
	}

	ih.rootStates = roots
	ih.allStates = stateMap

	return nil
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
	from, ok := ih.allStates[tdata.From]
	if !ok {
		return nil, nil, fmt.Errorf("cannot find from state %q", tdata.From)
	}

	to, ok := ih.allStates[tdata.To]
	if !ok {
		return nil, nil, fmt.Errorf("cannot find to state %q", tdata.To)
	}

	// See if there is a trigger available (if not, it's a null transition).
	var trigger *Trigger
	if tdata.Trigger != "" {
		t, ok := ih.triggers[tdata.Trigger]
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
