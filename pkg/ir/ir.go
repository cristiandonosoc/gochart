// Package ir holds the intermediate representation of the Statechart, which will then permit us to
// generate representations from it (eg. to C++).
package ir

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

// Statechart represents a single statechart state machine.
type Statechart struct {
	Name string

	Triggers []*Trigger

	// Roots are the list of states that are "roots" (have no parents) within the state tree.
	Roots []*State

	// States are all the states, as defined in the order from the frontend.
	States []*State

	TriggerMap   map[string]*Trigger
	StateMap     map[string]*State
	frontendData *frontend.StatechartData
}

// TRIGGER -----------------------------------------------------------------------------------------

type Trigger struct {
	Name string
	Args []*TriggerArgument

	frontendData *frontend.TriggerData
}

func (t *Trigger) ArgsStringList() []string {
	strings := make([]string, 0, len(t.Args))
	for _, arg := range t.Args {
		strings = append(strings, arg.String())
	}

	return strings
}

// ArgsNameList returns a list with only the name of the arguments.
func (t *Trigger) ArgsNameList() []string {
	strings := make([]string, 0, len(t.Args))
	for _, arg := range t.Args {
		strings = append(strings, arg.Name)
	}

	return strings
}

type TriggerArgument struct {
	Type string
	Name string
}

func (ta *TriggerArgument) String() string {
	return fmt.Sprintf("%s %s", ta.Type, ta.Name)
}

// STATE -------------------------------------------------------------------------------------------

// State represents a single state withing a statechart.
type State struct {
	Name    string
	Initial bool

	// States represents the substates that this state has.
	Children    []*State
	Transitions []*Transition

	DefaultEnter   bool
	EnterReactions []*StateReaction

	DefaultExit   bool
	ExitReactions []*StateReaction

	Parent       *State
	frontendData *frontend.StateData
}

func (s *State) Equals(other *State) bool {
	return s.Name == other.Name
}

func (s *State) IsParentOf(other *State) bool {
	for _, child := range s.Children {
		if child.Equals(other) {
			return true
		}
	}

	return false
}

func (s *State) ParentName() string {
	if s.Parent != nil {
		return s.Parent.Name
	}
	return "None"
}

// STATE REACTION ----------------------------------------------------------------------------------

type StateReaction struct {
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
