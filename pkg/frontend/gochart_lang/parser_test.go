package gochart_lang

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSimple(t *testing.T) {
	data, err := os.ReadFile("testdata/example.gochart")
	require.NoError(t, err)

	scanner := NewScanner()
	tokens, errors := scanner.gatherTokens(bytes.NewReader(data))
	require.Empty(t, errors)

	parser := NewParser()
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
						Transitions: []*ASTNodeStateTransition{
							{
								Trigger:     "",
								TargetState: "Third",
							},
							{
								Trigger:     "Execute",
								TargetState: "Second",
							},
						},
					},
					{
						Name: "Second",
						Options: []*ASTNodeStateOption{
							{
								Name:      "Option1",
								ValueType: ValueType_None,
							},
						},
						Transitions: []*ASTNodeStateTransition{
							{
								Trigger:     "Execute",
								TargetState: "First",
							},
						},
					},
					{
						Name: "Third",
						Options: []*ASTNodeStateOption{
							{
								Name:      "Opt1",
								ValueType: ValueType_None,
							},
							{
								Name:      "Opt2",
								ValueType: ValueType_None,
							},
							{
								Name:      "Opt3",
								ValueType: ValueType_String,
								Value:     "Identifier",
							},
							{
								Name:      "Opt4",
								ValueType: ValueType_String,
								Value:     "String",
							},
							{
								Name:      "Opt5",
								ValueType: ValueType_Integer,
								Value:     "2123",
							},
							{
								Name:      "Opt6",
								ValueType: ValueType_Boolean,
								Value:     "true",
							},
							{
								Name:      "Opt7",
								ValueType: ValueType_Boolean,
								Value:     "false",
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
								Name: "ThirdChild1",
								Options: []*ASTNodeStateOption{
									{
										Name: "OptChild1",
									},
									{
										Name:      "OptChild2",
										ValueType: ValueType_String,
										Value:     "IdentifierChild",
									},
								},
								Children: []*ASTNodeState{
									{
										Name: "ThirdGrandChild",
										Options: []*ASTNodeStateOption{
											{
												Name:      "OptGrandChild",
												ValueType: ValueType_Boolean,
												Value:     "false",
											},
										},
										Transitions: []*ASTNodeStateTransition{
											{
												Trigger:     "Execute",
												TargetState: "Second",
											},
										},
									},
								},
							},
							{
								Name: "ThirdChild2",
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
			assert.Equal(t, want.ValueType, got.ValueType, "%s, Option %d", ctxStr, i)
			assert.Equal(t, want.Value, got.Value, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare transitions.
	if assert.Equal(t, len(want.Transitions), len(got.Transitions), ctxStr) {
		for i := 0; i < len(want.Transitions); i++ {
			want := want.Transitions[i]
			got := got.Transitions[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Transition %d", ctxStr, i)
			assert.Equal(t, want.TargetState, got.TargetState, "%s, Transition %d", ctxStr, i)
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
}
