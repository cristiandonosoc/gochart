package frontend

import ()

// StatechartData is all the information a frontend needs to output about a statechart.
// This will be consumed by the |ir| package and validated.
type StatechartData struct {
	Name        string
	States      []*StateData
	Transitions []*TransitionData
	Triggers    []*TriggerData
}

type StateData struct {
	Name   string
	Parent string
}

type TriggerData struct {
	Name      string
	Arguments []string
}

type TransitionData struct {
	From    string
	To      string
	Trigger string
}
