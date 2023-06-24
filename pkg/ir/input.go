package ir

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

func InputStatechartData(scdata *frontend.StatechartData) (*Statechart, error) {
	rootStates, allStates, err := createStates(scdata)
	if err != nil {
		return nil, fmt.Errorf("creating states: %w", err)
	}

	// TODO(cdc): Transitions.
	// TODO(cdc): Triggers.

	return &Statechart{
		Name:         scdata.Name,
		Roots:        rootStates,
		states:       allStates,
		frontendData: scdata,
	}, nil
}

func createStates(scdata *frontend.StatechartData) (rootStates []*State, allStates map[string]*State, err error) {
	// We first create all the states and track its associated data.

	var stateMap map[string]*State
	for _, statedata := range scdata.States {
		// The state should not exist.
		if _, ok := stateMap[statedata.Name]; ok {
			return nil, nil, fmt.Errorf("state %q already exists", statedata.Name)
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
			return nil, nil, fmt.Errorf("state %q has unexistent parent state %q", name, state.frontendData.Parent)
		}

		// The parent should not have have this state already.
		// We mark the parent <=> child relationship.
		if parent.Contains(state) {
			return nil, nil, fmt.Errorf("state %q already has state %q as child", parent.Name, state.Name)
		}
		parent.States = append(parent.States, state)
		state.parent = parent
	}

	return roots, stateMap, nil
}
