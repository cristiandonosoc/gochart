package frontend

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
	Name            string
	ArgumentsString string

	// Index represents in what order it was found.
	Index int
}

type StateData struct {
	Name   string
	Parent string

	DefaultEnter bool
	EnterReactionTriggers []string

	DefaultExit bool
	ExitReactionTriggers []string

	// Index represents in what order it was found.
	Index int
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
