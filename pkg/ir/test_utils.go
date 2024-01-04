package ir

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// COMPARE FUNCTIONS -------------------------------------------------------------------------------

func CompareGochart(t *testing.T, want, got *Gochart) {
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.Namespace, got.Namespace)

	if assert.Equal(t, len(want.Options), len(got.Options)) {
		for i := 0; i < len(want.Options); i++ {
			want := want.Options[i]
			got := got.Options[i]

			CompareOption(t, want, got)
		}
	}

	if assert.Equal(t, len(want.Triggers), len(got.Triggers)) {
		for i := 0; i < len(want.Triggers); i++ {
			want := want.Triggers[i]
			got := got.Triggers[i]

			CompareTrigger(t, want, got)
		}
	}

	CompareState(t, want.Root, got.Root)
}

func CompareOption(t *testing.T, want, got *Option) {
	assert.Equal(t, want.Definition.Name, got.Definition.Name)
	assert.Equal(t, want.Definition.Kind, got.Definition.Kind)
	assert.Equal(t, want.stringValue, got.stringValue)
	assert.Equal(t, want.intValue, got.intValue)
	assert.Equal(t, want.boolValue, got.boolValue)
	assert.Equal(t, want.implicitBool, got.implicitBool)
}

func CompareTrigger(t *testing.T, want, got *Trigger) {
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

func CompareState(t *testing.T, want, got *State) {
	if !assert.Equal(t, want.Name, got.Name, "Parent: %s", want.Parent.SafeName()) {
		return
	}

	ctxStr := fmt.Sprintf("State: %s", want.Name)

	// Compare concurrentness.
	assert.Equal(t, want.Concurrent, got.Concurrent, ctxStr)
	assert.Equal(t, want.ConcurrentGroup, got.ConcurrentGroup, ctxStr)

	// Compare the enter reactions.
	CompareReactions(t, want, want.EnterReactions, got.EnterReactions)

	// Compare the exit reactions.
	CompareReactions(t, want, want.ExitReactions, got.ExitReactions)

	// Compare the captures.
	CompareCaptures(t, want, want.Captures, got.Captures)

	// Compare that we have the same parent.
	// We only Compare name, as otherwise we have a recursive call.
	if assert.Equal(t, want.Parent != nil, got.Parent != nil) && want.Parent != nil {
		assert.Equal(t, want.Parent.Name, got.Parent.Name, ctxStr)
	}

	// Compare that we have the same children.
	if assert.Equal(t, len(want.Children), len(got.Children), want.SafeName()) {
		for i := 0; i < len(want.Children); i++ {
			want := want.Children[i]
			got := got.Children[i]

			CompareState(t, want, got)
		}
	}

	// Compare the transitions.
	if assert.Equal(t, len(want.Transitions), len(got.Transitions)) {
		for i := 0; i < len(want.Transitions); i++ {
			want := want.Transitions[i]
			got := got.Transitions[i]

			CompareTransition(t, want, got)
		}
	}
}

func CompareReactions(t *testing.T, state *State, want, got []*StateReaction) {
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

func CompareCaptures(t *testing.T, state *State, want, got []*StateCapture) {
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

func CompareTransition(t *testing.T, want, got *Transition) {
	// We only compare names to avoid over checking.
	assert.Equal(t, want.From.Name, got.From.Name)
	assert.Equal(t, want.To.Name, got.To.Name)
	if assert.Equal(t, want.Trigger != nil, got.Trigger != nil) {
		if want.Trigger != nil {
			assert.Equal(t, want.Trigger.Name, got.Trigger.Name)
		}
	}
	assert.Equal(t, strings.TrimSpace(want.Condition), strings.TrimSpace(got.Condition))
	assert.Equal(t, strings.TrimSpace(want.Action), strings.TrimSpace(got.Action))
}
