package ir

import (
	"fmt"
	"slices"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/frontend/gochart_lang"
	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inputTestHelper struct {
	triggers []*Trigger
	states   []*State
}

func TestReadSimpleYaml(t *testing.T) {
	wantTriggers := []*Trigger{
		{
			Name: "Trigger1",
			Args: []*TriggerArgument{
				{
					Type: "int",
					Name: "foo",
				},
				{
					Type: "float",
					Name: "bar",
				},
			},
		},
		{
			Name: "Trigger2",
		},
	}

	wantStates := []*State{
		{
			Name: "StateA",
		},
		{
			Name:    "StateB",
			Default: true,
		},
		{
			Name: "StateC",
		},
	}

	ith := &inputTestHelper{
		triggers: wantTriggers,
		states:   wantStates,
	}

	// StateA
	ith.setEnterReaction(t, "StateA", "Trigger1")
	ith.setEnterReaction(t, "StateA", "")
	ith.setExitReaction(t, "StateA", "Trigger1")
	ith.setExitReaction(t, "StateA", "")
	ith.setCapture(t, "StateA", "Trigger1", "foo > 5")
	ith.setCapture(t, "StateA", "Trigger2", "")
	ith.setParenthood(t, "StateA", "StateB")
	ith.setParenthood(t, "StateA", "StateC")
	// StateB
	ith.setExitReaction(t, "StateB", "Trigger1")
	ith.setTransition(t, "StateB", "StateC", "Trigger1", "", "")
	ith.setTransition(t, "StateB", "StateA", "Trigger2", "return owner->SomeCondition();", "owner->SomeAction();")
	// StateC
	ith.setEnterReaction(t, "StateC", "Trigger1")
	ith.setEnterReaction(t, "StateC", "")
	ith.setTransition(t, "StateC", "StateB", "", "", "StateC->Action();")

	yf, err := yaml.NewYamlFrontend()
	require.NoError(t, err)

	path, err := test_support.RunfilePath("testdata/simple.yaml")
	require.NoError(t, err)

	lf, err := files.LoadFileFromPath(path)
	require.NoError(t, err)

	scdata, err := yf.Process(lf)
	require.NoError(t, err)
	require.Equal(t, len(wantTriggers), len(scdata.Triggers))
	require.Equal(t, len(wantStates), len(scdata.States))

	sc, err := ProcessStatechartData(scdata)
	require.NoError(t, err)

	assert.Equal(t, "Simple", sc.Name)

	if assert.Equal(t, len(wantTriggers), len(sc.Triggers), "Different amount of triggers") {
		for i := 0; i < len(wantTriggers); i++ {
			want := wantTriggers[i]
			got := sc.Triggers[i]

			compareTrigger(t, want, got)
		}
	}

	wantRootStates := []*State{
		ith.findState(t, "StateA"),
	}

	if assert.Equal(t, len(wantRootStates), len(sc.Roots), "Different amount of root states") {
		for i := 0; i < len(wantRootStates); i++ {
			want := wantRootStates[i]
			got := sc.Roots[i]

			compareState(t, want, got)
		}
	}
}

func TestConcurrentGroups(t *testing.T) {
	wantTriggers := []*Trigger{
		{
			Name: "SomeTrigger",
		},
	}

	wantStates := []*State{
		{
			Name:            "A",
			ConcurrentGroup: 0,
		},
		{
			Name:            "B",
			Default:         true,
			ConcurrentGroup: 0,
		},
		{
			Name:            "C",
			Concurrent:      true,
			ConcurrentGroup: 0,
		},
		{
			Name:            "D",
			Default:         true,
			Concurrent:      true,
			ConcurrentGroup: 0,
		},
		{
			Name:            "E",
			ConcurrentGroup: 3,
		},
		{
			Name:            "F",
			Concurrent:      true,
			ConcurrentGroup: 4,
		},
		{
			Name:            "G",
			ConcurrentGroup: 7,
		},
		{
			Name:            "H",
			Default:         true,
			ConcurrentGroup: 1,
		},
		{
			Name:            "I",
			ConcurrentGroup: 2,
		},
		{
			Name:            "J",
			Default:         true,
			ConcurrentGroup: 3,
		},
		{
			Name:            "K",
			ConcurrentGroup: 3,
		},
		{
			Name:            "L",
			Default:         true,
			ConcurrentGroup: 5,
		},
		{
			Name:            "M",
			ConcurrentGroup: 6,
		},
		{
			Name:            "N",
			Default:         true,
			ConcurrentGroup: 7,
		},
		{
			Name:            "O",
			ConcurrentGroup: 7,
		},
	}

	ith := &inputTestHelper{
		triggers: wantTriggers,
		states:   wantStates,
	}

	ith.setParenthood(t, "A", "B", "C")
	ith.setParenthood(t, "B", "D")
	ith.setParenthood(t, "C", "E", "F", "G")
	ith.setParenthood(t, "D", "H", "I")
	ith.setParenthood(t, "E", "J", "K")
	ith.setParenthood(t, "F", "L", "M")
	ith.setParenthood(t, "G", "N", "O")

	ith.setCapture(t, "H", "SomeTrigger", "")
	ith.setCapture(t, "I", "SomeTrigger", "")
	ith.setCapture(t, "J", "SomeTrigger", "")
	ith.setCapture(t, "K", "SomeTrigger", "")
	ith.setCapture(t, "L", "SomeTrigger", "")
	ith.setCapture(t, "M", "SomeTrigger", "")
	ith.setCapture(t, "N", "SomeTrigger", "")
	ith.setCapture(t, "O", "SomeTrigger", "")

	yf, err := gochart_lang.NewGochartLangFrontend()
	require.NoError(t, err)

	path, err := test_support.RunfilePath("testdata/concurrent.gochart")
	require.NoError(t, err)

	lf, err := files.LoadFileFromPath(path)
	require.NoError(t, err)

	scdata, err := yf.Process(lf)
	require.NoError(t, err)
	require.Equal(t, len(wantTriggers), len(scdata.Triggers))
	require.Equal(t, len(wantStates), len(scdata.States))

	sc, err := ProcessStatechartData(scdata)
	require.NoError(t, err)

	assert.Equal(t, "Concurrent", sc.Name)

	if assert.Equal(t, len(wantTriggers), len(sc.Triggers), "Different amount of triggers") {
		for i := 0; i < len(wantTriggers); i++ {
			want := wantTriggers[i]
			got := sc.Triggers[i]

			compareTrigger(t, want, got)
		}
	}

	wantRootStates := []*State{
		ith.findState(t, "A"),
	}

	if assert.Equal(t, len(wantRootStates), len(sc.Roots), "Different amount of root states") {
		// Only compare the root states.
		for i := 0; i < len(wantRootStates); i++ {
			want := wantRootStates[i]
			got := sc.Roots[i]

			compareState(t, want, got)
		}
	}
}

// COMPARE FUNCTIONS -------------------------------------------------------------------------------

func compareTrigger(t *testing.T, want, got *Trigger) {
	assert.Equal(t, want.Name, got.Name)
	if assert.Equal(t, len(want.Args), len(got.Args)) {
		for i := 0; i < len(want.Args); i++ {
			want := want.Args[i]
			got := got.Args[i]

			assert.Equal(t, want.Type, got.Type)
			assert.Equal(t, want.Name, got.Name)
		}
	}
}

func compareState(t *testing.T, want, got *State) {
	if !assert.Equal(t, want.Name, got.Name, "Parent: %s", want.Parent.SafeName()) {
		return
	}

	ctxStr := fmt.Sprintf("State: %s", want.Name)

	// Compare concurrentness.
	assert.Equal(t, want.Concurrent, got.Concurrent, ctxStr)
	assert.Equal(t, want.ConcurrentGroup, got.ConcurrentGroup, ctxStr)

	// Compare the enter reactions.
	compareReactions(t, want, want.EnterReactions, got.EnterReactions)

	// Compare the exit reactions.
	compareReactions(t, want, want.ExitReactions, got.ExitReactions)

	// Compare the captures.
	compareCaptures(t, want, want.Captures, got.Captures)

	// Compare that we have the same parent.
	// We only compare name, as otherwise we have a recursive call.
	if assert.Equal(t, want.Parent != nil, got.Parent != nil) && want.Parent != nil {
		assert.Equal(t, want.Parent.Name, got.Parent.Name, ctxStr)
	}

	// Compare that we have the same children.
	if assert.Equal(t, len(want.Children), len(got.Children)) {
		for i := 0; i < len(want.Children); i++ {
			want := want.Children[i]
			got := got.Children[i]

			compareState(t, want, got)
		}
	}

	// Compare the transitions.
	if assert.Equal(t, len(want.Transitions), len(got.Transitions)) {
		for i := 0; i < len(want.Transitions); i++ {
			want := want.Transitions[i]
			got := got.Transitions[i]

			compareTransition(t, want, got)
		}
	}
}

func compareReactions(t *testing.T, state *State, want, got []*StateReaction) {
	if assert.Equal(t, len(want), len(got), state.Name) {
		// Collect the names of the triggers for good error messages.
		var wantList []string
		var gotList []string
		for i := 0; i < len(want); i++ {
			want := want[i]
			got := got[i]

			wantList = append(wantList, want.String())
			gotList = append(gotList, got.String())
		}

		for i := 0; i < len(want); i++ {
			want := want[i]
			got := got[i]

			// Either the triggers are both nil or none.

			if assert.Equal(t, want.Trigger != nil, got.Trigger != nil, "%s:\n- %s\n- %s", state.Name, wantList, gotList) {
				if want.Trigger != nil {
					assert.Equal(t, want.Trigger.Name, got.Trigger.Name, state.Name)
				}
			}
		}
	}
}

func compareCaptures(t *testing.T, state *State, want, got []*StateCapture) {
	if assert.Equal(t, len(want), len(got), state.Name) {
		// Collect the names of the triggers for good error messages.
		var wantList []string
		var gotList []string
		for i := 0; i < len(want); i++ {
			want := want[i]
			got := got[i]

			wantList = append(wantList, want.String())
			gotList = append(gotList, got.String())
		}

		for i := 0; i < len(want); i++ {
			want := want[i]
			got := got[i]

			assert.NotNil(t, got.Trigger, "%s:\n- %s\n- %s", state.Name, wantList, gotList)
			assert.Equal(t, want.Trigger.Name, got.Trigger.Name, "%s:\n- %s\n- %s", state.Name, wantList, gotList)
			assert.Equal(t, want.Condition, got.Condition, "%s:\n- %s\n- %s", state.Name, want.Condition, got.Condition)
		}
	}
}

func compareTransition(t *testing.T, want, got *Transition) {
	// We only compare names to avoid over checking.
	assert.Equal(t, want.From.Name, got.From.Name)
	assert.Equal(t, want.To.Name, got.To.Name)
	if assert.Equal(t, want.Trigger != nil, got.Trigger != nil) {
		if want.Trigger != nil {
			assert.Equal(t, want.Trigger.Name, got.Trigger.Name)
		}
	}
	assert.Equal(t, want.Condition, got.Condition)
	assert.Equal(t, want.Action, got.Action)
}

// HELPER FUNCTIONS --------------------------------------------------------------------------------

func (ith *inputTestHelper) findState(t *testing.T, name string) *State {
	i := slices.IndexFunc(ith.states, func(s *State) bool {
		return s.Name == name
	})
	require.NotEqual(t, i, -1)
	return ith.states[i]
}

func (ith *inputTestHelper) findTrigger(t *testing.T, name string) *Trigger {
	i := slices.IndexFunc(ith.triggers, func(t *Trigger) bool {
		return t.Name == name
	})
	if i == -1 {
		return nil
	}
	return ith.triggers[i]
}

func (ith *inputTestHelper) setParenthood(t *testing.T, parentName string, childNames ...string) {
	parent := ith.findState(t, parentName)

	for _, childName := range childNames {
		state := ith.findState(t, childName)
		state.Parent = parent
		parent.Children = append(parent.Children, state)
	}
}

func (ith *inputTestHelper) setTransition(t *testing.T, fromName, toName, triggerName, condition, action string) {
	from := ith.findState(t, fromName)
	to := ith.findState(t, toName)
	trigger := ith.findTrigger(t, triggerName)

	transition := &Transition{
		From:      from,
		To:        to,
		Trigger:   trigger,
		Condition: condition,
		Action:    action,
	}
	from.Transitions = append(from.Transitions, transition)
}

func (ith *inputTestHelper) setEnterReaction(t *testing.T, stateName, triggerName string) {
	state := ith.findState(t, stateName)

	// In this case, we can have an empty reaction, so we can have a nil trigger.
	var trigger *Trigger
	if triggerName != "" {
		trigger = ith.findTrigger(t, triggerName)
		require.NotNil(t, trigger)
	}

	state.EnterReactions = ith.appendReaction(t, state.EnterReactions, state, trigger, StateReactionKind_Enter)
}

func (ith *inputTestHelper) setExitReaction(t *testing.T, stateName, triggerName string) {
	state := ith.findState(t, stateName)

	// In this case, we can have an empty reaction, so we can have a nil trigger.
	var trigger *Trigger
	if triggerName != "" {
		trigger = ith.findTrigger(t, triggerName)
		require.NotNil(t, trigger)
	}

	state.ExitReactions = ith.appendReaction(t, state.ExitReactions, state, trigger, StateReactionKind_Exit)
}

func (ith *inputTestHelper) setCapture(t *testing.T, stateName, triggerName, condition string) {
	state := ith.findState(t, stateName)
	trigger := ith.findTrigger(t, triggerName)

	state.Captures = ith.appendCapture(t, state.Captures, state, trigger, condition)
}

func (ith *inputTestHelper) appendReaction(t *testing.T, reactions []*StateReaction, state *State, trigger *Trigger, kind StateReactionKind) []*StateReaction {
	reactions = append(reactions, &StateReaction{
		Trigger: trigger,
		Kind:    kind,
	})

	// Validate that the reactions are unique.
	set := map[*Trigger]struct{}{}
	for _, reaction := range reactions {
		if _, ok := set[reaction.Trigger]; ok {
			t.Fatalf("state %s already has a transition for trigger %s", state.SafeName(), trigger.SafeName())
		}
		set[reaction.Trigger] = struct{}{}
	}

	return reactions
}

func (ith *inputTestHelper) appendCapture(t *testing.T, captures []*StateCapture, state *State, trigger *Trigger, condition string) []*StateCapture {
	captures = append(captures, &StateCapture{
		Trigger:   trigger,
		Condition: condition,
	})

	// Validate that the captures are unique.
	set := map[*Trigger]struct{}{}
	for _, capture := range captures {
		if _, ok := set[capture.Trigger]; ok {
			t.Fatalf("State %q already has a capture for trigger %q", state.SafeName(), trigger.SafeName())
		}
		set[capture.Trigger] = struct{}{}
	}

	return captures
}
