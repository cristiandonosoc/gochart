package ir

import (
	"io"

	"github.com/cristiandonosoc/gochart/internal/printer"
)

type irPrinter struct {
	sc *Statechart
	p  *printer.Printer
}

func NewPrinter(w io.Writer) *irPrinter {
	p := printer.NewPrinter(w)

	return &irPrinter{
		p: p,
	}
}

func (ip *irPrinter) PrintStatechart(sc *Statechart) {
	ip.sc = sc

	ip.p.WriteLine("Statechart %s", sc.Name)

	ip.p.PushIndent()
	defer ip.p.PopIndent()

	for _, trigger := range sc.Triggers {
		ip.printTrigger(trigger)
	}

	for _, root := range sc.Roots {
		ip.printState(root)
	}
}

func (ip *irPrinter) printTrigger(trigger *Trigger) {
	ip.p.WriteLine("Trigger %s(%s)", trigger.Name, trigger.ArgsStringList())
}

func (ip *irPrinter) printState(state *State) {
	ip.p.WriteLine("State %s", state.Name)

	ip.p.PushIndent()
	defer ip.p.PopIndent()

	ip.p.WriteLine("Parent %q", state.Parent.SafeName())

	{
		ip.p.WriteLine("Options:")
		ip.p.PushIndent()
		ip.printStateOptions(state)
		ip.p.PopIndent()
	}

	{
		// Print all the transitions for this state.
		ip.p.WriteLine("Transitions:")
		ip.p.PushIndent()
		ip.printTransitions(state.Transitions)
		ip.p.PopIndent()
	}

	{
		ip.p.WriteLine("Enter Reactions:")
		ip.p.PushIndent()
		ip.printReactions("Enter", state.EnterReactions)
		ip.p.PopIndent()
	}

	{
		ip.p.WriteLine("Exit Reactions:")
		ip.p.PushIndent()
		ip.printReactions("Exit", state.ExitReactions)
		ip.p.PopIndent()
	}

	// Children.
	for _, child := range state.Children {
		ip.p.WriteLine("*** %s CHILD STATE ***", state.Name)
		ip.printState(child)
	}
}

func (ip *irPrinter) printStateOptions(state *State) {
	ip.p.WriteLine("Default: %v", state.Default)
}

func (ip *irPrinter) printTransitions(transitions []*Transition) {
	if len(transitions) == 0 {
		ip.p.WriteLine("No transitions")
		return
	}

	for _, transition := range transitions {
		ip.p.WriteLine("Transition %s > %s", transition.Trigger.SafeName(), transition.To.Name)
	}
}

func (ip *irPrinter) printReactions(kind string, reactions []*StateReaction) {
	if len(reactions) == 0 {
		ip.p.WriteLine("No %s reactions", kind)
		return
	}

	for _, reaction := range reactions {
		ip.p.WriteLine("%s(%s)", kind, reaction.Trigger.SafeName())
	}
}
