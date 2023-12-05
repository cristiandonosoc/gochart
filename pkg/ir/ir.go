// Package ir holds the intermediate representation of the Statechart, which will then permit us to
// generate representations from it (eg. to C++).
package ir

import (
	"fmt"
	"log"

	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

const (
	InvalidTriggerName = "Invalid"
	InvalidStateName   = "Invalid"
)

type ConcurrentHead struct {
	Id        int
	HeadState *State
}

// Statechart represents a single statechart state machine.
type Statechart struct {
	Name string

	Triggers []*Trigger

	// Roots are the list of states that are "roots" (have no parents) within the state tree.
	Roots           []*State
	ConcurrentHeads []*ConcurrentHead

	// States are all the states, as defined in the order from the frontend.
	States []*State

	TriggerMap   map[string]*Trigger
	StateMap     map[string]*State
	frontendData *output.StatechartData
}

func (sc *Statechart) ClassName() string {
	return sc.Name
}

func (sc *Statechart) DefaultRootState() *State {
	for _, state := range sc.States {
		if state.Default {
			return state
		}
	}

	panic("No default root state found. This should've been caught in validation")
}

func (sc *Statechart) ConcurrentGroupCount() int {
	return len(sc.ConcurrentHeads)
}

// TRIGGER -----------------------------------------------------------------------------------------

type Trigger struct {
	Name string
	Args []*TriggerArgument

	frontendData *output.TriggerData
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

	Concurrent      bool
	ConcurrentGroup int

	Options []*Option

	// States represents the substates that this state has.
	Children    []*State
	Transitions []*Transition

	EnterReactions []*StateReaction
	ExitReactions  []*StateReaction

	Captures []*StateCapture

	Parent *State

	statechart   *Statechart
	frontendData *output.StateData
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

func (s *State) GetConcurrentGroup() int {
	// Create a set of states.
	headStates := make(map[*State]int)
	for group, head := range s.statechart.ConcurrentHeads {
		headStates[head.HeadState] = group
	}

	// Now we search until we find a group.
	current := s
	for current != nil {
		// See if this is a concurrent head.
		if group, ok := headStates[current]; ok {
			return group
		}

		current = current.Parent
	}

	// Arriving here means that this state wasn't child of any concurrent head.
	return 0
}

func (s *State) HasTransition(trigger *Trigger) bool {
	return s.GetTransition(trigger) != nil
}

func (s *State) GetTransition(trigger *Trigger) *Transition {
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

	if reactionType == "Capture" {
		return false
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

	if reactionType == "Capture" {
		return s.HasCapture(trigger)
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

func (s *State) HasCapture(trigger *Trigger) bool {
	return s.GetCapture(trigger) != nil
}

func (s *State) GetCapture(trigger *Trigger) *StateCapture {
	for _, capture := range s.Captures {
		if capture.Trigger == trigger {
			return capture
		}
	}

	return nil
}

// STATE OPTIONS -----------------------------------------------------------------------------------

// STATE REACTION ----------------------------------------------------------------------------------

type StateReactionKind int

const (
	StateReactionKind_Enter StateReactionKind = iota
	StateReactionKind_Exit
)

func (srk StateReactionKind) String() string {
	switch srk {
	case StateReactionKind_Enter:
		return "Enter"
	case StateReactionKind_Exit:
		return "Exit"
	}

	panic(fmt.Sprintf("unknown reaction kind %d", srk))
}

type StateReaction struct {
	Trigger *Trigger
	Kind    StateReactionKind

	frontendData *output.ReactionData
}

func (sr *StateReaction) String() string {
	kind := sr.Kind.String()
	if sr.Trigger == nil {
		return fmt.Sprintf("Default%s", kind)
	}

	return fmt.Sprintf("%sOn%s", kind, sr.Trigger.Name)
}

// STATE CAPTURE -----------------------------------------------------------------------------------

type StateCapture struct {
	Trigger   *Trigger
	Condition string

	frontendData *output.CaptureData
}

func (sc *StateCapture) String() string {
	return fmt.Sprintf("capture(%s)", sc.Trigger.SafeName())
}

func (sc *StateCapture) HasCondition() bool {
	return sc.Condition != ""
}

// TRANSITION --------------------------------------------------------------------------------------

// Transition represents a transition from one state to another given a particular trigger.
type Transition struct {
	From      *State
	To        *State
	Trigger   *Trigger
	Condition string
	Action    string

	frontendData *output.TransitionData
}

func (t *Transition) IsNullTransition() bool {
	return t.Trigger == nil
}

func (t *Transition) HasCondition() bool {
	return t.Condition != ""
}

func (t *Transition) HasAction() bool {
	return t.Action != ""
}
