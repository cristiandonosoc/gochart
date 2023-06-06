// Package ir holds the intermediate representation of the Statechart, which will then permit us to
// generate representations from it (eg. to C++).
package ir

type Statechart struct {
	States      []*State
	Transitions []*Transition
}

type State struct {
}

type Transition struct {
	From    *State
	To      *State
	Trigger *Trigger
}

type Trigger struct {
	Name string
}
