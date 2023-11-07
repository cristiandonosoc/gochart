package ir

import (
	"bytes"
	"os"
	"slices"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
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
	ith.setCapture(t, "StateA", "Trigger1")
	ith.setCapture(t, "StateA", "Trigger2")
	ith.setParenthood(t, "StateA", "StateB")
	ith.setParenthood(t, "StateA", "StateC")
	// StateB
	ith.setTransition(t, "StateB", "StateC", "Trigger1")
	ith.setExitReaction(t, "StateB", "Trigger1")
	// StateC
	ith.setEnterReaction(t, "StateC", "Trigger1")
	ith.setEnterReaction(t, "StateC", "")
	ith.setCapture(t, "StateC", "Trigger1")

	yf, err := yaml.NewYamlFrontend()
	require.NoError(t, err)

	path, err := test_support.RunfilePath("testdata/simple.yaml")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	scdata, err := yf.Process(bytes.NewReader(data))
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

	if assert.Equal(t, len(wantStates), len(sc.States), "Different amount of states") {
		for i := 0; i < len(wantStates); i++ {
			want := wantStates[i]
			got := sc.States[i]

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
	assert.Equal(t, want.Name, got.Name)

	// Compare the enter reactions.
	compareReactions(t, want, want.EnterReactions, got.EnterReactions)

	// Compare the exit reactions.
	compareReactions(t, want, want.ExitReactions, got.ExitReactions)

	// Compare the captures.
	compareCaptures(t, want, want.Captures, got.Captures)

	// Compare that we have the same parent.
	// We only compare name, as otherwise we have a recursive call.
	if assert.Equal(t, want.Parent != nil, got.Parent != nil) && want.Parent != nil {
		assert.Equal(t, want.Parent.Name, got.Parent.Name)
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
		}
	}
}

func compareTransition(t *testing.T, want, got *Transition) {
	// We only compare names to avoid over checking.
	assert.Equal(t, want.From.Name, got.From.Name)
	assert.Equal(t, want.To.Name, got.To.Name)
	assert.Equal(t, want.Trigger.Name, got.Trigger.Name)
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
	require.NotEqual(t, i, -1)
	return ith.triggers[i]
}

func (ith *inputTestHelper) setParenthood(t *testing.T, parentName, childName string) {
	state := ith.findState(t, childName)
	parent := ith.findState(t, parentName)

	state.Parent = parent
	parent.Children = append(parent.Children, state)
}

func (ith *inputTestHelper) setTransition(t *testing.T, fromName, toName, triggerName string) {
	from := ith.findState(t, fromName)
	to := ith.findState(t, toName)
	trigger := ith.findTrigger(t, triggerName)

	transition := &Transition{
		From:    from,
		To:      to,
		Trigger: trigger,
	}
	from.Transitions = append(from.Transitions, transition)
}

func (ith *inputTestHelper) setEnterReaction(t *testing.T, stateName, triggerName string) {
	state := ith.findState(t, stateName)

	// In this case, we can have an empty reaction, so we can have a nil trigger.
	var trigger *Trigger
	if triggerName != "" {
		trigger = ith.findTrigger(t, triggerName)
	}

	state.EnterReactions = ith.appendReaction(t, state.EnterReactions, state, trigger, StateReactionKind_Enter)
}

func (ith *inputTestHelper) setExitReaction(t *testing.T, stateName, triggerName string) {
	state := ith.findState(t, stateName)

	// In this case, we can have an empty reaction, so we can have a nil trigger.
	var trigger *Trigger
	if triggerName != "" {
		trigger = ith.findTrigger(t, triggerName)
	}

	state.ExitReactions = ith.appendReaction(t, state.ExitReactions, state, trigger, StateReactionKind_Exit)
}

func (ith *inputTestHelper) setCapture(t *testing.T, stateName, triggerName string) {
	state := ith.findState(t, stateName)
	trigger := ith.findTrigger(t, triggerName)

	state.Captures = ith.appendCapture(t, state.Captures, state, trigger)
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

func (ith *inputTestHelper) appendCapture(t *testing.T, captures []*StateCapture, state *State, trigger *Trigger) []*StateCapture {
	captures = append(captures, &StateCapture{
		Trigger: trigger,
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
