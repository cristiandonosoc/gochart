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

	Options []*OptionData

	EnterReactions []string
	ExitReactions  []string

	Captures []*CaptureData

	// Index represents in what order it was found.
	Index int
}

type CaptureData struct {
	Trigger   string
	Condition string
}

type TransitionData struct {
	From      string
	To        string
	Trigger   string
	Condition string
	Action    string

	// Index represents in what order it was found.
	Index int
}

func (tdata *TransitionData) String() string {
	return fmt.Sprintf("transition %s: %s > %s", tdata.Trigger, tdata.From, tdata.To)
}
