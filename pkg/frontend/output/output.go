// Package output represents the output of a frontend, normally used by the IR.
package output

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/files"
)

// StatechartData is all the information a frontend needs to output about a statechart.
// This will be consumed by the |ir| package and validated.
type StatechartData struct {
	Name        string
	Triggers    []*TriggerData
	States      []*StateData
	Transitions []*TransitionData

	FrontendPosition *files.LoadedFilePosition
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

	FrontendPosition *files.LoadedFilePosition
}

type TriggerData struct {
	Name       string
	ArgsString string

	FrontendPosition *files.LoadedFilePosition
}

type StateData struct {
	Name   string
	Parent string

	Options []*OptionData

	// TODO(cdc): Reactions should be its own struct, so we can append position data.
	EnterReactions []string
	ExitReactions  []string

	Captures []*CaptureData

	FrontendPosition *files.LoadedFilePosition
}

type CaptureData struct {
	Trigger   string
	Condition string

	FrontendPosition *files.LoadedFilePosition
}

type TransitionData struct {
	From      string
	To        string
	Trigger   string
	Condition string
	Action    string

	FrontendPosition *files.LoadedFilePosition
}

func (tdata *TransitionData) String() string {
	return fmt.Sprintf("transition %s: %s > %s", tdata.Trigger, tdata.From, tdata.To)
}
