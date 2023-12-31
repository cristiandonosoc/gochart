package formatter

import (
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/cristiandonosoc/gochart/internal/files"
	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

type formatterEntry interface {
	FrontendEntry() frontend_common.FrontendEntry
	Name() string
	Indents() bool
	ChildEntries() []formatterEntry

	Write(w io.Writer)
	DebugWrite(w io.Writer)
}

func newFormatterEntry(e any) formatterEntry {
	switch v := e.(type) {
	case *ir.Statechart:
		return newFormatStatechartWrapper(v)
	case *ir.Trigger:
		return newFormatTrigger(v)
	case *ir.State:
		return newFormatState(v)
	case *ir.StateReaction:
		return newFormatReaction(v)
	case *ir.StateCapture:
		return newFormatCapture(v)
	case *ir.Transition:
		return newFormatTransition(v)
	default:
		panic(fmt.Sprintf("no formatter for type %q", reflect.TypeOf(v).Name()))
	}
}

// StatechartWrapper & Statechart ------------------------------------------------------------------
//
// Statechart has a wrapper because there can be comments outside the statechart and we don't have a
// "higher-level" entity above the statechart that contains both the statechart and the commnets.
// So in this case the statechart contains both the comments and external comments.

type formatStatechartWrapper struct {
	sc      *ir.Statechart
	entries []formatterEntry
}

func newFormatStatechartWrapper(sc *ir.Statechart) *formatStatechartWrapper {
	f := &formatStatechartWrapper{
		sc: sc,
	}

	// Create a child with the statechart as a child.
	// TODO(cdc): Add comments into this wrapper.
	entries := []formatterEntry{
		newFormatStatechart(sc),
	}

	f.entries = entries
	return f
}

// The frontend_common.FrontendEntry interface
func (f *formatStatechartWrapper) Index() int { return 0 }
func (f *formatStatechartWrapper) Position() *files.LoadedFilePosition {
	scPos := f.sc.FrontendData.Position()

	// The wrapper is positioned at the first character of the first line.
	return &files.LoadedFilePosition{
		File: scPos.File,
		Line: 1,
		Char: 1,
	}
}

func (f *formatStatechartWrapper) FrontendEntry() frontend_common.FrontendEntry { return f }
func (f *formatStatechartWrapper) Name() string                                 { return "<statechart_wrapper>" }
func (f *formatStatechartWrapper) Indents() bool                                { return false }
func (f *formatStatechartWrapper) ChildEntries() []formatterEntry               { return f.entries }

func (f *formatStatechartWrapper) Write(w io.Writer) {
	// TODO(cdc): Print out of statechart comments.
}
func (f *formatStatechartWrapper) DebugWrite(w io.Writer) {
	// TODO(cdc): Print out of statechart comments.
}

type formatStatechart struct {
	sc      *ir.Statechart
	entries []formatterEntry
}

func newFormatStatechart(sc *ir.Statechart) *formatStatechart {
	f := &formatStatechart{
		sc: sc,
	}

	var entries []formatterEntry
	for _, trigger := range sc.Triggers {
		entries = append(entries, newFormatterEntry(trigger))
	}

	// Go over the root state ang gets the entries from it.
	entries = append(entries, collectChildrenFromState(sc.Root)...)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FrontendEntry().Index() < entries[j].FrontendEntry().Index()
	})

	f.entries = entries
	return f
}

func (f *formatStatechart) FrontendEntry() frontend_common.FrontendEntry { return f.sc.FrontendData }

func (f *formatStatechart) Name() string                   { return fmt.Sprintf("statechart:%s", f.sc.Name) }
func (f *formatStatechart) Indents() bool                  { return true }
func (f *formatStatechart) ChildEntries() []formatterEntry { return f.entries }

func (f *formatStatechart) Write(w io.Writer) {}
func (f *formatStatechart) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Trigger -----------------------------------------------------------------------------------------

type formatTrigger struct {
	trigger *ir.Trigger
}

func newFormatTrigger(trigger *ir.Trigger) *formatTrigger {
	f := &formatTrigger{
		trigger: trigger,
	}

	return f
}

func (f *formatTrigger) FrontendEntry() frontend_common.FrontendEntry { return f.trigger.FrontendData }

