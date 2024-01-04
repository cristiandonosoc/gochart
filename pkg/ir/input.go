package ir

import (
	"container/list"
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/errors"
	"github.com/cristiandonosoc/gochart/internal/files"
	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

// inputHandler is a helper struct to keep running state while we process the frontend input.
type inputHandler struct {
	gcdata *frontend_common.GochartData

	triggers []*Trigger
	root     *State
	states   []*State

	nextConcurrentGroup int
	concurrentHeads     []*ConcurrentHead

	triggerMap map[string]*Trigger
	stateMap   map[string]*State
}

func ProcessGochartData(gcdata *frontend_common.GochartData) (*Gochart, error) {
	gc, err := processGochartDataInternal(gcdata)
	if err != nil {
		return nil, errors.NewErrorContainer([]error{err})
	}
	return gc, nil
}

func processGochartDataInternal(gcdata *frontend_common.GochartData) (*Gochart, error) {
	ih := inputHandler{
		gcdata:     gcdata,
		triggerMap: make(map[string]*Trigger),
		stateMap:   make(map[string]*State),
	}

	// We create the base root state.
	root := &State{
		Name:         gcdata.Root.Name,
		FrontendData: gcdata.Root,
	}

	// We create the base state object (will be filled more later).
	statechart := &Gochart{
		Name: gcdata.Name,
		// By default the namespace is same as the name. This might be override by an option.
		Namespace:    gcdata.Name,
		Root:         root,
		FrontendData: gcdata,
	}
	if err := parseGochartOptions(statechart, gcdata.Options); err != nil {
		return nil, fmt.Errorf("statechart %q: parsing options: %w", gcdata.Name, err)
	}

	if err := ih.collectTriggers(); err != nil {
		return nil, fmt.Errorf("collecting triggers: %w", err)
	}

	if err := ih.collectStates(root); err != nil {
		return nil, fmt.Errorf("collecting states: %w", err)
	}

	if err := ih.collectTransitions(); err != nil {
		return nil, fmt.Errorf("collecting transitions: %w", err)
	}

	if err := validate(&ih); err != nil {
		return nil, fmt.Errorf("validating input: %w", err)
	}

	statechart.ConcurrentHeads = ih.concurrentHeads
	statechart.Triggers = ih.triggers
	statechart.States = ih.states
	statechart.TriggerMap = ih.triggerMap
	statechart.StateMap = ih.stateMap

	// Now that we have the statechart, we go over each state creating the back pointer.
	for _, state := range statechart.States {
		state.statechart = statechart
	}

	return statechart, nil
}

func (ih *inputHandler) collectTriggers() error {
	triggers := make([]*Trigger, 0, len(ih.gcdata.Triggers))

	for _, tdata := range ih.gcdata.Triggers {
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

func (ih *inputHandler) createTrigger(tdata *frontend_common.TriggerData) (*Trigger, error) {
	// We make sure that the trigger doesn't exist already.
	if _, ok := ih.triggerMap[tdata.Name]; ok {
		return nil, errors.Errorf(tdata.Frontend_NamePos, "Semantic", "trigger %q defined more than once", tdata.Name)
	}

	// Parse the arguments.
	var args []*TriggerArgument
	if tdata.ArgsString != "" {
		// For now we only support C++, but we could support other languages as well if needed.
		parsedArgs, err := ParseCppArguments(tdata.ArgsString)
		if err != nil {
			return nil, errors.Errorf(tdata.Frontend_ArgsPos, "Semantic", "parsing arguments for trigger %q: %w", tdata.Name, err)
		}
		args = parsedArgs
	}

	return &Trigger{
		Name:         tdata.Name,
		Args:         args,
		FrontendData: tdata,
	}, nil
}

func (ih *inputHandler) collectStates(rootState *State) error {
	// We first create all the states and track its associated data.
	states := []*State{}
	stateMap := map[string]*State{}
	for _, statedata := range ih.gcdata.States {
		// The state should not exist.
		if _, ok := stateMap[statedata.Name]; ok {
			return errors.Errorf(statedata.Frontend_NamePos, "Semantic", "state %q already exists", statedata.Name)
		}

		// We special case the root.
		if statedata.Name == rootState.Name {
			states = append(states, rootState)
			stateMap[statedata.Name] = rootState
			continue
		}

		// For now, we simply create the state. Parenthood will be set on a second pass.
		state := &State{
			Name:         statedata.Name,
			FrontendData: statedata,
		}
		states = append(states, state)
		stateMap[statedata.Name] = state
	}

	// Now we check for parenthood.
	for _, state := range states {
		// We skip the root we just added.
		if state.Name == rootState.Name {
			continue
		}

		// Search for the parent and mark it as a child of the other.
		parent, ok := stateMap[state.FrontendData.Parent]
		if !ok {
			return errors.InternalErrorf("state %q has unexistent parent state %q", state.Name, state.FrontendData.Parent)
		}

		// The parent should not have have this state already.
		// We mark the parent <=> child relationship.
		if parent.IsParentOf(state) {
			return errors.InternalErrorf("state %q already has state %q as child", parent.Name, state.Name)
		}
		parent.Children = append(parent.Children, state)
		state.Parent = parent

	}

	// Collect extra data from the states.
	for name, state := range stateMap {
		// Parse the options for this state.
		if err := parseStateOptions(state, state.FrontendData.Options); err != nil {
			return fmt.Errorf("state %q: parsing options: %w", state.Name, err)
		}

		// Collect the enter reactions.
		enters, err := ih.collectReactions(state.FrontendData.EnterReactions, StateReactionKind_Enter)
		if err != nil {
			return fmt.Errorf("state %q: collecting enter reactions: %w", name, err)
		}
		state.EnterReactions = enters

		// Collect the exit reactions.
		exits, err := ih.collectReactions(state.FrontendData.ExitReactions, StateReactionKind_Exit)
		if err != nil {
			return fmt.Errorf("state %q: collecting exit reactions: %w", name, err)
		}
		state.ExitReactions = exits

		// Collect the captures.
		captures, err := ih.collectCaptures(state.FrontendData.Captures)
		if err != nil {
			return fmt.Errorf("state %q: collecting captures: %w", name, err)
		}
		state.Captures = captures
	}

	// Now that we have a pass of the state, we can set some options (such as default child).
	for _, state := range stateMap {
		// For each state, if there is no default child, the first child should be default.
		// TODO(cdc): Make the first-child-as-default heuristic toggable via options.
		prefix := fmt.Sprintf("state %q", state.Name)
		if err := ih.determineDefaultChild(prefix, state.FrontendData.Frontend_NamePos, state.Children); err != nil {
			return fmt.Errorf("determining default child: %w", err)
		}
	}

	// Calculate concurrent-ness.
	// This is better if calculated top down.
	rootState.ConcurrentGroup = 0
	ih.concurrentHeads = append(ih.concurrentHeads, &ConcurrentHead{
		Id:        0,
		HeadState: rootState,
	})
	if err := ih.calculateConcurrentness(rootState); err != nil {
		return fmt.Errorf("calculating concurrentness for root state %q: %w", rootState.Name, err)
	}

	ih.root = rootState
	// ih.rootStates = roots
	ih.states = states
	ih.stateMap = stateMap

	return nil
}

// Separate function for validating child state, as this is also used at the root level.
func (ih *inputHandler) determineDefaultChild(parentPrefix string, pos *files.LoadedFilePosition, states []*State) error {
	// No children, there is nothing to do.
	if len(states) == 0 {
		return nil
	}

	stateHasDefaultOption := func(state *State) bool {
		for _, option := range state.Options {
			if option.Name() == kStateOption_Default {
				return true
			}
		}
		return false
	}

	// We see if a state defined a default option.
	// If so, no automatic handling is to be done.
	defaultOptionFound := false
	for _, state := range states {
		if stateHasDefaultOption(state) {
			defaultOptionFound = true
		}
	}

	// If no option was set, we simply select the first state as default.
	if !defaultOptionFound {
		states[0].Default = true
	}

	return nil
}

func (ih *inputHandler) collectReactions(frontendReactions []*frontend_common.ReactionData, kind StateReactionKind) ([]*StateReaction, error) {
	defaultFound := false

	reactions := make([]*StateReaction, 0, len(frontendReactions))
	for _, fr := range frontendReactions {
		var trigger *Trigger

		// Special case the empty trigger name, which is a default reaction with a nil trigger.
		if fr.Trigger == "" {
			if defaultFound {
				return nil, errors.Errorf(fr.Frontend_KeywordPos, "Semantic", "many default %q reactions found, only one allowed", kind)
			}

			defaultFound = true
		} else {
			tr, ok := ih.triggerMap[fr.Trigger]
			if !ok {
				return nil, errors.Errorf(fr.Frontend_TriggerPos, "Semantic", "cannot find trigger %q", fr.Trigger)
			}
			trigger = tr
		}

		reactions = append(reactions, &StateReaction{
			Trigger:      trigger,
			Kind:         kind,
			FrontendData: fr,
		})
	}

	return reactions, nil
}

func (ih *inputHandler) collectCaptures(frontendCaptures []*frontend_common.CaptureData) ([]*StateCapture, error) {
	captures := make([]*StateCapture, 0, len(frontendCaptures))

	for _, fc := range frontendCaptures {
		trigger, ok := ih.triggerMap[fc.Trigger]
		if !ok {
			return nil, errors.Errorf(fc.Frontend_TriggerPos, "Semantic", "cannot find trigger %q", fc.Trigger)
		}

		captures = append(captures, &StateCapture{
			Trigger:      trigger,
			Condition:    fc.Condition,
			FrontendData: fc,
		})
	}

	return captures, nil
}

func (ih *inputHandler) collectTransitions() error {
	// We go over all the transitions and generate the actual mapping.
	for _, tdata := range ih.gcdata.Transitions {
		transition, fromState, err := ih.createTransition(tdata)
		if err != nil {
			return fmt.Errorf("creating transition %q: %w", tdata.String(), err)
		}

		fromState.Transitions = append(fromState.Transitions, transition)
	}

	return nil
}

// createTransition returns a new transition, as well as the state it stems from.
func (ih *inputHandler) createTransition(tdata *frontend_common.TransitionData) (*Transition, *State, error) {
	from, ok := ih.stateMap[tdata.From]
	if !ok {
		return nil, nil, errors.InternalErrorf("cannot find from state %q", tdata.From)
	}

	to, ok := ih.stateMap[tdata.To]
	if !ok {
		return nil, nil, errors.Errorf(tdata.Frontend_TargetPos, "Semantic", "cannot find target state %q", tdata.To)
	}

	// See if there is a trigger available (if not, it's a null transition).
	var trigger *Trigger
	if tdata.Trigger != "" {
		t, ok := ih.triggerMap[tdata.Trigger]
		if !ok {
			return nil, nil, errors.Errorf(tdata.Frontend_TriggerPos, "Semantic", "cannot find trigger %q", tdata.Trigger)
		}
		trigger = t
	}

	transition := &Transition{
		From:         from,
		To:           to,
		Trigger:      trigger,
		Condition:    tdata.Condition,
		Action:       tdata.Action,
		FrontendData: tdata,
	}

	return transition, from, nil
}

func (ih *inputHandler) calculateConcurrentness(state *State) error {
	return ih.calculateConcurrentnessDFS(state)
}

// calculateConcurrentnessDFS traverse the statechart tree in BFS order, calculating concurrent
// groups as it finds them.
func (ih *inputHandler) calculateConcurrentnessDFS(state *State) error {
	for _, child := range state.Children {
		// If it's not a concurrent state, we use the parent concurrent group.
		// Otherwise we create a new group for this child.
		group := state.ConcurrentGroup
		if state.Concurrent {
			ih.nextConcurrentGroup++
			group = ih.nextConcurrentGroup
			ih.concurrentHeads = append(ih.concurrentHeads, &ConcurrentHead{
				Id:        group,
				HeadState: child,
			})
		}
		child.ConcurrentGroup = group

		// We iterate over the children immediatelly to perform the DFS search.
		if err := ih.calculateConcurrentnessDFS(child); err != nil {
			return fmt.Errorf("calculating concurrent group for state %q: %w", child.SafeName(), err)
		}
	}

	return nil
}

// calculateConcurrentnessBFS traverse the statechart tree in BFS order, calculating concurrent
// groups as it finds them.
func (ih *inputHandler) calculateConcurrentnessBFS(state *State) error {
	// Create a queue with states to calculate.
	queue := list.New()
	for _, child := range state.Children {
		queue.PushBack(child)
	}

	for queue.Len() > 0 {
		current, ok := queue.Remove(queue.Front()).(*State)
		if !ok {
			panic("error casting state to State")
		}
		parent := current.Parent
		if parent == nil {
			return fmt.Errorf("state %q has a nil parent", current.Name)
		}

		// If we're not a concurrent state, we pass use the parent concurrent Group.
		// Otherwise, we create a new concurrent group for each child.
		group := parent.ConcurrentGroup
		if parent.Concurrent {
			ih.nextConcurrentGroup++
			group = ih.nextConcurrentGroup
			ih.concurrentHeads = append(ih.concurrentHeads, &ConcurrentHead{
				Id:        group,
				HeadState: current,
			})
		}
		current.ConcurrentGroup = group

		// Push the children of this state into the queue.
		for _, child := range current.Children {
			queue.PushBack(child)
		}
	}

	return nil
}

func (ih *inputHandler) getConcurrentGroupTitle(concurrentGroup int) string {
	head := ih.concurrentHeads[concurrentGroup]

	// If the state is nil, it means that this is a root state.
	if head.HeadState == nil {
		return fmt.Sprintf("statechart %q", ih.gcdata.Name)
	}

	return fmt.Sprintf("state %q", head.HeadState.Name)
}
