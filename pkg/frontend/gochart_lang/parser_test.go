package gochart_lang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cristiandonosoc/gochart/internal/test_support"
	"github.com/cristiandonosoc/gochart/pkg/frontend/common"

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
		gocharts: []*ASTNodeGochart{
			{
				Name: "Example",
				Options: []*ASTNodeOption{
					{
						Name:      "Opt1",
						ValueKind: common.OptionValueKind_Boolean,
						Value:     "true",
					},
					{
						Name:      "Opt2",
						ValueKind: common.OptionValueKind_String,
						Value:     "String",
					},
					{
						Name:      "Opt3",
						ValueKind: common.OptionValueKind_Integer,
						Value:     "33",
					},
				},
				Comments: []*ASTNodeComment{
					{
						comment:    "Outside Comment 1",
						entryIndex: 0,
					},
					{
						comment:    "Outside Comment 2",
						entryIndex: 1,
					},
					{
						comment:    "Outside Comment 3",
						entryIndex: 3,
					},
					{
						comment:    "Outside Comment 4",
						entryIndex: 4,
					},
				},
				Triggers: []*ASTNodeTrigger{
					{
						Name:       "Execute",
						entryIndex: 1,
					},
					{
						Name:       "Update",
						ArgsString: "float dt",
						entryIndex: 3,
					},
					{
						Name:       "SomeMessage",
						ArgsString: "const hvn::Message* message, int other, const std::string& another",
						entryIndex: 4,
					},
				},
				Root: &ASTNodeState{
					Name:       common.RootStateName,
					entryIndex: 2,
					Comments: []*ASTNodeComment{
						{
							comment:    "Inside Comment 1",
							entryIndex: 0,
						},
						{
							comment:    "Inside Comment 2",
							entryIndex: 2,
						},
						{
							comment:    "Inside Comment 3",
							entryIndex: 9,
						},
						{
							comment:    "Inside Comment 4",
							entryIndex: 11,
						},
					},
					EnterReactions: []*ASTNodeStateReaction{
						{
							Trigger:    "",
							entryIndex: 5,
						},
						{
							Trigger:    "Execute",
							entryIndex: 6,
						},
					},
					ExitReactions: []*ASTNodeStateReaction{
						{
							Trigger:    "Execute",
							entryIndex: 7,
						},
						{
							Trigger:    "",
							entryIndex: 8,
						},
					},
					Transitions: []*ASTNodeStateTransition{
						{
							Trigger:     "Execute",
							TargetState: "First",
							Action:      " owner->Foo(); ",
							entryIndex:  10,
						},
					},
					Children: []*ASTNodeState{
						{
							Name:       "First",
							ParentName: common.RootStateName,
							entryIndex: 12,
							Comments: []*ASTNodeComment{
								{
									comment:    "First Comment 1",
									entryIndex: 2,
								},
							},
							EnterReactions: []*ASTNodeStateReaction{
								{
									Trigger:    "",
									entryIndex: 0,
								},
								{
									Trigger:    "Execute",
									entryIndex: 1,
								},
							},
							ExitReactions: []*ASTNodeStateReaction{
								{
									Trigger:    "Execute",
									entryIndex: 3,
								},
								{
									Trigger:    "",
									entryIndex: 4,
								},
							},
							Transitions: []*ASTNodeStateTransition{
								{
									Trigger:     "",
									TargetState: "Third",
									Action:      " owner->SomeNullAction(); ",
									entryIndex:  5,
								},
								{
									Trigger:     "Execute",
									TargetState: "Second",
									entryIndex:  6,
								},
							},
							Captures: []*ASTNodeStateCapture{
								{
									Trigger:    "Update",
									Condition:  " owner->SomeCondition ",
									entryIndex: 7,
								},
							},
						},
						{
							Name:       "Second",
							ParentName: common.RootStateName,
							entryIndex: 13,
							Options: []*ASTNodeOption{
								{
									Name:      "Option1",
									ValueKind: common.OptionValueKind_Boolean,
									Value:     "",
								},
							},
							Transitions: []*ASTNodeStateTransition{
								{
									Trigger:     "Execute",
									TargetState: "First",
									Condition:   "\r\n\t\t\treturn owner->SomeCondition();\r\n\t\t",
									Action:      "\r\n\t\t\towner->ExecuteSomeAction();\r\n\t\t",
									entryIndex:  0,
								},
							},
						},
						{
							Name:       "Third",
							ParentName: common.RootStateName,
							entryIndex: 14,
							Options: []*ASTNodeOption{
								{
									Name:      "Opt1",
									ValueKind: common.OptionValueKind_Boolean,
									Value:     "",
								},
								{
									Name:      "Opt2",
									ValueKind: common.OptionValueKind_Boolean,
									Value:     "",
								},
								{
									Name:      "Opt3",
									ValueKind: common.OptionValueKind_String,
									Value:     "Identifier",
								},
								{
									Name:      "Opt4",
									ValueKind: common.OptionValueKind_String,
									Value:     "String",
								},
								{
									Name:      "Opt5",
									ValueKind: common.OptionValueKind_Integer,
									Value:     "2123",
								},
								{
									Name:      "Opt6",
									ValueKind: common.OptionValueKind_Boolean,
									Value:     "true",
								},
								{
									Name:      "Opt7",
									ValueKind: common.OptionValueKind_Boolean,
									Value:     "false",
								},
								{
									Name:      "Opt8",
									ValueKind: common.OptionValueKind_Integer,
									Value:     "22",
								},
							},
							Transitions: []*ASTNodeStateTransition{
								{
									Trigger:     "SomeMessage",
									TargetState: "Second",
									entryIndex:  0,
								},
							},
							Children: []*ASTNodeState{
								{
									Name:       "ThirdChild1",
									entryIndex: 1,
									ParentName: "Third",
									Options: []*ASTNodeOption{
										{
											Name:      "OptChild1",
											ValueKind: common.OptionValueKind_Boolean,
											Value:     "",
										},
										{
											Name:      "OptChild2",
											ValueKind: common.OptionValueKind_String,
											Value:     "IdentifierChild",
										},
									},
									Comments: []*ASTNodeComment{
										{
											comment:    "ThirdChild1 Comment 1",
											entryIndex: 1,
										},
									},
									Children: []*ASTNodeState{
										{
											Name:       "ThirdGrandChild",
											entryIndex: 0,
											ParentName: "ThirdChild1",
											Options: []*ASTNodeOption{
												{
													Name:      "OptGrandChild",
													ValueKind: common.OptionValueKind_Boolean,
													Value:     "false",
												},
											},
											Comments: []*ASTNodeComment{
												{
													comment:    "ThirdGrandChild Comment 1",
													entryIndex: 1,
												},
											},
											EnterReactions: []*ASTNodeStateReaction{
												{
													Trigger:    "",
													entryIndex: 2,
												},
											},
											ExitReactions: []*ASTNodeStateReaction{
												{
													Trigger:    "Update",
													entryIndex: 3,
												},
											},
											Transitions: []*ASTNodeStateTransition{
												{
													Trigger:     "Execute",
													TargetState: "Second",
													entryIndex:  0,
												},
											},
											Captures: []*ASTNodeStateCapture{
												{
													Trigger:    "Execute",
													entryIndex: 4,
												},
											},
										},
									},
								},
								{
									Name:       "ThirdChild2",
									entryIndex: 2,
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
		},
	}

	compareASTNodeRoot(t, want, got)
}

// Helpers -----------------------------------------------------------------------------------------

func compareASTNodeRoot(t *testing.T, want, got *ASTNodeRoot) {
	if assert.Equal(t, len(want.gocharts), len(got.gocharts)) {
		for i := 0; i < len(want.gocharts); i++ {
			want := want.gocharts[i]
			got := got.gocharts[i]

			compareASTNodeGochart(t, want, got)
		}
	}
}

func compareASTNodeGochart(t *testing.T, want, got *ASTNodeGochart) {
	assert.Equal(t, want.Name, got.Name)
	gcName := want.Name

	if assert.Equal(t, len(want.Triggers), len(got.Triggers), "Not same amount of triggers") {
		for i := 0; i < len(want.Triggers); i++ {
			want := want.Triggers[i]
			got := got.Triggers[i]

			tc := &triggerComparer{
				index:       i,
				gochartName: want.Name,
			}
			tc.compare(t, want, got)
		}

		parent := fmt.Sprintf("gochart:%s", gcName)
		if assert.Equal(t, len(want.Comments), len(got.Comments), parent) {
			for i := 0; i < len(want.Comments); i++ {
				want := want.Comments[i]
				got := got.Comments[i]

				cc := &commentComparer{
					parent: parent,
				}
				cc.compare(t, want, got)
			}
		}
	}

	gochartName := want.Name

	sc := &stateComparer{
		index:       0,
		gochartName: gochartName,
	}
	sc.compare(t, want.Root, got.Root)
}

type triggerComparer struct {
	index       int
	gochartName string
}

func (tc *triggerComparer) compare(t *testing.T, want, got *ASTNodeTrigger) {
	assert.Equal(t, want.Name, got.Name, "Gochart %q, Trigger %d", tc.gochartName, tc.index)
	assert.Equal(t, want.entryIndex, got.entryIndex, tc.gochartName, tc.index, want.Name)

	ctxStr := fmt.Sprintf("Gochart %q, Trigger %d (%s)", tc.gochartName, tc.index, want.Name)

	assert.Equal(t, want.ArgsString, got.ArgsString, ctxStr)
}

type stateComparer struct {
	index       int
	gochartName string
	parentState string
}

func (sc *stateComparer) compare(t *testing.T, want, got *ASTNodeState) {
	assert.Equal(t, want.Name, got.Name, "Gochart %s, State %d", sc.gochartName, sc.index)
	stateName := want.Name

	assert.Equal(t, want.entryIndex, got.entryIndex, sc.gochartName, want.Name, sc.parentState, sc.index, want.Name)

	ctxStr := fmt.Sprintf("Gochart %s, State %q, Parent %q, Trigger %d (%s)", sc.gochartName, want.Name, sc.parentState, sc.index, want.Name)

	// Compare options.
	if assert.Equal(t, len(want.Options), len(got.Options), ctxStr) {
		for i := 0; i < len(want.Options); i++ {
			want := want.Options[i]
			got := got.Options[i]

			assert.Equal(t, want.Name, got.Name, "%s, Option %d", ctxStr, i)
			assert.Equal(t, want.ValueKind, got.ValueKind, "%s, Option %d", ctxStr, i)
			assert.Equal(t, want.Value, got.Value, "%s, Option %d (%s)", ctxStr, i, want.Name)
		}
	}

	// Compare enter reactions.
	if assert.Equal(t, len(want.EnterReactions), len(got.EnterReactions), ctxStr) {
		for i := 0; i < len(want.EnterReactions); i++ {
			want := want.EnterReactions[i]
			got := got.EnterReactions[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Option %d", ctxStr, i)
			assert.Equal(t, want.entryIndex, got.entryIndex, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare exit reactions.
	if assert.Equal(t, len(want.ExitReactions), len(got.ExitReactions), ctxStr) {
		for i := 0; i < len(want.ExitReactions); i++ {
			want := want.ExitReactions[i]
			got := got.ExitReactions[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Option %d", ctxStr, i)
			assert.Equal(t, want.entryIndex, got.entryIndex, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare transitions.
	if assert.Equal(t, len(want.Transitions), len(got.Transitions), ctxStr) {
		for i := 0; i < len(want.Transitions); i++ {
			want := want.Transitions[i]
			got := got.Transitions[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Transition %d", ctxStr, i)
			assert.Equal(t, want.TargetState, got.TargetState, "%s, Transition %d", ctxStr, i)
			assert.Equal(t, escapeString(want.Condition), escapeString(got.Condition), "%s, Transition %d", ctxStr, i)
			assert.Equal(t, escapeString(want.Action), escapeString(got.Action), "%s, Transition %d", ctxStr, i)
			assert.Equal(t, want.entryIndex, got.entryIndex, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare captures.
	if assert.Equal(t, len(want.Captures), len(got.Captures), ctxStr) {
		for i := 0; i < len(want.Captures); i++ {
			want := want.Captures[i]
			got := got.Captures[i]

			assert.Equal(t, want.Trigger, got.Trigger, "%s, Capture %d", ctxStr, i)
			assert.Equal(t, escapeString(want.Condition), escapeString(got.Condition), "%s, Capture %d", ctxStr, i)
			assert.Equal(t, want.entryIndex, got.entryIndex, "%s, Option %d", ctxStr, i)
		}
	}

	// Compare children.
	parentStateName := want.Name
	if assert.Equal(t, len(want.Children), len(got.Children), ctxStr) {
		for i := 0; i < len(want.Children); i++ {
			want := want.Children[i]
			got := got.Children[i]

			subComparer := &stateComparer{
				index:       i,
				gochartName: sc.gochartName,
				parentState: parentStateName,
			}
			subComparer.compare(t, want, got)
		}
	}

	// Compare parent.
	assert.Equal(t, want.ParentName, got.ParentName, ctxStr)

	// Compare comments.
	parent := fmt.Sprintf("state:%s", stateName)
	if assert.Equal(t, len(want.Comments), len(got.Comments), parent) {
		for i := 0; i < len(want.Comments); i++ {
			want := want.Comments[i]
			got := got.Comments[i]

			cc := &commentComparer{
				parent: parent,
			}
			cc.compare(t, want, got)
		}
	}
}

type commentComparer struct {
	parent string
}

func (cc *commentComparer) compare(t *testing.T, want, got *ASTNodeComment) {
	assert.Equal(t, want.comment, got.comment, "parent %q", cc.parent)
	assert.Equal(t, want.entryIndex, got.entryIndex, "parent %q (COMMENT=%q)", cc.parent, want.comment)
}

func escapeString(str string) string {
	return strings.ReplaceAll(str, "\r", "")
}
