package rust

import (
	"fmt"
	"strings"
)

// writer is a small helper to emit indented code.
type writer struct {
	b      strings.Builder
	indent int
}

func (w *writer) line(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	if text == "" {
		w.b.WriteString("\n")
		return
	}
	w.b.WriteString(strings.Repeat("    ", w.indent))
	w.b.WriteString(text)
	w.b.WriteString("\n")
}

func (w *writer) raw(text string) { w.b.WriteString(text) }

func (w *writer) open(format string, args ...any) {
	w.line(format, args...)
	w.indent++
}

func (w *writer) close(text string) {
	w.indent--
	w.line("%s", text)
}

// Path binding ------------------------------------------------------------------------------------
//
// The chart is a tree of nested enums (or structs, for parallel states). To reach the node of a
// state, we bind the nodes along its path: "n1" for the top level state, "n2" for its child, etc.

type bindMode struct {
	mutable bool
	// mismatch is what to do if the state is not active ("unreachable!()", "return false", ...).
	mismatch string
}

var (
	bindRefActive = bindMode{mutable: false, mismatch: `unreachable!("state is active")`}
	bindMutActive = bindMode{mutable: true, mismatch: `unreachable!("state is active")`}
)

func (b bindMode) ref() string {
	if b.mutable {
		return "&mut "
	}
	return "&"
}

func nodeVar(s *rustState) string { return fmt.Sprintf("n%d", s.Depth) }

// bindTree binds |tree| to the root children.
func bindTree(w *writer, mode bindMode) {
	accessor := "as_ref"
	if mode.mutable {
		accessor = "as_mut"
	}
	bindSome(w, "tree", fmt.Sprintf("c.tree.%s()", accessor), mode)
}

// bindSome binds |name| to the content of the |expr| option.
func bindSome(w *writer, name, expr string, mode bindMode) {
	if mode.mismatch == "return None" {
		w.line("let %s = %s?;", name, expr)
		return
	}
	w.line("let Some(%s) = %s else { %s };", name, expr, mode.mismatch)
}

// childrenExpr is the expression for the children of |parent|, once it is bound.
func childrenExpr(parent *rustState, mode bindMode) string {
	if parent.isRoot {
		return "tree"
	}
	return fmt.Sprintf("%s%s.children", mode.ref(), nodeVar(parent))
}

// bindPath binds the nodes from the top level down to |target| (included). The root is not bound
// beyond |tree|.
func bindPath(w *writer, target *rustState, mode bindMode) {
	bindTree(w, mode)
	if target == nil || target.isRoot {
		return
	}

	for _, state := range target.Path {
		parent := state.Parent
		if parent.IsParallel() {
			accessor := "as_ref"
			if mode.mutable {
				accessor = "as_mut"
			}
			var expr string
			if parent.isRoot {
				expr = fmt.Sprintf("tree.%s.%s()", state.FieldName(), accessor)
			} else {
				expr = fmt.Sprintf("%s.children.%s.%s()", nodeVar(parent), state.FieldName(), accessor)
			}
			bindSome(w, nodeVar(state), expr, mode)
		} else {
			w.line("let %s::%s(%s) = %s else { %s };",
				parent.ChildrenType(), state.Name, nodeVar(state), childrenExpr(parent, mode), mode.mismatch)
		}
	}
}

// bindForData binds the path down to the deepest state of |states| (which are on a path), and
// returns the expressions for their data.
func bindForData(w *writer, states []*rustState, mode bindMode) []string {
	if len(states) == 0 {
		return nil
	}
	bindPath(w, states[len(states)-1], mode)
	exprs := make([]string, 0, len(states))
	for _, state := range states {
		exprs = append(exprs, fmt.Sprintf("%s%s.data", mode.ref(), nodeVar(state)))
	}
	return exprs
}

func call(method string, args []string) string {
	return fmt.Sprintf("%s(%s)", method, strings.Join(args, ", "))
}
