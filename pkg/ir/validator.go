package ir

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/errors"
	"github.com/cristiandonosoc/gochart/internal/files"
)

// validate is called once the complete statechart has been parsed and collected.
// This means that all options have already been calculated and are certain that the statechart
// description is not going to change anymore.
func validate(ih *inputHandler) error {
	// Validate top level statechart.
	if err := validateTopLevelStatechart(ih); err != nil {
		return fmt.Errorf("validating top level statechart: %w", err)
	}

	// Validate states.
	for _, state := range ih.states {
		sv := newStateValidator(ih, state)
		if err := sv.validate(); err != nil {
			return fmt.Errorf("validating state %q: %w", state.Name, err)
		}
	}

	return nil
}

func validateTopLevelStatechart(ih *inputHandler) error {
	if len(ih.rootStates) == 0 {
		return errors.Errorf(ih.scdata.Frontend_NamePos, "Semantic", "no root states found")
	}

	if err := validateDefaultChildExists(ih.scdata.Frontend_NamePos, ih.rootStates); err != nil {
		return fmt.Errorf("validating that default child exists: %w", err)
	}

	return nil
}

type stateValidator struct {
	ih         *inputHandler
	state      *State
	triggerSet map[*Trigger]struct{}
}

func newStateValidator(ih *inputHandler, state *State) *stateValidator {
	return &stateValidator{
		ih:         ih,
		state:      state,
		triggerSet: make(map[*Trigger]struct{}),
	}
}

func (sv *stateValidator) validate() error {
	// If it has children, at least one of them has to be marked default child.
	if len(sv.state.Children) > 0 {
		if err := validateDefaultChildExists(sv.state.frontendData.Frontend_NamePos, sv.state.Children); err != nil {
			return fmt.Errorf("state %q: validating that default child exists: %w", sv.state.Name, err)
		}
	}

	if err := sv.validateTransitions(); err != nil {
		return fmt.Errorf("validating transitions: %w", err)
	}

	if err := sv.validateCaptures(); err != nil {
		return fmt.Errorf("validating captures: %w", err)
	}

	return nil
}

func (sv *stateValidator) validateTransitions() error {
	for _, transition := range sv.state.Transitions {
		// Validate that this trigger hasn't been used yet.
		if _, ok := sv.triggerSet[transition.Trigger]; ok {
			return errors.Errorf(transition.frontendData.Frontend_TriggerPos, "Semantic",
				"state %q: Trigger %q already handled by transition or capture",
				sv.state.Name, transition.Trigger.Name)
		}
		sv.triggerSet[transition.Trigger] = struct{}{}

		// Validate that the transition has the same concurrent group.
		if sv.state.ConcurrentGroup != transition.To.ConcurrentGroup {
			return errors.Errorf(transition.frontendData.Frontend_TargetPos, "Semantic",
				"state %q can only transition to states within its own concurrent group, under %s",
				sv.state.Name, sv.ih.getConcurrentGroupTitle(sv.state.ConcurrentGroup))
		}

		if transition.IsNullTransition() {
			if !transition.HasCondition() {
				// A state cannot have a non-conditional null transition and children state.
				if len(sv.state.Children) > 0 {
					return errors.Errorf(transition.frontendData.Frontend_TriggerPos, "Semantic",
						"state %q cannot have a null transition and children states",
						sv.state.Name)
				}

				// A non-conditional null transition should be alone.
				if len(sv.state.Transitions) > 1 {
					return errors.Errorf(sv.state.frontendData.Frontend_NamePos, "Semantic",
						"state %q has a non-conditional null transition. It can be the only transition on the state.",
						sv.state.Name)
				}
			}
		}
	}

	return nil
}

func (sv *stateValidator) validateCaptures() error {
	for _, capture := range sv.state.Captures {
		if _, ok := sv.triggerSet[capture.Trigger]; ok {
			return errors.Errorf(capture.frontendData.Frontend_TriggerPos, "Semantic",
				"state %q: Trigger %q already handled by transition or capture",
				sv.state.Name, capture.Trigger.Name)
		}
		sv.triggerSet[capture.Trigger] = struct{}{}
	}

	return nil
}

func validateDefaultChildExists(parentPos *files.LoadedFilePosition, states []*State) error {
	foundDefault := false
	for _, state := range states {
		if state.Default {
			// There can only be one default child state.
			if foundDefault {
				return errors.Errorf(state.frontendData.Frontend_NamePos, "Semantic",
					"state %q marked as default when a default state already exists",
					state.Name)
			}
			foundDefault = true
		}
	}

	if !foundDefault {
		return errors.Errorf(parentPos, "Semantic", "no default child state found")
	}

	return nil
}
