package ir

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

// inputHandler is a helper struct to keep running state while we process the frontend input.
type inputHandler struct {
	scdata *output.StatechartData

	triggers   []*Trigger
	rootStates []*State
	states     []*State

	triggerMap map[string]*Trigger
	stateMap   map[string]*State
}

func ProcessStatechartData(scdata *output.StatechartData) (*Statechart, error) {
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

	if err := validate(&ih); err != nil {
		return nil, fmt.Errorf("validating input: %w", err)
	}

	return &Statechart{
		Name:         scdata.Name,
		Roots:        ih.rootStates,
		Triggers:     ih.triggers,
		States:       ih.states,
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

func (ih *inputHandler) createTrigger(tdata *output.TriggerData) (*Trigger, error) {
	// We make sure that the trigger doesn't exist already.
	if _, ok := ih.triggerMap[tdata.Name]; ok {
		return nil, fmt.Errorf("trigger %q defined twice", tdata.Name)
	}

	// Parse the arguments.
	var args []*TriggerArgument
	if tdata.ArgsString != "" {
		// For now we only support C++, but we could support other languages as well if needed.
		parsedArgs, err := ParseCppArguments(tdata.ArgsString)
		if err != nil {
			return nil, fmt.Errorf("parsing arguments for trigger %q: %w", tdata.Name, err)
		}
		args = parsedArgs
	}

	return &Trigger{
		Name:         tdata.Name,
		Args:         args,
		frontendData: tdata,
	}, nil
}

func (ih *inputHandler) collectStates() error {
	// We first create all the states and track its associated data.
	var states []*State
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
		states = append(states, state)
		stateMap[statedata.Name] = state
	}

	// Now we check for parenthood.
	var roots []*State
	// for name, state := range stateMap {
	for _, state := range states {
		// If the parent name is null, it means that this is a root state.
		if state.frontendData.Parent == "" {
			roots = append(roots, state)
			continue
		}

		// Search for the parent and mark it as a child of the other.
		parent, ok := stateMap[state.frontendData.Parent]
		if !ok {
			return fmt.Errorf("state %q has unexistent parent state %q", state.Name, state.frontendData.Parent)
		}

		// The parent should not have have this state already.
		// We mark the parent <=> child relationship.
		if parent.IsParentOf(state) {
			return fmt.Errorf("state %q already has state %q as child", parent.Name, state.Name)
		}
		parent.Children = append(parent.Children, state)
		state.Parent = parent

	}

	// Collect extra data from the states.
	for name, state := range stateMap {
		// Parse the options for this state.
		if err := parseStateOptions(state, state.frontendData.Options); err != nil {
			return fmt.Errorf("state %q: parsing options: %w", state.Name, err)
		}

		// Collect the enter reactions.
		enters, err := ih.collectReactions(state.frontendData.EnterReactions, StateReactionKind_Enter)
		if err != nil {
			return fmt.Errorf("state %q: collecting enter reactions: %w", name, err)
		}
		state.EnterReactions = enters

		// Collect the exit reactions.
		exits, err := ih.collectReactions(state.frontendData.ExitReactions, StateReactionKind_Exit)
		if err != nil {
			return fmt.Errorf("state %q: collecting exit reactions: %w", name, err)
		}
		state.ExitReactions = exits

		// Collect the captures.
		captures, err := ih.collectCaptures(state.frontendData.Captures)
		if err != nil {
			return fmt.Errorf("state %q: collecting captures: %w", name, err)
		}
		state.Captures = captures
	}

	// Now that we have a pass of the state, we can set some options (such as default child).
	for _, state := range stateMap {
		// For each state, if there is no default child, the first child should be default.
		// TODO(cdc): Make the first-child-as-default heuristic toggable via options.
		if err := ih.determineDefaultChild(state.Name, state.Children); err != nil {
			return fmt.Errorf("determining default child: %w", err)
		}
	}

	// Perform some setup at the root level as well.
	{
		if err := ih.determineDefaultChild("Statechart Root", roots); err != nil {
			return fmt.Errorf("determining default child: %w", err)
		}
	}

	ih.rootStates = roots
	ih.states = states
	ih.stateMap = stateMap

	return nil
}

// Separate function for validating child state, as this is also used at the root level.
func (ih *inputHandler) determineDefaultChild(parentName string, states []*State) error {
	// No children, there is nothing to do.
	if len(states) == 0 {
		return nil
	}

	defaultFound := false
	for _, state := range states {
		if state.Default {
			if defaultFound {
				return fmt.Errorf("state %q has more than one default child", parentName)
			}

			defaultFound = true
		}
	}
	if !defaultFound {
		states[0].Default = true
	}

	return nil
}

func (ih *inputHandler) collectReactions(triggerNames []string, kind StateReactionKind) ([]*StateReaction, error) {
	defaultFound := false

	var triggers []*Trigger
	for _, triggerName := range triggerNames {
		// Special case the empty trigger name, which is a default reaction with a nil trigger.
		if triggerName == "" {
			if defaultFound {
				return nil, fmt.Errorf("two default reactions found, only one allowed")
			}

			triggers = append(triggers, nil)
			defaultFound = true
			continue
		}

		trigger, ok := ih.triggerMap[triggerName]
		if !ok {
			return nil, fmt.Errorf("cannot find trigger %q", triggerName)
		}
		triggers = append(triggers, trigger)
	}

	reactions := make([]*StateReaction, 0, len(triggers))
	for _, trigger := range triggers {
		reactions = append(reactions, &StateReaction{
			Trigger: trigger,
			Kind:    kind,
		})
	}

	return reactions, nil
}

func (ih *inputHandler) collectCaptures(frontendCaptures []*output.CaptureData) ([]*StateCapture, error) {
	captures := make([]*StateCapture, 0, len(frontendCaptures))

	for _, fc := range frontendCaptures {
		trigger, ok := ih.triggerMap[fc.Trigger]
		if !ok {
			return nil, fmt.Errorf("cannot find trigger %q", fc.Trigger)
		}

		captures = append(captures, &StateCapture{
			Trigger:   trigger,
			Condition: fc.Condition,
		})
	}

	return captures, nil
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
func (ih *inputHandler) createTransition(tdata *output.TransitionData) (*Transition, *State, error) {
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
