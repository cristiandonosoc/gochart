package formatter

import (
	"fmt"
	"io"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/files"
	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// gochartFormatter --------------------------------------------------------------------------------

type gochartFormatter struct {
	iw    *IndentedWriter
	gc    *ir.Gochart
	debug bool

	writtenEntries []formatterEntry
}

func newGochartFormatter(iw *IndentedWriter, debug bool) *gochartFormatter {
	return &gochartFormatter{
		iw:    iw,
		debug: debug,
	}
}

func (gf *gochartFormatter) formatGochart(gc *ir.Gochart) {
	gf.gc = gc
	f := gf.newFormatterEntry(gc)
	gf.formatEntry(f)
}

func (gf *gochartFormatter) formatEntry(f formatterEntry) {
	if !gf.debug {
		f.Write(gf.iw)
	} else {
		f.DebugWrite(gf.iw)
	}
	gf.writtenEntries = append(gf.writtenEntries, f)

	if f.Indents() {
		gf.iw.MustIncreaseIndent()
	}

	for _, entry := range f.ChildEntries() {
		gf.formatEntry(entry)
	}

	if f.Indents() {
		gf.iw.MustDecreateIndent()
	}

	if !gf.debug {
		f.Pop(gf.iw)
	}
}

func (gf *gochartFormatter) lastWrittenEntry() formatterEntry {
	if len(gf.writtenEntries) == 0 {
		return nil
	}

	return gf.writtenEntries[len(gf.writtenEntries)-1]
}

// formatterEntry ----------------------------------------------------------------------------------

type formatterEntry interface {
	Formatter() *gochartFormatter
	FrontendEntry() frontend_common.FrontendEntry
	Name() string
	Indents() bool
	ChildEntries() []formatterEntry

	// Write writes when the nodes appears.
	Write(w io.Writer)
	// Pop writes when the node is being exited. Normally would write the closing }.
	Pop(w io.Writer)
	DebugWrite(w io.Writer)
}

func (gf *gochartFormatter) newFormatterEntry(e any) formatterEntry {
	switch v := e.(type) {
	case *ir.Gochart:
		return gf.newFormatGochartWrapper(v)
	case *ir.Trigger:
		return gf.newFormatTrigger(v)
	case *ir.State:
		return gf.newFormatState(v)
	case *ir.StateReaction:
		return gf.newFormatReaction(v)
	case *ir.StateCapture:
		return gf.newFormatCapture(v)
	case *ir.Transition:
		return gf.newFormatTransition(v)
	// Comment don't need a IR representation, so we use the frontend directly.
	case *frontend_common.CommentData:
		return gf.newFormatComment(v)
	default:
		panic(fmt.Sprintf("no formatter for type %q", reflect.TypeOf(v).Name()))
	}
}

// GochartWrapper -------------------------------------------------------------------------------
//
// Gochart has a wrapper because there can be comments outside the gochart and we don't have a
// "higher-level" entity above the gochart that contains both the gochart and the commnets.
// So in this case the gochart contains both the comments and external comments.

type formatGochartWrapper struct {
	gc      *ir.Gochart
	entries []formatterEntry
	gf      *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatGochartWrapper(gc *ir.Gochart) *formatGochartWrapper {
	f := &formatGochartWrapper{
		gc: gc,
		gf: gf,
	}

	// Create a child with the gochart as a child.

	entries := []formatterEntry{
		gf.newFormatGochart(gc),
	}

	for _, entry := range gc.FrontendData.Comments {
		entries = append(entries, gf.newFormatterEntry(entry))
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FrontendEntry().Index() < entries[j].FrontendEntry().Index()
	})

	f.entries = entries
	return f
}

// The frontend_common.FrontendEntry interface
func (f *formatGochartWrapper) Index() int { return 0 }
func (f *formatGochartWrapper) Position() *files.LoadedFilePosition {
	gcPos := f.gc.FrontendData.Position()

	// The wrapper is positioned at the first character of the first line.
	return &files.LoadedFilePosition{
		File: gcPos.File,
		Line: 1,
		Char: 1,
	}
}

func (f *formatGochartWrapper) Formatter() *gochartFormatter                 { return f.gf }
func (f *formatGochartWrapper) FrontendEntry() frontend_common.FrontendEntry { return f }
func (f *formatGochartWrapper) Name() string                                 { return "<gochart_wrapper>" }
func (f *formatGochartWrapper) Indents() bool                                { return false }
func (f *formatGochartWrapper) ChildEntries() []formatterEntry               { return f.entries }

func (f *formatGochartWrapper) Write(w io.Writer)      {}
func (f *formatGochartWrapper) Pop(w io.Writer)        {}
func (f *formatGochartWrapper) DebugWrite(w io.Writer) {}

// Gochart --------------------------------------------------------------------------------------

type formatGochart struct {
	gc      *ir.Gochart
	entries []formatterEntry
	gf      *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatGochart(gc *ir.Gochart) *formatGochart {
	f := &formatGochart{
		gc: gc,
		gf: gf,
	}

	var entries []formatterEntry

	for _, entry := range gc.Triggers {
		entries = append(entries, gf.newFormatterEntry(entry))
	}

	// Go over the root state ang gets the entries from it.
	entries = append(entries, gf.collectChildrenFromState(gc.Root)...)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FrontendEntry().Index() < entries[j].FrontendEntry().Index()
	})

	f.entries = entries
	return f
}

