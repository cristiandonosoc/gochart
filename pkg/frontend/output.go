package frontend

import (
	"fmt"
)

// StatechartData is all the information a frontend needs to output about a statechart.
// This will be consumed by the |ir| package and validated.
type StatechartData struct {
	Name        string            `yaml:"name"`
	Triggers    []*TriggerData    `yaml:"triggers"`
	States      []*StateData      `yaml:"states"`
	Transitions []*TransitionData `yaml:"transitions"`
}

type TriggerData struct {
	Name       string `yaml:"name"`
	ArgsString string `yaml:"arguments_string"`

	// Index represents in what order it was found.
	Index int
}

type StateData struct {
	Name   string `yaml:"name"`
	Parent string `yaml:"parent"`

	Options []*StateOptionData

	EnterReactionTriggers []string `yaml:"enter_reaction_triggers"`
	ExitReactionTriggers  []string `yaml:"exit_reaction_triggers"`

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
	From    string `yaml:"from"`
	To      string `yaml:"to"`
	Trigger string `yaml:"trigger"`

	// Index represents in what order it was found.
	Index int
}

func (tdata *TransitionData) String() string {
	return fmt.Sprintf("transition %s: %s > %s", tdata.Trigger, tdata.From, tdata.To)
}
