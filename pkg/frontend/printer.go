package frontend

import (
	"io"

	"github.com/cristiandonosoc/gochart/internal/printer"
)

const (
	indentAmount = 4
)

// Printer is a helper struct meant to pretty print the frontend data.
type frontendPrinter struct {
	sc *StatechartData
	p  *printer.Printer
}

func NewPrinter(w io.Writer) *frontendPrinter {
	p := printer.NewPrinter(w)

	return &frontendPrinter{
		p: p,
	}
}

func (fp *frontendPrinter) PrintStatechart(sc *StatechartData) {
	fp.sc = sc

	fp.p.WriteLine("Statechart %s", sc.Name)

	fp.p.PushIndent()
	defer fp.p.PopIndent()

	for _, trigger := range sc.Triggers {
		fp.printTrigger(trigger)
	}

	for _, state := range sc.States {
		fp.printState(state)
	}
}

func (fp *frontendPrinter) printTrigger(trigger *TriggerData) {
	fp.p.WriteLine("Trigger %s(%s)", trigger.Name, trigger.ArgsString)
}

func (fp *frontendPrinter) printState(state *StateData) {
	fp.p.WriteLine("State %s", state.Name)

	fp.p.PushIndent()
	defer fp.p.PopIndent()

	fp.p.WriteLine("Parent %q", state.Parent)

	{
		fp.p.WriteLine("Options:")
		fp.p.PushIndent()
		fp.printStateOptions(state.Options)
		fp.p.PopIndent()
	}

	{
		// Print all the transitions for this state.
		var transitions []*TransitionData
		for _, transition := range fp.sc.Transitions {
			if transition.From == state.Name {
				transitions = append(transitions, transition)
			}
		}
		fp.p.WriteLine("Transitions:")
		fp.p.PushIndent()
		fp.printTransitions(transitions)
		fp.p.PopIndent()
	}

	{
		fp.p.WriteLine("Enter Reactions:")
		fp.p.PushIndent()
		fp.printReactions("Enter", state.EnterReactions)
		fp.p.PopIndent()
	}

	{
		fp.p.WriteLine("Exit Reactions:")
		fp.p.PushIndent()
		fp.printReactions("Exit", state.ExitReactions)
		fp.p.PopIndent()
	}

}

func (fp *frontendPrinter) printStateOptions(options []*StateOptionData) {
	if len(options) == 0 {
		fp.p.WriteLine("No options")
		return
	}

	for _, option := range options {
		fp.p.WriteLine("Option %s (Kind: %s, Value: %q)", option.Name, option.ValueKind, option.Value)
	}
}

func (fp *frontendPrinter) printTransitions(transitions []*TransitionData) {
	if len(transitions) == 0 {
		fp.p.WriteLine("No transitions")
		return
	}

	for _, transition := range transitions {
		trigger := "null"
		if transition.Trigger != "" {
			trigger = transition.Trigger
		}

		fp.p.WriteLine("Transition %s > %s", trigger, transition.To)
	}
}

func (fp *frontendPrinter) printReactions(kind string, reactions []string) {
	if len(reactions) == 0 {
		fp.p.WriteLine("No %s reactions", kind)
		return
	}

	for _, reaction := range reactions {
		if reaction == "" {
			fp.p.WriteLine("%s", kind)
			continue
		}

		fp.p.WriteLine("%s(%s)", kind, reaction)
	}
}
