// Package ir holds the intermediate representation of the Statechart, which will then permit us to
// generate representations from it (eg. to C++).
package ir

import (
	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

// Statechart represents a single statechart state machine.
type Statechart struct {
	Name     string
	Triggers map[string]*Trigger

	// Roots are the list of states that are "roots" (have no parents) within the state tree.
	Roots []*State

	// States are all the states, as defined in the order from the frontend.
	States []*State

	StateMap     map[string]*State
	frontendData *frontend.StatechartData
}

// TRIGGER -----------------------------------------------------------------------------------------

type Trigger struct {
	Name string
	Args []*TriggerArgument

	frontendData *frontend.TriggerData
}

type TriggerArgument struct {
	Type string
	Name string
}

// STATE -------------------------------------------------------------------------------------------

// State represents a single state withing a statechart.
type State struct {
	Name string

	// States represents the substates that this state has.
	States      []*State
	Transitions []*Transition

	DefaultEnter   bool
	EnterReactions []*TransitionReaction

	DefaultExit   bool
	ExitReactions []*TransitionReaction

	parent       *State
	frontendData *frontend.StateData
}

func (s *State) Equals(other *State) bool {
	return s.Name == other.Name
}

func (s *State) IsLeaf() bool {
	return len(s.States) == 0
}

func (s *State) Contains(other *State) bool {
	for _, child := range s.States {
		if child.Equals(other) {
			return true
		}
	}

	return false
}

type TransitionReaction struct {
	Trigger *Trigger
}

// TRANSITION --------------------------------------------------------------------------------------

// Transition represents a transition from one state to another given a particular trigger.
type Transition struct {
	From    *State
	To      *State
	Trigger *Trigger

	frontendData *frontend.TransitionData
}

func (t *Transition) IsNullTransition() bool {
	return t.Trigger == nil
}
