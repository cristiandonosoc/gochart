package rust

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cristiandonosoc/gochart/pkg/ir"
)

// The model is a view of the IR with everything the Rust generation needs precomputed: Rust names,
// the tree layout and the callbacks the delegate has to implement.

type rustState struct {
	ir *ir.State

	isRoot bool
	// Name is the state name as written in the gochart ("ToPickup"). Empty for the root.
	Name string
	// Snake is the prefix of the callbacks of the state ("to_pickup"). "gochart" for the root.
	Snake string

	Parent   *rustState
	Children []*rustState
	// Path goes from the top level state down to this one. The root is not part of any path.
	Path []*rustState

	Depth  int
	Region int

	DataType    string
	DataDefault bool
}

func (s *rustState) HasData() bool     { return s.DataType != "" }
func (s *rustState) IsLeaf() bool      { return len(s.Children) == 0 }
func (s *rustState) IsParallel() bool  { return s.ir.Parallel }
func (s *rustState) HasChildren() bool { return len(s.Children) > 0 }

// StateId is the Rust expression for the state as an |Option<StateId>|.
func (s *rustState) OptionId() string {
	if s.isRoot {
		return "None"
	}
	return fmt.Sprintf("Some(StateId::%s)", s.Name)
}

func (s *rustState) NodeType() string { return s.Name + "Node" }

func (s *rustState) ChildrenType() string {
	if s.isRoot {
		return "ChartChildren"
	}
	return s.Name + "Children"
}

// FieldName is the name of the field holding this state in a parallel parent.
func (s *rustState) FieldName() string { return s.Snake }

// DataPath is the states with data from the top level down to this state (included).
func (s *rustState) DataPath() []*rustState {
	var result []*rustState
	for _, state := range s.Path {
		if state.HasData() {
			result = append(result, state)
		}
	}
	return result
}

// AncestorDataPath is like |DataPath|, but without the state itself.
func (s *rustState) AncestorDataPath() []*rustState {
	path := s.DataPath()
	if len(path) > 0 && path[len(path)-1] == s {
		return path[:len(path)-1]
	}
	return path
}

type rustTrigger struct {
	ir *ir.Trigger

	Name  string
	Snake string
	Args  []*ir.TriggerArgument
}

// Pattern returns the match pattern for the event of this trigger, binding the arguments to
// |__a0|, |__a1|, etc. (so they cannot shadow the generated code variables).
func (t *rustTrigger) Pattern() string {
	if len(t.Args) == 0 {
		return fmt.Sprintf("Event::%s", t.Name)
	}
	fields := make([]string, 0, len(t.Args))
	for i, arg := range t.Args {
		fields = append(fields, fmt.Sprintf("%s: __a%d", arg.Name, i))
	}
	return fmt.Sprintf("Event::%s { %s }", t.Name, strings.Join(fields, ", "))
}

// ArgBindings is the list of variables bound by |Pattern|.
func (t *rustTrigger) ArgBindings() []string {
	bindings := make([]string, 0, len(t.Args))
	for i := range t.Args {
		bindings = append(bindings, fmt.Sprintf("__a%d", i))
	}
	return bindings
}

type model struct {
	gc *ir.Gochart

	Root   *rustState
	States []*rustState // Every state but the root, in definition order.
	Leaves []*rustState

	Triggers   []*rustTrigger
	triggerMap map[*ir.Trigger]*rustTrigger
	stateMap   map[*ir.State]*rustState

	RegionCount int
	MaxDepth    int

	Methods []*delegateMethod
}

