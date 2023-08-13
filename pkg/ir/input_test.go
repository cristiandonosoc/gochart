package ir

import (
	"slices"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/frontend"
	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	// StateA
	setEnterReaction(t, wantTriggers, wantStates, "StateA", "Trigger1")
	setEnterReaction(t, wantTriggers, wantStates, "StateA", "")
	setExitReaction(t, wantTriggers, wantStates, "StateA", "Trigger1")
	setExitReaction(t, wantTriggers, wantStates, "StateA", "")
	setParenthood(t, wantStates, "StateA", "StateB")
	setParenthood(t, wantStates, "StateA", "StateC")
	// StateB
	setTransition(t, wantTriggers, wantStates, "StateB", "StateC", "Trigger1")
	setExitReaction(t, wantTriggers, wantStates, "StateB", "Trigger1")
	// StateC
	setEnterReaction(t, wantTriggers, wantStates, "StateC", "Trigger1")
	setEnterReaction(t, wantTriggers, wantStates, "StateC", "")

	yf := yaml.NewYamlFrontend()

	path, err := test_support.RunfilePath("testdata/simple.yaml")
	require.NoError(t, err)

	scdata, err := frontend.ProcessFromFile(yf, path)
	require.NoError(t, err)
	require.Equal(t, len(wantTriggers), len(scdata.Triggers))

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

func compareTransition(t *testing.T, want, got *Transition) {
	// We only compare names to avoid over checking.
	assert.Equal(t, want.From.Name, got.From.Name)
	assert.Equal(t, want.To.Name, got.To.Name)
	assert.Equal(t, want.Trigger.Name, got.Trigger.Name)
}

// HELPER FUNCTIONS --------------------------------------------------------------------------------

func findState(t *testing.T, states []*State, name string) *State {
	i := slices.IndexFunc(states, func(s *State) bool {
		return s.Name == name
	})
	require.NotEqual(t, i, -1)
	return states[i]
}

func findTrigger(t *testing.T, triggers []*Trigger, name string) *Trigger {
	i := slices.IndexFunc(triggers, func(t *Trigger) bool {
		return t.Name == name
	})
	require.NotEqual(t, i, -1)
	return triggers[i]
}

func setParenthood(t *testing.T, states []*State, parentName, childName string) {
	state := findState(t, states, childName)
	parent := findState(t, states, parentName)

	state.Parent = parent
	parent.Children = append(parent.Children, state)
}

func setTransition(t *testing.T, triggers []*Trigger, states []*State, fromName, toName, triggerName string) {
	from := findState(t, states, fromName)
	to := findState(t, states, toName)
	trigger := findTrigger(t, triggers, triggerName)

	transition := &Transition{
		From:    from,
		To:      to,
		Trigger: trigger,
	}
	from.Transitions = append(from.Transitions, transition)
}

func appendReaction(t *testing.T, reactions []*StateReaction, state *State, trigger *Trigger) []*StateReaction {
	reactions = append(reactions, &StateReaction{
		Trigger: trigger,
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

func setEnterReaction(t *testing.T, triggers []*Trigger, states []*State, stateName, triggerName string) {
	state := findState(t, states, stateName)

	// In this case, we can have an empty reaction, so we can have a nil trigger.
	var trigger *Trigger
	if triggerName != "" {
		trigger = findTrigger(t, triggers, triggerName)
	}

	state.EnterReactions = appendReaction(t, state.EnterReactions, state, trigger)
}

func setExitReaction(t *testing.T, triggers []*Trigger, states []*State, stateName, triggerName string) {
	state := findState(t, states, stateName)

	// In this case, we can have an empty reaction, so we can have a nil trigger.
	var trigger *Trigger
	if triggerName != "" {
		trigger = findTrigger(t, triggers, triggerName)
	}

	state.ExitReactions = appendReaction(t, state.ExitReactions, state, trigger)
}
