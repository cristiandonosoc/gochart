// Package yaml is a simple frontend that reads yaml. Mostly used to quickly test the whole pipeline
// instead of requiring a custom language/parser.
package yaml

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/files"
	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/frontend/common"

	"gopkg.in/yaml.v2"
)

func NewYamlFrontend(opts ...option.Option) (*yamlFrontend, error) {
	return &yamlFrontend{}, nil
}

type yamlFrontend struct {
}

func (yf *yamlFrontend) Process(lf *files.LoadedFile) (*common.GochartData, error) {
	var yamlData YamlGochart
	if err := yaml.Unmarshal(lf.Data, &yamlData); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}

	ft := newFrontendTranslator()
	gcdata := ft.translateGochart(&yamlData)

	return gcdata, nil
}

// YAML TO OUTPUT TRANSLATION ----------------------------------------------------------------------

type frontendTranslator struct {
	triggerIndex    int
	stateIndex      int
	transitionIndex int
}

func newFrontendTranslator() *frontendTranslator {
	return &frontendTranslator{}
}

func (ft *frontendTranslator) translateGochart(ygc *YamlGochart) *common.GochartData {
	// Collect Triggers.
	triggers := make([]*common.TriggerData, 0, len(ygc.Triggers))
	for _, yt := range ygc.Triggers {
		triggers = append(triggers, ft.translateTrigger(yt))
	}

	// Collect states.
	states := make([]*common.StateData, 0, len(ygc.States))
	for _, ys := range ygc.States {
		states = append(states, ft.translateState(ys))
	}

	transitions := make([]*common.TransitionData, 0, len(ygc.Transitions))
	for _, yt := range ygc.Transitions {
		transitions = append(transitions, ft.translateTransition(yt))
	}

	// We create the root state and parent the yaml roots to it.
	root := &common.StateData{
		Name: common.RootStateName,
	}
	for _, state := range states {
		if state.Parent == "" {
			state.Parent = common.RootStateName
		}
	}
	states = append([]*common.StateData{root}, states...)

	return &common.GochartData{
		Name:        ygc.Name,
		Triggers:    triggers,
		Root:        root,
		States:      states,
		Transitions: transitions,
	}
}

func (ft *frontendTranslator) translateTrigger(yt *YamlTrigger) *common.TriggerData {
	ft.triggerIndex += 1
	return &common.TriggerData{
		Name:       yt.Name,
		ArgsString: yt.ArgsString,
	}
}

func (ft *frontendTranslator) translateState(ys *YamlState) *common.StateData {
	ft.stateIndex++
	state := &common.StateData{
		Name:   ys.Name,
		Parent: ys.Parent,
	}

	for _, opt := range ys.Options {
		state.Options = append(state.Options, ft.translateOption(opt))
	}

	for _, er := range ys.EnterReactions {
		state.EnterReactions = append(state.EnterReactions, &common.ReactionData{
			Trigger: er,
		})
	}

	for _, er := range ys.ExitReactions {
		state.ExitReactions = append(state.ExitReactions, &common.ReactionData{
			Trigger: er,
		})
	}

	for _, c := range ys.Captures {
		state.Captures = append(state.Captures, &common.CaptureData{
			Trigger:   c.Trigger,
			Condition: c.Condition,
		})
	}

	return state
}

func (ft *frontendTranslator) translateOption(yso *YamlOption) *common.OptionData {
	option := &common.OptionData{
		Name:  yso.Name,
		Kind:  common.OptionValueKind_Boolean,
		Value: "true",
	}

	if yso.IntValue != "" {
		option.Kind = common.OptionValueKind_Integer
		option.Value = yso.IntValue
	} else if yso.StringValue != "" {
		option.Kind = common.OptionValueKind_String
		option.Value = yso.IntValue
	} else if yso.BoolValue != "" {
		option.Kind = common.OptionValueKind_Boolean
		option.Value = yso.IntValue
	}

	return option
}

func (ft *frontendTranslator) translateTransition(yt *YamlTransition) *common.TransitionData {
	ft.transitionIndex++

	return &common.TransitionData{
		From:      yt.From,
		To:        yt.To,
		Trigger:   yt.Trigger,
		Condition: yt.Condition,
		Action:    yt.Action,
	}
}