func newModel(gc *ir.Gochart) (*model, error) {
	if gc.Language != ir.Language_Rust {
		return nil, fmt.Errorf("gochart %q is written for language %q, the Rust backend needs language = %q",
			gc.Name, gc.Language, ir.Language_Rust)
	}

	m := &model{
		gc:          gc,
		triggerMap:  make(map[*ir.Trigger]*rustTrigger),
		stateMap:    make(map[*ir.State]*rustState),
		RegionCount: gc.ParallelGroupCount(),
	}

	if err := m.collectTriggers(); err != nil {
		return nil, err
	}
	if err := m.collectStates(); err != nil {
		return nil, err
	}
	if err := m.collectMethods(); err != nil {
		return nil, err
	}
	if err := m.checkNames(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *model) collectTriggers() error {
	snakes := map[string]string{}
	for _, trigger := range m.gc.Triggers {
		if !ir.IsRustIdentifier(trigger.Name) {
			return fmt.Errorf("trigger %q: not a valid Rust identifier", trigger.Name)
		}
		snake := toSnakeCase(trigger.Name)
		if other, ok := snakes[snake]; ok {
			return fmt.Errorf("triggers %q and %q have the same Rust name %q", other, trigger.Name, snake)
		}
		snakes[snake] = trigger.Name

		for _, arg := range trigger.Args {
			if arg.Name == "delegate" {
				return fmt.Errorf("trigger %q: argument name \"delegate\" is reserved", trigger.Name)
			}
		}

		rt := &rustTrigger{
			ir:    trigger,
			Name:  trigger.Name,
			Snake: snake,
			Args:  trigger.Args,
		}
		m.Triggers = append(m.Triggers, rt)
		m.triggerMap[trigger] = rt
	}
	return nil
}

func (m *model) collectStates() error {
	root := &rustState{
		ir:     m.gc.Root,
		isRoot: true,
		Snake:  "gochart",
	}
	m.Root = root
	m.stateMap[m.gc.Root] = root

	snakes := map[string]string{"gochart": "<gochart>"}
	for _, state := range m.gc.States {
		if state == m.gc.Root {
			continue
		}

		if !ir.IsRustIdentifier(state.Name) {
			return fmt.Errorf("state %q: not a valid Rust identifier", state.Name)
		}
		snake := toSnakeCase(state.Name)
		if isRustKeyword(snake) {
			return fmt.Errorf("state %q: its Rust name %q is a keyword", state.Name, snake)
		}
		if other, ok := snakes[snake]; ok {
			return fmt.Errorf("states %q and %q have the same Rust name %q", other, state.Name, snake)
		}
		snakes[snake] = state.Name

		rs := &rustState{
			ir:          state,
			Name:        state.Name,
			Snake:       snake,
			Region:      state.GetParallelGroup(),
			DataType:    state.DataType,
			DataDefault: state.DataDefault,
		}
		m.States = append(m.States, rs)
		m.stateMap[state] = rs
	}

	// Parenthood, in definition order.
	for _, rs := range append([]*rustState{root}, m.States...) {
		for _, child := range rs.ir.Children {
			crs := m.stateMap[child]
			crs.Parent = rs
			rs.Children = append(rs.Children, crs)
		}
	}

	for _, rs := range m.States {
		var path []*rustState
		for current := rs; current != nil && !current.isRoot; current = current.Parent {
			path = append([]*rustState{current}, path...)
		}
		rs.Path = path
		rs.Depth = len(path)
		if rs.Depth > m.MaxDepth {
			m.MaxDepth = rs.Depth
		}

		if rs.IsLeaf() {
			m.Leaves = append(m.Leaves, rs)
		}

		// C++ evaluates null transitions only when entering leaves. We keep the same semantics, but
		// rather than silently ignoring them, we reject them on other states.
		if !rs.IsLeaf() && len(rs.ir.GetNullTransitions()) > 0 {
			return fmt.Errorf("state %q: null transitions are only evaluated on leaf states", rs.Name)
		}
	}
	if len(root.ir.GetNullTransitions()) > 0 {
		return fmt.Errorf("gochart %q: null transitions are only evaluated on leaf states", m.gc.Name)
	}

	return nil
}

// HandledEntry is a transition or capture of a state for a given trigger.
type handledEntry struct {
	Transition *ir.Transition
	Capture    *ir.StateCapture
	index      int
}

// HandledEntries returns the transitions and captures of |state| for |trigger| in the order they
// were written, which is the order in which they are evaluated.
func (m *model) HandledEntries(state *rustState, trigger *rustTrigger) []*handledEntry {
	var entries []*handledEntry
	for _, transition := range state.ir.Transitions {
		if transition.Trigger == trigger.ir {
			entries = append(entries, &handledEntry{Transition: transition, index: transition.FrontendData.Frontend_Index})
		}
	}
	for _, capture := range state.ir.Captures {
		if capture.Trigger == trigger.ir {
			entries = append(entries, &handledEntry{Capture: capture, index: capture.FrontendData.Frontend_Index})
		}
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].index < entries[j].index })
	return entries
}

func (m *model) AllStates() []*rustState {
	return append([]*rustState{m.Root}, m.States...)
}

// checkNames verifies that the generated names don't collide with each other, and that the types
// written by the user are not shadowed by generated types.
func (m *model) checkNames() error {
	generated := map[string]string{
		"StateId":       "the generated state enum",
		"Leaf":          "the generated leaf enum",
		"TriggerId":     "the generated trigger enum",
		"Event":         "the generated event enum",
		"Chart":         "the generated chart",
		"Delegate":      "the generated delegate trait",
		"ChartChildren": "the generated children of the gochart level",
		"RtEntered":     "a generated runtime helper",
	}
	addType := func(name, what string) error {
		if other, ok := generated[name]; ok {
			return fmt.Errorf("type %q (%s) collides with %s", name, what, other)
		}
		generated[name] = what
		return nil
	}
	for _, state := range m.States {
		if err := addType(state.NodeType(), fmt.Sprintf("the generated node of state %q", state.Name)); err != nil {
			return err
		}
		if state.HasChildren() {
			if err := addType(state.ChildrenType(), fmt.Sprintf("the generated children of state %q", state.Name)); err != nil {
				return err
			}
		}
	}

	checkType := func(typeExpr, where string) error {
		for _, id := range leadingIdentifiers(typeExpr) {
			if what, ok := generated[id]; ok {
				return fmt.Errorf("%s: type %q is shadowed by %s, use a path instead (eg. \"super::%s\")",
					where, id, what, id)
			}
		}
		return nil
	}
	for _, state := range m.States {
		if state.HasData() {
			if err := checkType(state.DataType, fmt.Sprintf("data of state %q", state.Name)); err != nil {
				return err
			}
		}
	}
	for _, trigger := range m.Triggers {
		for _, arg := range trigger.Args {
			if err := checkType(arg.Type, fmt.Sprintf("argument %q of trigger %q", arg.Name, trigger.Name)); err != nil {
				return err
			}
		}
	}

	// Chart methods.
	chartMethods := map[string]string{
		"is_active": "a chart query",
		"is_in":     "a chart query",
		"leaf":      "a chart query",
		"leaves":    "a chart query",
	}
	for _, state := range m.States {
		if !state.HasData() {
			continue
		}
		for _, name := range []string{state.Snake, state.Snake + "_mut"} {
			if other, ok := chartMethods[name]; ok {
				return fmt.Errorf("data getter %q of state %q collides with %s", name, state.Name, other)
			}
			chartMethods[name] = fmt.Sprintf("the data getter of state %q", state.Name)
		}
	}

	return nil
}
