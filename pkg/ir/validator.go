package ir

import (
	"fmt"
)

func validate(ih *inputHandler) error {
	// Validate top level statechart.
	if err := validateTopLevelStatechart(ih); err != nil {
		return fmt.Errorf("validating top level statechart: %w", err)
	}

	// Validate states.
	for _, state := range ih.states {
		if err := validateState(state); err != nil {
			return fmt.Errorf("validating state %q: %w", state.Name, err)
		}
	}

	return nil
}

func validateTopLevelStatechart(ih *inputHandler) error {
	if len(ih.rootStates) == 0 {
		return fmt.Errorf("no root states found")
	}

	if err := validateInitialExists(ih.rootStates); err != nil {
		return fmt.Errorf("validating that initial child exists: %w", err)
	}

	return nil
}

func validateState(state *State) error {
	// If it has children, at least one of them has to be marked initial.
	if len(state.States) > 0 {
		if err := validateInitialExists(state.States); err != nil {
			return fmt.Errorf("validating that initial child exists: %w", err)
		}
	}

	return nil
}

func validateInitialExists(states []*State) error {
	initial := false
	for _, state := range states {
		if state.Initial {
			// There can only be one initial state.
			if initial {
				return fmt.Errorf("state %q marked as initial when an initial state already exists", state.Name)
			}
			initial = true
		}
	}

	if !initial {
		return fmt.Errorf("no initial state found")
	}

	return nil
}
