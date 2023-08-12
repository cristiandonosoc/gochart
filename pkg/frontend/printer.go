package frontend

import (
	"fmt"
	"io"
)

const (
	indentAmount = 4
)

// Printer is a helper struct meant to pretty print the frontend data.
type printer struct {
	indent int
	w      io.Writer

	sc *StatechartData
}

func NewPrinter(w io.Writer) *printer {
	return &printer{
		w: w,
	}
}

func (p *printer) PrintStatechart(sc *StatechartData) {
	p.sc = sc

	p.writeLine("Statechart %s", sc.Name)

	p.pushIndent()
	defer p.popIndent()

	for _, trigger := range sc.Triggers {
		p.printTrigger(trigger)
	}

	for _, state := range sc.States {
		p.printState(state)
	}
}

func (p *printer) printTrigger(trigger *TriggerData) {
	p.writeLine("Trigger %s(%s)", trigger.Name, trigger.ArgsString)
}

func (p *printer) printState(state *StateData) {
	p.writeLine("State %s", state.Name)

	p.pushIndent()
	defer p.popIndent()

	p.writeLine("Parent %q", state.Parent)

	{
		p.writeLine("Options:")
		p.pushIndent()
		p.printStateOptions(state.Options)
		p.popIndent()
	}

	{
		// Print all the transitions for this state.
		var transitions []*TransitionData
		for _, transition := range p.sc.Transitions {
			if transition.From == state.Name {
				transitions = append(transitions, transition)
			}
		}
		p.writeLine("Transitions:")
		p.pushIndent()
		p.printTransitions(transitions)
		p.popIndent()
	}

	{
		p.writeLine("Enter Reactions:")
		p.pushIndent()
		p.printReactions("Enter", state.EnterReactions)
		p.popIndent()
	}

	{
		p.writeLine("Exit Reactions:")
		p.pushIndent()
		p.printReactions("Exit", state.ExitReactions)
		p.popIndent()
	}

}

func (p *printer) printStateOptions(options []*StateOptionData) {
	if len(options) == 0 {
		p.writeLine("No options")
		return
	}

	for _, option := range options {
		p.writeLine("Option %s (Kind: %s, Value: %q)", option.Name, option.ValueKind, option.Value)
	}
}

func (p *printer) printTransitions(transitions []*TransitionData) {
	if len(transitions) == 0 {
		p.writeLine("No transitions")
		return
	}

	for _, transition := range transitions {
		trigger := "null"
		if transition.Trigger != "" {
			trigger = transition.Trigger
		}

		p.writeLine("Transition %s > %s", trigger, transition.To)
	}
}

func (p *printer) printReactions(kind string, reactions []string) {
	if len(reactions) == 0 {
		p.writeLine("No %s reactions", kind)
		return
	}

	for _, reaction := range reactions {
		if reaction == "" {
			p.writeLine("%s", kind)
			continue
		}

		p.writeLine("%s(%s)", kind, reaction)
	}
}

func (p *printer) writeLine(format string, args ...interface{}) {
	result := fmt.Sprintf(format, args...)
	final := fmt.Sprintf("%*s%s\n", p.indent, " ", result)
	p.w.Write([]byte(final))
}

func (p *printer) pushIndent() {
	p.indent += indentAmount
}

func (p *printer) popIndent() {
	p.indent -= indentAmount
	if p.indent < 0 {
		panic("Unmatched popIndent")
	}
}
