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

	Frontend_Index      int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_NamePos    *files.LoadedFilePosition
}

var _ FrontendEntry = (*StatechartData)(nil)

func (sd *StatechartData) Index() int                          { return sd.Frontend_Index }
func (sd *StatechartData) Position() *files.LoadedFilePosition { return sd.Frontend_KeywordPos }

// TriggerData is all the information the frontend has about a trigger definition.
type TriggerData struct {
	Name       string
	ArgsString string

	Frontend_Index      int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_NamePos    *files.LoadedFilePosition
	Frontend_ArgsPos    *files.LoadedFilePosition
}

var _ FrontendEntry = (*StatechartData)(nil)

func (td *TriggerData) Index() int                          { return td.Frontend_Index }
func (td *TriggerData) Position() *files.LoadedFilePosition { return td.Frontend_KeywordPos }

// StateData is all the information the frontend has about a state definition.
type StateData struct {
	Name   string
	Parent string

	Options []*OptionData

	EnterReactions []*ReactionData
	ExitReactions  []*ReactionData

	Captures []*CaptureData

	Frontend_Index      int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_NamePos    *files.LoadedFilePosition
}

var _ FrontendEntry = (*StatechartData)(nil)

func (sd *StateData) Index() int                          { return sd.Frontend_Index }
func (sd *StateData) Position() *files.LoadedFilePosition { return sd.Frontend_KeywordPos }

// ReactionData represents entries for "reactions" to triggers.
// This is mostly enter/exit statements.
type ReactionData struct {
	Trigger string

	Frontend_Index      int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_TriggerPos *files.LoadedFilePosition
}

var _ FrontendEntry = (*StatechartData)(nil)

func (rd *ReactionData) Index() int                          { return rd.Frontend_Index }
func (rd *ReactionData) Position() *files.LoadedFilePosition { return rd.Frontend_KeywordPos }

// CaptureData represents captures, which "capture" a trigger and do things with them.
type CaptureData struct {
	Trigger   string
	Condition string

	Frontend_Index      int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_TriggerPos *files.LoadedFilePosition
}

var _ FrontendEntry = (*StatechartData)(nil)

func (cd *CaptureData) Index() int                          { return cd.Frontend_Index }
func (cd *CaptureData) Position() *files.LoadedFilePosition { return cd.Frontend_KeywordPos }

// TransitionData represents a transition, which, like captures, hold a trigger and perform a
// transition to another state within the gochart.
type TransitionData struct {
	From      string
	To        string
	Trigger   string
	Condition string
	Action    string

	Frontend_Index      int
	Frontend_KeywordPos *files.LoadedFilePosition
	Frontend_TriggerPos *files.LoadedFilePosition
	Frontend_TargetPos  *files.LoadedFilePosition
}

var _ FrontendEntry = (*StatechartData)(nil)

func (td *TransitionData) Index() int                          { return td.Frontend_Index }
func (td *TransitionData) Position() *files.LoadedFilePosition { return td.Frontend_KeywordPos }

func (tdata *TransitionData) String() string {
	return fmt.Sprintf("transition %s: %s > %s", tdata.Trigger, tdata.From, tdata.To)
}
