package ir

import (
	"testing"

	"github.com/cristiandonosoc/gochart/pkg/frontend/yaml"

	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSimpleYaml(t *testing.T) {
	wantTriggers := []*Trigger{
		{
			Name: "Trigger1",
			Args: []*TriggerArgument{
				{
					Type: "const std::string&",
					Name: "foo",
				},
			},
		},
	}

	wantStates := []*State{
		{
			Name: "StateA",
		},
		{
			Name: "StateB",
		},
		{
			Name: "StateC",
		},
	}
	setParenthood(t, wantStates, "StateA", "StateB")
	setParenthood(t, wantStates, "StateA", "StateC")
	setTransition(t, wantTriggers, wantStates, "StateB", "StateC", "Trigger1")

	yf := yaml.NewYamlFrontend()

	scdata, err := yf.ProcessFromFile("testdata/simple.yaml")
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

func findState(t *testing.T, states []*State, name string) *State {
	res := xslices.Filter(states, func(s *State) bool {
		return s.Name == name
	})
	require.Equal(t, 1, len(res))
	return res[0]
}

func findTrigger(t *testing.T, triggers []*Trigger, name string) *Trigger {
	res := xslices.Filter(triggers, func(trigger *Trigger) bool {
		return trigger.Name == name
	})
	require.Equal(t, 1, len(res))
	return res[0]
}

func setParenthood(t *testing.T, states []*State, parentName, childName string) {
	state := findState(t, states, childName)
	parent := findState(t, states, parentName)

	state.Parent = parent
	parent.States = append(parent.States, state)
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

	// Compare that we have the same parent.
	// We only compare name, as otherwise we have a recursive call.
	if assert.Equal(t, want.Parent != nil, got.Parent != nil) && want.Parent != nil {
		assert.Equal(t, want.Parent.Name, got.Parent.Name)
	}

	// Compare that we have the same children.
	if assert.Equal(t, len(want.States), len(got.States)) {
		for i := 0; i < len(want.States); i++ {
			want := want.States[i]
			got := got.States[i]

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

func compareTransition(t *testing.T, want, got *Transition) {
	// We only compare names to avoid over checking.
	assert.Equal(t, want.From.Name, got.From.Name)
	assert.Equal(t, want.To.Name, got.To.Name)
	assert.Equal(t, want.Trigger.Name, got.Trigger.Name)
}
