// Package ir holds the intermediate representation of the Statechart, which will then permit us to
// generate representations from it (eg. to C++).
package ir

import (
	"fmt"
	"log"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

const (
	InvalidTriggerName = "Invalid"
	InvalidStateName   = "Invalid"
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

func (sc *Statechart) DefaultRootState() *State {
	for _, state := range sc.States {
		if state.Default {
			return state
		}
	}

	panic("No default root state found. This should've been caught in validation")
}

// TRIGGER -----------------------------------------------------------------------------------------

type Trigger struct {
	Name string
	Args []*TriggerArgument

	frontendData *frontend.TriggerData
}

func (t *Trigger) SafeName() string {
	if t == nil {
		return InvalidTriggerName
	}

	return t.Name
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
	Default bool

	// States represents the substates that this state has.
	Children    []*State
	Transitions []*Transition

	EnterReactions []*StateReaction
	ExitReactions  []*StateReaction

	Parent       *State
	frontendData *frontend.StateData
}

func (s *State) SafeName() string {
	if s == nil {
		return InvalidStateName
	}

	return s.Name
}

func (s *State) Equals(other *State) bool {
	return s.Name == other.Name
}

func (s *State) DefaultChild() *State {
	if len(s.Children) == 0 {
		return nil
	}

	for _, child := range s.Children {
		if child.Default {
			return child
		}
	}

	panic("No default child found. This should've been caught in validation")
}

func (s *State) IsParentOf(other *State) bool {
	for _, child := range s.Children {
		if child.Equals(other) {
			return true
		}
	}

	return false
}

func (s *State) HasTransitionFor(trigger *Trigger) bool {
	return s.GetTransitionFor(trigger) != nil
}

func (s *State) GetTransitionFor(trigger *Trigger) *Transition {
	for _, transition := range s.Transitions {
		if transition.Trigger == trigger {
			return transition
		}
	}

	return nil
}

func (s *State) HasDefaultReaction(reactionType string) bool {
	if reactionType == "Enter" {
		return s.HasDefaultEnterReaction()
	}

	if reactionType == "Exit" {
		return s.HasDefaultExitReaction()
	}

	log.Fatalf("Unknown reaction %q", reactionType)
	panic(false)
}

func (s *State) HasDefaultEnterReaction() bool {
	return s.HasEnterReaction(nil)
}

func (s *State) HasDefaultExitReaction() bool {
	return s.HasExitReaction(nil)
}

func (s *State) HasReaction(reactionType string, trigger *Trigger) bool {
	if reactionType == "Enter" {
		return s.HasEnterReaction(trigger)
	}

	if reactionType == "Exit" {
		return s.HasExitReaction(trigger)
	}

	log.Fatalf("Unknown reaction %q", reactionType)
	panic(false)
}

func (s *State) HasEnterReaction(trigger *Trigger) bool {
	for _, reaction := range s.EnterReactions {
		if reaction.Trigger == trigger {
			return true
		}
	}

	return false
}

func (s *State) HasExitReaction(trigger *Trigger) bool {
	for _, reaction := range s.ExitReactions {
		if reaction.Trigger == trigger {
			return true
		}
	}

	return false
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
