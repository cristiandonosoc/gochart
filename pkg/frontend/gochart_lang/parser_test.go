package gochart_lang

import (
	"fmt"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSimple(t *testing.T) {
	lf, err := test_support.LoadRunfile("testdata/example.gochart")
	require.NoError(t, err)

	scanner := NewScanner()
	tokens, errors := scanner.gatherTokens(lf)
	require.Empty(t, errors)

	parser := NewParser(lf, &Options{})
	got, errors := parser.Parse(tokens)
	require.Empty(t, errors)

	want := &ASTNodeRoot{
		statecharts: []*ASTNodeStatechart{
			{
				Name: "Example",
				Triggers: []*ASTNodeTrigger{
					{
						Name: "Execute",
					},
					{
						Name:       "Update",
						ArgsString: "float dt",
					},
					{
						Name:       "SomeMessage",
						ArgsString: "const hvn::Message* message, int other, const std::string& another",
					},
				},
				States: []*ASTNodeState{
					{
						Name: "First",
						EnterReactions: []*ASTNodeStateReaction{
							{
								Trigger: "",
							},
							{
								Trigger: "Execute",
							},
						},
						ExitReactions: []*ASTNodeStateReaction{
							{
								Trigger: "Execute",
							},
							{
								Trigger: "",
							},
						},
						Transitions: []*ASTNodeStateTransition{
							{
								Trigger:     "",
								TargetState: "Third",
								Action:      " owner->SomeNullAction(); ",
							},
							{
								Trigger:     "Execute",
								TargetState: "Second",
							},
						},
						Captures: []*ASTNodeStateCapture{
							{
								Trigger:   "Update",
								Condition: " owner->SomeCondition ",
							},
						},
					},
					{
						Name: "Second",
						Options: []*ASTNodeOption{
							{
								Name:      "Option1",
								ValueKind: output.OptionValueKind_Boolean,
								Value:     "true",
							},
						},
						Transitions: []*ASTNodeStateTransition{
							{
								Trigger:     "Execute",
								TargetState: "First",
								Condition:   "\r\n\t\t\treturn owner->SomeCondition();\r\n\t\t",
								Action:      "\r\n\t\t\towner->ExecuteSomeAction();\r\n\t\t",
							},
						},
					},
					{
						Name: "Third",
						Options: []*ASTNodeOption{
							{
								Name:      "Opt1",
								ValueKind: output.OptionValueKind_Boolean,
								Value:     "true",
							},
							{
								Name:      "Opt2",
								ValueKind: output.OptionValueKind_Boolean,
								Value:     "true",
							},
							{
								Name:      "Opt3",
								ValueKind: output.OptionValueKind_String,
								Value:     "Identifier",
							},
							{
								Name:      "Opt4",
								ValueKind: output.OptionValueKind_String,
								Value:     "String",
							},
							{
								Name:      "Opt5",
								ValueKind: output.OptionValueKind_Integer,
								Value:     "2123",
							},
							{
								Name:      "Opt6",
								ValueKind: output.OptionValueKind_Boolean,
								Value:     "true",
							},
							{
								Name:      "Opt7",
								ValueKind: output.OptionValueKind_Boolean,
								Value:     "false",
							},
							{
								Name:      "Opt8",
								ValueKind: output.OptionValueKind_Integer,
								Value:     "22",
							},
						},
						Transitions: []*ASTNodeStateTransition{
							{
								Trigger:     "SomeMessage",
								TargetState: "Second",
							},
						},
						Children: []*ASTNodeState{
							{
								Name:       "ThirdChild1",
								ParentName: "Third",
								Options: []*ASTNodeOption{
									{
										Name:      "OptChild1",
										ValueKind: output.OptionValueKind_Boolean,
										Value:     "true",
									},
									{
										Name:      "OptChild2",
										ValueKind: output.OptionValueKind_String,
										Value:     "IdentifierChild",
									},
								},
								Children: []*ASTNodeState{
									{
										Name:       "ThirdGrandChild",
										ParentName: "ThirdChild1",
										Options: []*ASTNodeOption{
											{
												Name:      "OptGrandChild",
												ValueKind: output.OptionValueKind_Boolean,
												Value:     "false",
											},
										},
										EnterReactions: []*ASTNodeStateReaction{
											{
												Trigger: "",
											},
										},
										ExitReactions: []*ASTNodeStateReaction{
											{
												Trigger: "Update",
											},
										},
										Transitions: []*ASTNodeStateTransition{
											{
												Trigger:     "Execute",
												TargetState: "Second",
											},
										},
										Captures: []*ASTNodeStateCapture{
											{
												Trigger: "Execute",
											},
										},
									},
								},
							},
							{
								Name:       "ThirdChild2",
								ParentName: "Third",
								Transitions: []*ASTNodeStateTransition{
									{
										Trigger:     "",
										TargetState: "ThirdGrandChild",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compareASTNodeRoot(t, want, got)
}

// Helpers -----------------------------------------------------------------------------------------

func compareASTNodeRoot(t *testing.T, want, got *ASTNodeRoot) {
	if assert.Equal(t, len(want.statecharts), len(got.statecharts)) {
		for i := 0; i < len(want.statecharts); i++ {
			want := want.statecharts[i]
			got := got.statecharts[i]

			compareASTNodeStatechart(t, want, got)
		}
	}
}

func compareASTNodeStatechart(t *testing.T, want, got *ASTNodeStatechart) {
	assert.Equal(t, want.Name, got.Name)

	if assert.Equal(t, len(want.Triggers), len(got.Triggers), "Not same amount of triggers") {
		for i := 0; i < len(want.Triggers); i++ {
			want := want.Triggers[i]
			got := got.Triggers[i]

			tc := &triggerComparer{
				index:          i,
				statechartName: want.Name,
			}
			tc.compare(t, want, got)
		}
	}

	statechartName := want.Name
	if assert.Equal(t, len(want.States), len(got.States), "Not same amount of states") {
		for i := 0; i < len(want.States); i++ {
			want := want.States[i]
			got := got.States[i]

			sc := &stateComparer{
				index:          i,
				statechartName: statechartName,
			}
			sc.compare(t, want, got)
		}
	}
}

type triggerComparer struct {
	index          int
	statechartName string
}

func (tc *triggerComparer) compare(t *testing.T, want, got *ASTNodeTrigger) {
	assert.Equal(t, want.Name, got.Name, "Statechart %q, Trigger %d", tc.statechartName, tc.index)

	ctxStr := fmt.Sprintf("Statechart %q, Trigger %d (%s)", tc.statechartName, tc.index, want.Name)

	assert.Equal(t, want.ArgsString, got.ArgsString, ctxStr)
}

type stateComparer struct {
	index          int
	statechartName string
	parentState    string
}

func (sc *stateComparer) compare(t *testing.T, want, got *ASTNodeState) {
	assert.Equal(t, want.Name, got.Name, "Statechart %s, State %d", sc.statechartName, sc.index)

	ctxStr := fmt.Sprintf("Statechart %s, State %q, Parent %q, Trigger %d (%s)", sc.statechartName, want.Name, sc.parentState, sc.index, want.Name)

	// Compare options.
	if assert.Equal(t, len(want.Options), len(got.Options), ctxStr) {
		for i := 0; i < len(want.Options); i++ {
			want := want.Options[i]
			got := got.Options[i]

			assert.Equal(t, want.Name, got.Name, "%s, Option %d", ctxStr, i)
			assert.Equal(t, want.ValueKind, got.ValueKind, "%s, Option %d", ctxStr, i)
			assert.Equal(t, want.Value, got.Value, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare enter reactions.
	if assert.Equal(t, len(want.EnterReactions), len(got.EnterReactions), ctxStr) {
		for i := 0; i < len(want.EnterReactions); i++ {
			want := want.EnterReactions[i]
			got := got.EnterReactions[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare exit reactions.
	if assert.Equal(t, len(want.ExitReactions), len(got.ExitReactions), ctxStr) {
		for i := 0; i < len(want.ExitReactions); i++ {
			want := want.ExitReactions[i]
			got := got.ExitReactions[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare transitions.
	if assert.Equal(t, len(want.Transitions), len(got.Transitions), ctxStr) {
		for i := 0; i < len(want.Transitions); i++ {
			want := want.Transitions[i]
			got := got.Transitions[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Transition %d", ctxStr, i)
			assert.Equal(t, want.TargetState, got.TargetState, "%s, Transition %d", ctxStr, i)
			assert.Equal(t, want.Condition, got.Condition, "%s, Transition %d", ctxStr, i)
			assert.Equal(t, want.Action, got.Action, "%s, Transition %d", ctxStr, i)
		}
	}

	// Compare captures.
	if assert.Equal(t, len(want.Captures), len(got.Captures), ctxStr) {
		for i := 0; i < len(want.Captures); i++ {
			want := want.Captures[i]
			got := got.Captures[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Capture %d", ctxStr, i)
			assert.Equal(t, want.Condition, got.Condition, "%s, Capture %d", ctxStr, i)
		}
	}

	// Compare children.
	parentStateName := want.Name
	if assert.Equal(t, len(want.Children), len(got.Children), ctxStr) {
		for i := 0; i < len(want.Children); i++ {
			want := want.Children[i]
			got := got.Children[i]

			subComparer := &stateComparer{
				index:          i,
				statechartName: sc.statechartName,
				parentState:    parentStateName,
			}
			subComparer.compare(t, want, got)
		}
	}

	// Compare parent.
	assert.Equal(t, want.ParentName, got.ParentName, ctxStr)
}
