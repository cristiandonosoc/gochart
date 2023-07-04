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
	Name            string `yaml:"name"`
	ArgumentsString string `yaml:"arguments_string"`

	// Index represents in what order it was found.
	Index int
}

type StateData struct {
	Name    string `yaml:"name"`
	Initial bool   `yaml:"initial"`
	Parent  string `yaml:"parent"`

	DefaultEnter          bool     `yaml:"default_enter"`
	EnterReactionTriggers []string `yaml:"enter_reaction_triggers"`

	DefaultExit          bool     `yaml:"default_exit"`
	ExitReactionTriggers []string `yaml:"exit_reaction_triggers"`

	// Index represents in what order it was found.
	Index int
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
