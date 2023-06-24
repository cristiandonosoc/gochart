package ir

import (
	"fmt"

)

type Validator struct {
	sc *Statechart

	stateCache []*State
	transitionCache []*Transition
}

func NewValidator(sc *Statechart) *Validator {
	// TODO(cdc): Collect all states.
	var states []*State

	// TODO(cdc): Collect all transitions.
	var transitions []*Transition

	return &Validator{
		sc: sc,
		stateCache: states,
		transitionCache: transitions,
	}
}

func (v *Validator) Validate() error {
	return fmt.Errorf("IMPLEMENT ME");
}

func ValidateStatecharts(statecharts []*Statechart) error {
	for _, sc := range statecharts {
		validator := NewValidator(sc)
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validating statechart %q: %w", sc.Name, err)
		}
	}

	return nil
}
