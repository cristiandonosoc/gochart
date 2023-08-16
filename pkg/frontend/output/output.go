// Package output represents the output of a frontend, normally used by the IR.
package output

import (
	"fmt"
)

// StatechartData is all the information a frontend needs to output about a statechart.
// This will be consumed by the |ir| package and validated.
type StatechartData struct {
	Name        string
	Triggers    []*TriggerData
	States      []*StateData
	Transitions []*TransitionData
}

type TriggerData struct {
	Name       string
	ArgsString string

	// Index represents in what order it was found.
	Index int
}

type StateData struct {
	Name   string
	Parent string

	Options []*StateOptionData

	EnterReactions []string
	ExitReactions  []string

	// Index represents in what order it was found.
	Index int
}

//go:generate stringer -type=StateOptionValueKind
type StateOptionValueKind int

const (
	StateOptionValueKind_None StateOptionValueKind = iota
	StateOptionValueKind_String
	StateOptionValueKind_Integer
	StateOptionValueKind_Boolean
)

type StateOptionData struct {
	Name      string
	ValueKind StateOptionValueKind
	Value     string
}

type TransitionData struct {
	From    string
	To      string
	Trigger string

	// Index represents in what order it was found.
	Index int
}

func (tdata *TransitionData) String() string {
	return fmt.Sprintf("transition %s: %s > %s", tdata.Trigger, tdata.From, tdata.To)
}
