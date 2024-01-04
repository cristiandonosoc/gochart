package frontend

import (
	"io"

	"github.com/cristiandonosoc/gochart/internal/printer"
	"github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

type frontendPrinter struct {
	gc *common.GochartData
	p  *printer.Printer
}

func NewPrinter(w io.Writer) *frontendPrinter {
	p := printer.NewPrinter(w)

	return &frontendPrinter{
		p: p,
	}
}

func (fp *frontendPrinter) PrintGochart(gc *common.GochartData) {
	fp.gc = gc

	fp.p.WriteLine("Gochart %s", gc.Name)

	fp.p.PushIndent()
	defer fp.p.PopIndent()

	for _, trigger := range gc.Triggers {
		fp.printTrigger(trigger)
	}

	fp.printState(gc.Root)
}

func (fp *frontendPrinter) printTrigger(trigger *common.TriggerData) {
	fp.p.WriteLine("Trigger %s(%s)", trigger.Name, trigger.ArgsString)
}

func (fp *frontendPrinter) printState(state *common.StateData) {
	fp.p.WriteLine("State %s", state.Name)

	fp.p.PushIndent()
	defer fp.p.PopIndent()

	fp.p.WriteLine("Parent %q", state.Parent)

	{
		fp.p.WriteLine("Options:")
		fp.p.PushIndent()
		fp.printOptions(state.Options)
		fp.p.PopIndent()
	}

	{
		// Print all the transitions for this state.
		var transitions []*common.TransitionData
		for _, transition := range fp.gc.Transitions {
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

func (fp *frontendPrinter) printOptions(options []*common.OptionData) {
	if len(options) == 0 {
		fp.p.WriteLine("No options")
		return
	}

	for _, option := range options {
		fp.p.WriteLine("Option %s (Kind: %s, Value: %q)", option.Name, option.Kind, option.Value)
	}
}

func (fp *frontendPrinter) printTransitions(transitions []*common.TransitionData) {
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

func (fp *frontendPrinter) printReactions(kind string, reactions []*common.ReactionData) {
	if len(reactions) == 0 {
		fp.p.WriteLine("No %s reactions", kind)
		return
	}

	for _, reaction := range reactions {
		if reaction.Trigger == "" {
			fp.p.WriteLine("%s", kind)
			continue
		}

		fp.p.WriteLine("%s(%s)", kind, reaction.Trigger)
	}
}
