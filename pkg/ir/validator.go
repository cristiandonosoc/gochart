package ir

import (
	"fmt"
	"sort"

	"github.com/cristiandonosoc/gochart/internal/errors"
	"github.com/cristiandonosoc/golib/pkg/files"
)

// validate is called once the complete gochart has been parsed and collected.
// This means that all options have already been calculated and are certain that the gochart
// description is not going to change anymore.
func validate(ih *inputHandler) error {
	if len(ih.states) < 2 {
		return errors.Errorf(ih.gcdata.Frontend_NamePos, "Semantic", "no root states found")
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

type stateValidator struct {
	ih    *inputHandler
	state *State
}

func newStateValidator(ih *inputHandler, state *State) *stateValidator {
	return &stateValidator{
		ih:    ih,
		state: state,
	}
}

func (sv *stateValidator) validate() error {
	// If it has children, at least one of them has to be marked default child.
	if len(sv.state.Children) > 0 {
		if err := validateDefaultChildExists(sv.state.FrontendData.Frontend_NamePos, sv.state.Children); err != nil {
			return fmt.Errorf("state %q: validating that default child exists: %w", sv.state.Name, err)
		}
	}

	if err := sv.validateTransitions(); err != nil {
		return fmt.Errorf("validating transitions: %w", err)
	}

	if err := sv.validateCaptures(); err != nil {
		return fmt.Errorf("validating captures: %w", err)
	}

	if err := sv.validateTriggerBasedEntries(); err != nil {
		return fmt.Errorf("validating trigger based entries: %w", err)
	}

	return nil
}

func (sv *stateValidator) validateTransitions() error {
	for _, transition := range sv.state.Transitions {
		// Validate that the transition has the same parallel group.
		if sv.state.ParallelGroup != transition.To.ParallelGroup {
			return errors.Errorf(transition.FrontendData.Frontend_TargetPos, "Semantic",
				"state %q can only transition to states within its own parallel group, under %s",
				sv.state.Name, sv.ih.getParallelGroupTitle(sv.state.ParallelGroup))
		}
	}

	return nil
}

func (sv *stateValidator) validateCaptures() error {
	// Code for validating captures per se.

	return nil
}

type triggerBasedEntry struct {
	index       int
	conditional bool

	kind     string
	position *files.LoadedFilePosition
}

type triggerBasedTracker struct {
	entries         map[*Trigger][]*triggerBasedEntry
	transitionCount int
	captureCount    int
}

func newTriggerBasedTracker() *triggerBasedTracker {
	return &triggerBasedTracker{
		entries: make(map[*Trigger][]*triggerBasedEntry),
	}
}

func (tbt *triggerBasedTracker) appendTransition(transition *Transition) {
	// Ensure the list exists.
	if _, ok := tbt.entries[transition.Trigger]; !ok {
		tbt.entries[transition.Trigger] = []*triggerBasedEntry{}
	}

	tbt.entries[transition.Trigger] = append(tbt.entries[transition.Trigger], &triggerBasedEntry{
		index:       transition.FrontendData.Frontend_Index,
		conditional: transition.HasCondition(),
		kind:        "transition",
		position:    transition.FrontendData.Frontend_TriggerPos,
	})
	tbt.transitionCount++
}

func (tbt *triggerBasedTracker) appendCapture(capture *StateCapture) {
	// Ensure the list exists.
	if _, ok := tbt.entries[capture.Trigger]; !ok {
		tbt.entries[capture.Trigger] = []*triggerBasedEntry{}
	}

	tbt.entries[capture.Trigger] = append(tbt.entries[capture.Trigger], &triggerBasedEntry{
		index:       capture.FrontendData.Frontend_Index,
		conditional: capture.HasCondition(),
		kind:        "capture",
		position:    capture.FrontendData.Frontend_TriggerPos,
	})
	tbt.captureCount++
}

func (tbt *triggerBasedTracker) sort() {
	for trigger := range tbt.entries {
		sort.Slice(tbt.entries[trigger], func(i, j int) bool {
			return tbt.entries[trigger][i].index < tbt.entries[trigger][j].index
		})
	}
}

func (sv *stateValidator) validateTriggerBasedEntries() error {
	// Collect all trigger based entries, based on trigger.
	tbt := newTriggerBasedTracker()

	for _, transition := range sv.state.Transitions {
		tbt.appendTransition(transition)
	}
	for _, capture := range sv.state.Captures {
		tbt.appendCapture(capture)
	}
	tbt.sort()

	// Now that we have them sorted, we can create validation per list.
	for trigger, triggerBasedList := range tbt.entries {
		for i, tbe := range triggerBasedList {
			// If it is not conditional, then it has to be the last.
			if !tbe.conditional {
				if i != len(triggerBasedList)-1 {
					return errors.Errorf(tbe.position, "Semantic",
						"state %q: %s for trigger %q is non-conditional, it must be the last for this trigger",
						sv.state.Name, tbe.kind, trigger.SafeName())
				}
			}
		}
	}

	// We do a special case for the null transition case.
	for _, tbe := range tbt.entries[nil] {
		// Captures should not be able to capture null triggers.
		if tbe.kind == "capture" {
			return errors.InternalErrorf(
				"State %q: detected capture for null trigger. This should not happen.",
				sv.state.Name)
		}

		// Non-conditional null transitions should be alone.
		if !tbe.conditional {
			if tbt.transitionCount+tbt.captureCount > 1 {
				return errors.Errorf(tbe.position, "Semantic",
					"state %q has a non-conditional null transition. It can be the only transition or capture on the state.",
					sv.state.Name)
			}

			if len(sv.state.Children) > 0 {
				return errors.Errorf(tbe.position, "Semantic",
					"state %q cannot have a non-conditional null transition and children states (this null transition will never trigger)",
					sv.state.Name)
			}

		}
	}

	return nil
}

func validateDefaultChildExists(parentPos *files.LoadedFilePosition, states []*State) error {
	foundDefault := false
	for _, state := range states {
		if state.Default {
			// There can only be one default child state.
			if foundDefault {
				return errors.Errorf(state.FrontendData.Frontend_NamePos, "Semantic",
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
