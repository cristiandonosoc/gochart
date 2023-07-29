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

	if err := validateDefaultChildExists(ih.rootStates); err != nil {
		return fmt.Errorf("validating that default child exists: %w", err)
	}

	return nil
}

func validateState(state *State) error {
	// If it has children, at least one of them has to be marked default child.
	if len(state.Children) > 0 {
		if err := validateDefaultChildExists(state.Children); err != nil {
			return fmt.Errorf("validating that default child exists: %w", err)
		}
	}

	return nil
}

func validateDefaultChildExists(states []*State) error {
	foundDefault := false
	for _, state := range states {
		if state.Default {
			// There can only be one default child state.
			if foundDefault {
				return fmt.Errorf("state %q marked as default when a default state already exists", state.Name)
			}
			foundDefault = true
		}
	}

	if !foundDefault {
		return fmt.Errorf("no default child state found")
	}

	return nil
}