func (f *formatGochart) Formatter() *gochartFormatter { return f.gf }
func (f *formatGochart) FrontendEntry() frontend_common.FrontendEntry {
	return f.gc.Root.FrontendData
}
func (f *formatGochart) Name() string {
	return fmt.Sprintf("%d:gochart:%s", f.gc.Root.FrontendData.Index(), f.gc.Name)
}
func (f *formatGochart) Indents() bool                  { return true }
func (f *formatGochart) ChildEntries() []formatterEntry { return f.entries }

func (f *formatGochart) Write(w io.Writer) {
	collapseLineDistance(w, f)
	writeStringf(w, "gochart %s", f.gc.Name)
	writeOptions(w, f.gc.Options)
	writeStringf(w, "{ ")
}
func (f *formatGochart) Pop(w io.Writer) {
	collapseLineDistance(w, f)
	writeStringf(w, "}")
}
func (f *formatGochart) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Trigger -----------------------------------------------------------------------------------------

type formatTrigger struct {
	trigger *ir.Trigger
	gf      *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatTrigger(trigger *ir.Trigger) *formatTrigger {
	f := &formatTrigger{
		trigger: trigger,
		gf:      gf,
	}

	return f
}

func (f *formatTrigger) Formatter() *gochartFormatter                 { return f.gf }
func (f *formatTrigger) FrontendEntry() frontend_common.FrontendEntry { return f.trigger.FrontendData }
func (f *formatTrigger) Name() string {
	return fmt.Sprintf("%d:trigger:%s", f.FrontendEntry().Index(), f.trigger.SafeName())
}
func (f *formatTrigger) Indents() bool                  { return false }
func (f *formatTrigger) ChildEntries() []formatterEntry { return nil }

func (f *formatTrigger) Write(w io.Writer) {
	collapseLineDistance(w, f)
	writeStringf(w, "trigger %s(%s) ", f.trigger.Name, f.trigger.FrontendData.ArgsString)
}
func (f *formatTrigger) Pop(w io.Writer) {}
func (f *formatTrigger) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// State -------------------------------------------------------------------------------------------

type formatState struct {
	state   *ir.State
	entries []formatterEntry
	gf      *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatState(state *ir.State) *formatState {
	f := &formatState{
		state: state,
		gf:    gf,
	}

	f.entries = gf.collectChildrenFromState(state)
	return f
}

func (f *formatState) Formatter() *gochartFormatter                 { return f.gf }
func (f *formatState) FrontendEntry() frontend_common.FrontendEntry { return f.state.FrontendData }

func (f *formatState) Name() string {
	return fmt.Sprintf("%d:state:%s", f.FrontendEntry().Index(), f.state.SafeName())
}
func (f *formatState) Indents() bool                  { return true }
func (f *formatState) ChildEntries() []formatterEntry { return f.entries }

func (f *formatState) Write(w io.Writer) {
	collapseLineDistance(w, f)
	writeStringf(w, "state %s", f.state.Name)
	writeOptions(w, f.state.Options)
	writeStringf(w, "{ ")

}
func (f *formatState) Pop(w io.Writer) {
	collapseLineDistance(w, f)
	writeStringf(w, "}")
}
func (f *formatState) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

func (gf *gochartFormatter) collectChildrenFromState(state *ir.State) []formatterEntry {
	var entries []formatterEntry

	for _, entry := range state.Children {
		entries = append(entries, gf.newFormatterEntry(entry))
	}
	for _, entry := range state.Transitions {
		entries = append(entries, gf.newFormatterEntry(entry))
	}
	for _, entry := range state.EnterReactions {
		entries = append(entries, gf.newFormatterEntry(entry))
	}
	for _, entry := range state.ExitReactions {
		entries = append(entries, gf.newFormatterEntry(entry))
	}
	for _, entry := range state.Captures {
		entries = append(entries, gf.newFormatterEntry(entry))
	}
	for _, entry := range state.FrontendData.Comments {
		entries = append(entries, gf.newFormatterEntry(entry))
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FrontendEntry().Index() < entries[j].FrontendEntry().Index()
	})

	return entries
}

// Reaction ----------------------------------------------------------------------------------------

type formatReaction struct {
	reaction *ir.StateReaction
	gf       *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatReaction(reaction *ir.StateReaction) *formatReaction {
	f := &formatReaction{
		reaction: reaction,
		gf:       gf,
	}

	return f
}

func (f *formatReaction) Formatter() *gochartFormatter { return f.gf }
func (f *formatReaction) FrontendEntry() frontend_common.FrontendEntry {
	return f.reaction.FrontendData
}

func (f *formatReaction) Name() string {
	return fmt.Sprintf("%d:reaction:%s-%s", f.FrontendEntry().Index(), f.reaction.Kind, f.reaction.Trigger.SafeName())
}
func (f *formatReaction) Indents() bool                  { return false }
func (f *formatReaction) ChildEntries() []formatterEntry { return nil }

func (f *formatReaction) Write(w io.Writer) {
	collapseLineDistance(w, f)

	writeStringf(w, "%s ", strings.ToLower(f.reaction.Kind.String()))

	if f.reaction.Trigger != nil {
		writeStringf(w, "%s ", f.reaction.Trigger.Name)
	}
}
func (f *formatReaction) Pop(w io.Writer) {}
func (f *formatReaction) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Capture -----------------------------------------------------------------------------------------

type formatCapture struct {
	capture *ir.StateCapture
	gf      *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatCapture(capture *ir.StateCapture) *formatCapture {
	f := &formatCapture{
		capture: capture,
		gf:      gf,
	}

	return f
}

func (f *formatCapture) Formatter() *gochartFormatter                 { return f.gf }
func (f *formatCapture) FrontendEntry() frontend_common.FrontendEntry { return f.capture.FrontendData }

func (f *formatCapture) Name() string {
	return fmt.Sprintf("%d:capture:%s", f.FrontendEntry().Index(), f.capture.Trigger.SafeName())
}
func (f *formatCapture) Indents() bool                  { return false }
func (f *formatCapture) ChildEntries() []formatterEntry { return nil }

func (f *formatCapture) Write(w io.Writer) {
	collapseLineDistance(w, f)

	writeStringf(w, "capture %s ", f.capture.Trigger.Name)
	if f.capture.HasCondition() {
		writeCondition(w, f.capture.Condition)
	}
}
func (f *formatCapture) Pop(w io.Writer) {}
func (f *formatCapture) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Transition --------------------------------------------------------------------------------------

type formatTransition struct {
	transition *ir.Transition
	gf         *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatTransition(transition *ir.Transition) *formatTransition {
	f := &formatTransition{
		transition: transition,
		gf:         gf,
	}

	return f
}

func (f *formatTransition) Formatter() *gochartFormatter { return f.gf }
func (f *formatTransition) FrontendEntry() frontend_common.FrontendEntry {
	return f.transition.FrontendData
}

func (f *formatTransition) Name() string {
	return fmt.Sprintf("%d:transition:%s->%s", f.FrontendEntry().Index(), f.transition.Trigger.SafeName(), f.transition.To.SafeName())
}
func (f *formatTransition) Indents() bool                  { return false }
func (f *formatTransition) ChildEntries() []formatterEntry { return nil }

func (f *formatTransition) Write(w io.Writer) {
	collapseLineDistance(w, f)

	writeStringf(w, "transition ")
	triggerName := "null"
	if f.transition.Trigger != nil {
		triggerName = f.transition.Trigger.Name
	}
	writeStringf(w, "%s > %s ", triggerName, f.transition.To.Name)

	if f.transition.HasCondition() {
		writeCondition(w, f.transition.Condition)
	}

	if f.transition.HasAction() {
		writeAction(w, f.transition.Action)
	}
}
func (f *formatTransition) Pop(w io.Writer) {}
func (f *formatTransition) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Comment -----------------------------------------------------------------------------------------

type formatComment struct {
	comment *frontend_common.CommentData
	gf      *gochartFormatter // Back pointer.
}

func (gf *gochartFormatter) newFormatComment(comment *frontend_common.CommentData) *formatComment {
	f := &formatComment{
		comment: comment,
		gf:      gf,
	}

	return f
}

func (f *formatComment) Formatter() *gochartFormatter { return f.gf }
func (f *formatComment) FrontendEntry() frontend_common.FrontendEntry {
	return f.comment
}

func (f *formatComment) Name() string {
	return fmt.Sprintf("%d:comment:%q", f.FrontendEntry().Index(), f.comment.Comment)
}
func (f *formatComment) Indents() bool                  { return false }
func (f *formatComment) ChildEntries() []formatterEntry { return nil }

func (f *formatComment) Write(w io.Writer) {
	collapseLineDistance(w, f, lineCollapse_AllowInline)
	writeStringf(w, "// %s", f.comment.Comment)
}
func (f *formatComment) Pop(w io.Writer) {}
func (f *formatComment) DebugWrite(w io.Writer) {
	writeStringf(w, "%s\n", f.Name())
}

// Helpers -----------------------------------------------------------------------------------------

type lineCollapseOption int

const (
	lineCollapse_MostOneLine lineCollapseOption = iota
	lineCollapse_AllowInline
)

func collapseLineDistance(w io.Writer, entry formatterEntry, opts ...lineCollapseOption) {
	distance := lineDistanceWithLastEntry(entry)
	// Negative distance means that the entry was written long ago.
	// This is most likely because of a pop statement.
	if distance < 0 {
		if slices.Contains(opts, lineCollapse_AllowInline) {
			return
		}
		writeStringf(w, "\n")
	} else if distance == 0 {
		if slices.Contains(opts, lineCollapse_AllowInline) {
			return
		}
		writeStringf(w, "\n")
	} else if distance == 1 {
		writeStringf(w, "\n")
	} else if distance > 1 {
		// If we were separated by more than one new line, we collapse it to one new line.
		if !slices.Contains(opts, lineCollapse_MostOneLine) {
			writeStringf(w, "\n")
		}
		writeStringf(w, "\n")
	}
}

func lineDistanceWithLastEntry(entry formatterEntry) int {
	lastEntry := entry.Formatter().lastWrittenEntry()
	if lastEntry == nil {
		return 0
	}

	last := lastEntry.FrontendEntry().Position().Line
	current := entry.FrontendEntry().Position().Line

	distance := current - last

	// We special case if the previous line was only a '{', we consider it to be one less.
	if distance > 1 {
		// We actually want the line before current, but because line numbers are 1-indexed, just using
		// the current value will give us the previous line.
		if i := current - 2; i >= 0 {
			lines, _ := entry.FrontendEntry().Position().File.Lines()
			line := lines[i]
			line = strings.TrimSpace(line)
			if line == "{" {
				distance--
			}
		}
	}

	return distance
}

func writeOptions(w io.Writer, options []*ir.Option) {
	// If there are no options, we simply write a space, as that is expected on the stream.
	if len(options) == 0 {
		writeStringf(w, " ")
		return
	}

	// Collect the option strings
	optionStrings := make([]string, 0, len(options))
	for _, option := range options {
		optionString := ""
		switch kind := option.Definition.Kind; kind {
		case frontend_common.OptionValueKind_Boolean:
			if option.IsImplicitBool() {
				optionString = option.Name()
			} else {
				optionString = fmt.Sprintf("%s = %t", option.Name(), option.BoolValue())
			}
		case frontend_common.OptionValueKind_String:
			optionString = fmt.Sprintf("%s = %q", option.Name(), option.StringValue())
		case frontend_common.OptionValueKind_Integer:
			optionString = fmt.Sprintf("%s = %d", option.Name(), option.IntValue())
		}

		optionStrings = append(optionStrings, optionString)
	}

	writeStringf(w, "(%s) ", strings.Join(optionStrings, ", "))
}

func writeCondition(w io.Writer, condition string) {
	// We only want to have one space before and after.
	condition = strings.TrimFunc(condition, func(r rune) bool {
		return r == ' '
	})
	writeStringf(w, "if { ")
	writeStringfNoIndent(w, condition)
	writeStringf(w, " } ")
}

func writeAction(w io.Writer, action string) {
	// We only want to have one space before and after.
	action = strings.TrimFunc(action, func(r rune) bool {
		return r == ' '
	})

	writeStringf(w, "do { ")
	writeStringfNoIndent(w, action)
	writeStringf(w, " } ")
}
