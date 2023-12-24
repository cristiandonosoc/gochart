// Package common represents common types that is to be used by all frontend packages.
package common

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/files"
)

const (
	RootStateName = "__Root__"
)

type OptionValueKind int

const (
	OptionValueKind_Boolean OptionValueKind = iota
	OptionValueKind_String
	OptionValueKind_Integer
)

func (ovk OptionValueKind) String() string {
	switch ovk {
	case OptionValueKind_Boolean:
		return "boolean"
	case OptionValueKind_String:
		return "string"
	case OptionValueKind_Integer:
		return "int"
	default:
		panic(fmt.Sprintf("unknown OptionValueKind %v", int(ovk)))
	}
}

type OptionData struct {
	Name  string
	Kind  OptionValueKind
	Value string

	Frontend_NamePos *files.LoadedFilePosition
}

// StatechartData is all the information a frontend needs to output about a statechart.
// This will be consumed by the |ir| package and validated.
type StatechartData struct {
	Name string

	Options []*OptionData

	Triggers    []*TriggerData
	Root        *StateData
	States      []*StateData
	Transitions []*TransitionData

	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_NamePos    *files.LoadedFilePosition
}

type TriggerData struct {
	Name       string
	ArgsString string

	Frontend_EntryIndex int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_NamePos    *files.LoadedFilePosition
	Frontend_ArgsPos    *files.LoadedFilePosition
}

type StateData struct {
	Name   string
	Parent string

	Options []*OptionData

	EnterReactions []*ReactionData
	ExitReactions  []*ReactionData

	Captures []*CaptureData

	Frontend_EntryIndex int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_NamePos    *files.LoadedFilePosition
}

// ReactionData represents entries for "reactions" to triggers.
// This is mostly enter/exit statements.
type ReactionData struct {
	Trigger string

	Frontend_EntryIndex int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_TriggerPos *files.LoadedFilePosition
}

// CaptureData represents captures, which "capture" a trigger and do things with them.
type CaptureData struct {
	Trigger   string
	Condition string

	Frontend_EntryIndex int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_TriggerPos *files.LoadedFilePosition
}

// TransitionData represents a transition, which, like captures, hold a trigger and perform a
// transition to another state within the gochart.
type TransitionData struct {
	From      string
	To        string
	Trigger   string
	Condition string
	Action    string

	Frontend_EntryIndex int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_TriggerPos *files.LoadedFilePosition
	Frontend_TargetPos  *files.LoadedFilePosition
}

func (tdata *TransitionData) String() string {
	return fmt.Sprintf("transition %s: %s > %s", tdata.Trigger, tdata.From, tdata.To)
}
