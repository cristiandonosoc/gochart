// Package diagram draws a gochart as a Mermaid state diagram (stateDiagram-v2), for docs and
// reviews. It works from the IR, so it draws gocharts of any language.
package diagram

import (
	"fmt"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// kAnyState is the box that stands for the gochart level, which handles triggers in every state.
const kAnyState = "__gochart"

// Mermaid returns the diagram of |gc|.
//
// Composite states hold their children, with an arrow from [*] to the default child, and parallel
// regions are separated by "--". Edges are labeled "Trigger [condition] / action" (null
// transitions have no trigger). Enter and exit reactions, captures and state data are listed in
// the state's box. Transitions and captures of the gochart level start from an "any state" box.
// Conditions and actions written as inlined code (C++) are shown as "…".
func Mermaid(gc *ir.Gochart) string {
	d := &drawer{}
	d.line("stateDiagram-v2")
	d.indent++

	root := gc.Root
	if root.Parallel {
		// Mermaid only separates regions inside a composite state, so the gochart gets one.
		name := safeId(gc.Name)
		d.line("[*] --> %s", name)
		d.drawComposite(name, root)
	} else {
		d.drawChildren(root)
	}

	if hasLevelEntries(root) {
		d.line("state \"any state\" as %s", kAnyState)
		d.describe(kAnyState, root)
	}
	d.drawTransitions(root)

	return d.b.String()
}

type drawer struct {
	b      strings.Builder
	indent int
}

func (d *drawer) line(format string, args ...any) {
	d.b.WriteString(strings.Repeat("    ", d.indent))
	fmt.Fprintf(&d.b, format, args...)
	d.b.WriteString("\n")
}

// drawChildren draws the children of |s| inside the current block.
func (d *drawer) drawChildren(s *ir.State) {
	if len(s.Children) == 0 {
		return
	}

	if s.Parallel {
		for i, child := range s.Children {
			if i > 0 {
				d.line("--")
			}
			d.drawState(child)
		}
		return
	}

	d.line("[*] --> %s", safeId(s.DefaultChild().Name))
	for _, child := range s.Children {
		d.drawState(child)
	}
}

func (d *drawer) drawState(s *ir.State) {
	id := safeId(s.Name)
	if len(s.Children) == 0 {
		d.line("%s", id)
		d.describe(id, s)
		return
	}
	d.drawComposite(id, s)
}

func (d *drawer) drawComposite(id string, s *ir.State) {
	d.line("state %s {", id)
	d.indent++
	d.drawChildren(s)
	d.indent--
	d.line("}")
	if !s.IsRoot() {
		d.describe(id, s)
	}
}

// describe lists what |s| does without leaving: its data, reactions and captures.
func (d *drawer) describe(id string, s *ir.State) {
	if s.DataType != "" {
		d.line("%s : data %s", id, escape(s.DataType))
	}
	for _, reaction := range s.EnterReactions {
		d.line("%s : %s", id, reactionLabel("enter", reaction))
	}
	for _, reaction := range s.ExitReactions {
		d.line("%s : %s", id, reactionLabel("exit", reaction))
	}
	for _, capture := range s.Captures {
		label := "capture " + capture.Trigger.Name
		if condition := conditionLabel(capture.Condition, capture.ConditionCallback); condition != "" {
			label += " " + condition
		}
		d.line("%s : %s", id, label)
	}
}

// drawTransitions draws the edges of |s| and of every state below it, at the top level (state
// ids are global in Mermaid, and an edge inside a composite state would create its endpoints
// there).
func (d *drawer) drawTransitions(s *ir.State) {
	from := kAnyState
	if !s.IsRoot() {
		from = safeId(s.Name)
	}
	for _, transition := range s.Transitions {
		d.line("%s --> %s%s", from, safeId(transition.To.Name), transitionLabel(transition))
	}
	for _, child := range s.Children {
		d.drawTransitions(child)
	}
}

func hasLevelEntries(root *ir.State) bool {
	return len(root.Transitions) > 0 || len(root.Captures) > 0 || len(root.EnterReactions) > 0 ||
		len(root.ExitReactions) > 0
}

func reactionLabel(kind string, reaction *ir.StateReaction) string {
	if reaction.Trigger == nil {
		return kind
	}
	return kind + " on " + reaction.Trigger.Name
}

func transitionLabel(t *ir.Transition) string {
	var parts []string
	if !t.IsNullTransition() {
		parts = append(parts, t.Trigger.Name)
	}
	if condition := conditionLabel(t.Condition, t.ConditionCallback); condition != "" {
		parts = append(parts, condition)
	}
	if t.ActionCallback != "" {
		parts = append(parts, "/ "+t.ActionCallback)
	} else if t.Action != "" {
		parts = append(parts, "/ …")
	}
	if len(parts) == 0 {
		return ""
	}
	return " : " + strings.Join(parts, " ")
}

func conditionLabel(code, callback string) string {
	switch {
	case callback != "":
		return "[" + callback + "]"
	case code != "":
		return "[…]"
	}
	return ""
}

// safeId keeps Mermaid keywords from being read as syntax.
func safeId(name string) string {
	switch strings.ToLower(name) {
	case "state", "note", "end", "direction", "classdef", "class", "style":
		return name + "_"
	}
	return name
}

// escape writes the characters Mermaid would read as markup (eg. the generics of a data type) as
// Mermaid entities.
func escape(text string) string {
	return strings.NewReplacer("<", "#lt;", ">", "#gt;", ":", "#58;", ";", "#59;").Replace(text)
}