func (f *formatTrigger) Name() string {
	return fmt.Sprintf("trigger:%s(index=%d)", f.trigger.Name, f.FrontendEntry().Index())
}
func (f *formatTrigger) Indents() bool                  { return false }
func (f *formatTrigger) ChildEntries() []formatterEntry { return nil }

func (f *formatTrigger) Write(w io.Writer) {}
func (f *formatTrigger) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// State -------------------------------------------------------------------------------------------

type formatState struct {
	state   *ir.State
	entries []formatterEntry
}

func newFormatState(state *ir.State) *formatState {
	f := &formatState{
		state: state,
	}

	f.entries = collectChildrenFromState(state)
	return f
}

func (f *formatState) FrontendEntry() frontend_common.FrontendEntry { return f.state.FrontendData }

func (f *formatState) Name() string {
	return fmt.Sprintf("state:%s(index=%d)", f.state.Name, f.FrontendEntry().Index())
}
func (f *formatState) Indents() bool                  { return true }
func (f *formatState) ChildEntries() []formatterEntry { return f.entries }

func (f *formatState) Write(w io.Writer) {}
func (f *formatState) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

func collectChildrenFromState(state *ir.State) []formatterEntry {
	var entries []formatterEntry

	for _, entry := range state.Children {
		entries = append(entries, newFormatterEntry(entry))
	}
	for _, entry := range state.Transitions {
		entries = append(entries, newFormatterEntry(entry))
	}
	for _, entry := range state.EnterReactions {
		entries = append(entries, newFormatterEntry(entry))
	}
	for _, entry := range state.ExitReactions {
		entries = append(entries, newFormatterEntry(entry))
	}
	for _, entry := range state.Captures {
		entries = append(entries, newFormatterEntry(entry))
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FrontendEntry().Index() < entries[j].FrontendEntry().Index()
	})

	return entries
}

// Reaction ----------------------------------------------------------------------------------------

type formatReaction struct {
	reaction *ir.StateReaction
}

func newFormatReaction(reaction *ir.StateReaction) *formatReaction {
	f := &formatReaction{
		reaction: reaction,
	}

	return f
}

func (f *formatReaction) FrontendEntry() frontend_common.FrontendEntry {
	return f.reaction.FrontendData
}

func (f *formatReaction) Name() string {
	return fmt.Sprintf("reaction:%s-%s(index=%d)", f.reaction.Kind, f.reaction.Trigger.SafeName(), f.FrontendEntry().Index())
}
func (f *formatReaction) Indents() bool                  { return false }
func (f *formatReaction) ChildEntries() []formatterEntry { return nil }

func (f *formatReaction) Write(w io.Writer) {}
func (f *formatReaction) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Capture -----------------------------------------------------------------------------------------

type formatCapture struct {
	capture *ir.StateCapture
}

func newFormatCapture(capture *ir.StateCapture) *formatCapture {
	f := &formatCapture{
		capture: capture,
	}

	return f
}

func (f *formatCapture) FrontendEntry() frontend_common.FrontendEntry { return f.capture.FrontendData }

func (f *formatCapture) Name() string {
	return fmt.Sprintf("capture:%s(index=%d)", f.capture.Trigger.SafeName(), f.FrontendEntry().Index())
}
func (f *formatCapture) Indents() bool                  { return false }
func (f *formatCapture) ChildEntries() []formatterEntry { return nil }

func (f *formatCapture) Write(w io.Writer) {}
func (f *formatCapture) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Transition --------------------------------------------------------------------------------------

type formatTransition struct {
	transition *ir.Transition
}

func newFormatTransition(transition *ir.Transition) *formatTransition {
	f := &formatTransition{
		transition: transition,
	}

	return f
}

func (f *formatTransition) FrontendEntry() frontend_common.FrontendEntry {
	return f.transition.FrontendData
}

func (f *formatTransition) Name() string {
	return fmt.Sprintf("transition:%s->%s(index=%d)", f.transition.Trigger.SafeName(), f.transition.To.SafeName(), f.FrontendEntry().Index())
}
func (f *formatTransition) Indents() bool                  { return false }
func (f *formatTransition) ChildEntries() []formatterEntry { return nil }

func (f *formatTransition) Write(w io.Writer) {}
func (f *formatTransition) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}
