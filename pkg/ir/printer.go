package ir

import (
	"io"

	"github.com/cristiandonosoc/gochart/internal/printer"
)

type irPrinter struct {
	sc *Gochart
	p  *printer.Printer
}

func NewPrinter(w io.Writer) *irPrinter {
	p := printer.NewPrinter(w)

	return &irPrinter{
		p: p,
	}
}

func (ip *irPrinter) PrintGochart(sc *Gochart) {
	ip.sc = sc

	ip.p.WriteLine("Gochart %s", sc.Name)

	ip.p.PushIndent()
	defer ip.p.PopIndent()

	for _, trigger := range sc.Triggers {
		ip.printTrigger(trigger)
	}

	ip.printState(sc.Root)
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

	if len(state.Transitions) > 0 {
		// Print all the transitions for this state.
		ip.p.WriteLine("Transitions:")
		ip.p.PushIndent()
		ip.printTransitions(state.Transitions)
		ip.p.PopIndent()
	}

	if len(state.EnterReactions) > 0 {
		ip.p.WriteLine("Enter Reactions:")
		ip.p.PushIndent()
		ip.printReactions("Enter", state.EnterReactions)
		ip.p.PopIndent()
	}

	if len(state.ExitReactions) > 0 {
		ip.p.WriteLine("Exit Reactions:")
		ip.p.PushIndent()
		ip.printReactions("Exit", state.ExitReactions)
		ip.p.PopIndent()
	}

	if len(state.Captures) > 0 {
		ip.p.WriteLine("Captures:")
		ip.p.PushIndent()
		ip.printCaptures(state.Captures)
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
	ip.p.WriteLine("Concurrent: %t, ConcurrentGroup: %d", state.Concurrent, state.ConcurrentGroup)
}

func (ip *irPrinter) printTransitions(transitions []*Transition) {
	for _, transition := range transitions {
		ip.p.WriteLine("Transition %s > %s", transition.Trigger.SafeName(), transition.To.Name)
	}
}

func (ip *irPrinter) printReactions(kind string, reactions []*StateReaction) {
	for _, reaction := range reactions {
		ip.p.WriteLine("%s(%s)", kind, reaction.Trigger.SafeName())
	}
}

func (ip *irPrinter) printCaptures(captures []*StateCapture) {
	for _, capture := range captures {
		ip.p.WriteLine("Trigger: %s", capture.Trigger.SafeName())
	}
}
