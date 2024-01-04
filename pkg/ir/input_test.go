package ir

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/test_support"
	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
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
			Name: frontend_common.RootStateName,
		},
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
	ith.setParenthood(t, frontend_common.RootStateName, "StateA")
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
	require.Equal(t, len(wantStates), len(scdata.States),
		"WANT:", printStates(wantStates), "GOT:", printFrontendStates(scdata.States))

	t.Log(printFrontendStates(scdata.States))

	gc, err := ProcessGochartData(scdata)
	require.NoError(t, err)

	assert.Equal(t, "Simple", gc.Name)

	if assert.Equal(t, len(wantTriggers), len(gc.Triggers), "Different amount of triggers") {
		for i := 0; i < len(wantTriggers); i++ {
			want := wantTriggers[i]
			got := gc.Triggers[i]

			CompareTrigger(t, want, got)
		}
	}

	want := ith.findState(t, frontend_common.RootStateName)
	CompareState(t, want, gc.Root)
}

func TestConcurrentGroups(t *testing.T) {
	wantTriggers := []*Trigger{
		{
			Name: "SomeTrigger",
		},
	}

	wantStates := []*State{
		{
			Name:            frontend_common.RootStateName,
			ConcurrentGroup: 0,
		},
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

	ith.setParenthood(t, frontend_common.RootStateName, "A")
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

	gc, err := ProcessGochartData(scdata)
	require.NoError(t, err)

	assert.Equal(t, "Concurrent", gc.Name)

	if assert.Equal(t, len(wantTriggers), len(gc.Triggers), "Different amount of triggers") {
		for i := 0; i < len(wantTriggers); i++ {
			want := wantTriggers[i]
			got := gc.Triggers[i]

			CompareTrigger(t, want, got)
		}
	}

	want := ith.findState(t, frontend_common.RootStateName)
	CompareState(t, want, gc.Root)
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

func printStates(states []*State) string {
	var sb strings.Builder

	sb.WriteString("[\n")
	for _, state := range states {
		sb.WriteString(fmt.Sprintf("\t%s\n", state.Name))
	}
	sb.WriteString("]")

	return sb.String()
}

func printFrontendStates(states []*frontend_common.StateData) string {
	var sb strings.Builder

	sb.WriteString("[\n")
	for _, state := range states {
		sb.WriteString(fmt.Sprintf("\t%s, Parent: %s\n", state.Name, state.Parent))
	}
	sb.WriteString("]")

	return sb.String()
}
