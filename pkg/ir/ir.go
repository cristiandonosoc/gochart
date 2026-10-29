// Package ir holds the intermediate representation of the Gochart, which will then permit us to
// generate representations from it (eg. to C++).
package ir

import (
	"fmt"
	"log"

	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

const (
	InvalidTriggerName = "__Invalid__"
	InvalidStateName   = "__Invalid__"
)

type ParallelHead struct {
	Id        int
	HeadState *State
}

// Language is the language a gochart is written for. It determines how trigger arguments are parsed
// and which backend can generate it.
type Language string

const (
	Language_Cpp  Language = "cpp"
	Language_Rust Language = "rust"
)

// Gochart represents a single gochart state machine.
type Gochart struct {
	Name      string
	Namespace string

	Language Language
	// Derives are extra derives for the generated types (Rust only), as written in the option
	// (eg. "serde::Serialize, serde::Deserialize").
	Derives []string

	Options []*Option

	Triggers []*Trigger

	// Root is the base state of the complete gochart.
	// Normally this is a "hidden" state created by the gochart.
	Root          *State
	ParallelHeads []*ParallelHead

	// States are all the states, as defined in the order from the frontend.
	States []*State

	TriggerMap   map[string]*Trigger
	StateMap     map[string]*State
	FrontendData *frontend_common.GochartData
}

func (gc *Gochart) ParallelGroupCount() int {
	return len(gc.ParallelHeads)
}

// TRIGGER -----------------------------------------------------------------------------------------

type Trigger struct {
	Name string
	Args []*TriggerArgument

	FrontendData *frontend_common.TriggerData
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

// RustString returns the argument as written in a Rust signature.
func (ta *TriggerArgument) RustString() string {
	return fmt.Sprintf("%s: %s", ta.Name, ta.Type)
}

// STATE -------------------------------------------------------------------------------------------

// State represents a single state withing a gochart.
type State struct {
	Name    string
	Default bool

	Parallel      bool
	ParallelGroup int

	// DataType is the type of the data that lives while this state is active (Rust only).
	DataType string
	// DataDefault is whether the data is created with |Default::default()| rather than a callback.
	DataDefault bool

	Options []*Option

	// States represents the substates that this state has.
	Children    []*State
	Transitions []*Transition

	EnterReactions []*StateReaction
	ExitReactions  []*StateReaction

	Captures []*StateCapture

	Parent *State

	gochart      *Gochart
	FrontendData *frontend_common.StateData
}

func (s *State) IsRoot() bool {
	if s.Parent != nil {
		return false
	}
	return true
}

// DisplayName is the name to be used for user facing displays of this state.
// Mostly used to give an easier time to the backend.
func (s *State) DisplayName() string {
	if s.IsRoot() {
		return "Gochart"
	}
	return fmt.Sprintf("State%s", s.Name)
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

func (s *State) GetParallelGroup() int {
	// Create a set of states.
	headStates := make(map[*State]int)
	for group, head := range s.gochart.ParallelHeads {
		headStates[head.HeadState] = group
	}

	// Now we search until we find a group.
	current := s
	for current != nil {
		// See if this is a parallel head.
		if group, ok := headStates[current]; ok {
			return group
		}

		current = current.Parent
	}

	// Arriving here means that this state wasn't child of any parallel head.
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

func (s *State) GetNullTransitions() []*Transition {
	var result []*Transition
	for _, transition := range s.Transitions {
		if !transition.IsNullTransition() {
			continue
		}

		result = append(result, transition)
	}

	return result
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

	FrontendData *frontend_common.ReactionData
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
	// ConditionCallback is the name of the condition callback ("if some_condition").
	ConditionCallback string

	FrontendData *frontend_common.CaptureData
}

func (sc *StateCapture) String() string {
	return fmt.Sprintf("capture(%s)", sc.Trigger.SafeName())
}

func (sc *StateCapture) HasCondition() bool {
	return sc.Condition != ""
}

// IsConditional is whether the capture has a condition, either inlined or as a named callback.
func (sc *StateCapture) IsConditional() bool {
	return sc.Condition != "" || sc.ConditionCallback != ""
}

// TRANSITION --------------------------------------------------------------------------------------

// Transition represents a transition from one state to another given a particular trigger.
type Transition struct {
	From      *State
	To        *State
	Trigger   *Trigger
	Condition string
	Action    string

	// Named callbacks ("if some_condition", "do some_action").
	ConditionCallback string
	ActionCallback    string

	FrontendData *frontend_common.TransitionData
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

// IsConditional is whether the transition has a condition, either inlined or as a named callback.
func (t *Transition) IsConditional() bool {
	return t.Condition != "" || t.ConditionCallback != ""
}
