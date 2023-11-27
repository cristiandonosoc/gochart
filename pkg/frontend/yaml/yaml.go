// Package yaml is a simple frontend that reads yaml. Mostly used to quickly test the whole pipeline
// instead of requiring a custom language/parser.
package yaml

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/internal/option"
	"github.com/cristiandonosoc/gochart/pkg/frontend/output"

	"gopkg.in/yaml.v2"
)

func NewYamlFrontend(opts ...option.Option) (*yamlFrontend, error) {
	return &yamlFrontend{}, nil
}

type yamlFrontend struct {
}

func (yf *yamlFrontend) Process(lf *output.LoadedFile) (*output.StatechartData, error) {
	var yamlData YamlStatechart
	if err := yaml.Unmarshal(lf.Data, &yamlData); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}

	ft := newFrontendTranslator()
	scdata := ft.translateStatechart(&yamlData)

	return scdata, nil
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

func (ft *frontendTranslator) translateStatechart(ysc *YamlStatechart) *output.StatechartData {
	// Collect Triggers.
	triggers := make([]*output.TriggerData, 0, len(ysc.Triggers))
	for _, yt := range ysc.Triggers {
		triggers = append(triggers, ft.translateTrigger(yt))
	}

	// Collect states.
	states := make([]*output.StateData, 0, len(ysc.States))
	for _, ys := range ysc.States {
		states = append(states, ft.translateState(ys))
	}

	transitions := make([]*output.TransitionData, 0, len(ysc.Transitions))
	for _, yt := range ysc.Transitions {
		transitions = append(transitions, ft.translateTransition(yt))
	}

	return &output.StatechartData{
		Name:        ysc.Name,
		Triggers:    triggers,
		States:      states,
		Transitions: transitions,
	}
}

func (ft *frontendTranslator) translateTrigger(yt *YamlTrigger) *output.TriggerData {
	ft.triggerIndex += 1
	return &output.TriggerData{
		Name:       yt.Name,
		ArgsString: yt.ArgsString,
	}
}

func (ft *frontendTranslator) translateState(ys *YamlState) *output.StateData {
	ft.stateIndex++
	state := &output.StateData{
		Name:   ys.Name,
		Parent: ys.Parent,
	}

	for _, opt := range ys.Options {
		state.Options = append(state.Options, ft.translateOption(opt))
	}

	for _, er := range ys.EnterReactions {
		state.EnterReactions = append(state.EnterReactions, er)
	}

	for _, er := range ys.ExitReactions {
		state.ExitReactions = append(state.ExitReactions, er)
	}

	for _, c := range ys.Captures {
		state.Captures = append(state.Captures, &output.CaptureData{
			Trigger:   c.Trigger,
			Condition: c.Condition,
		})
	}

	return state
}

func (ft *frontendTranslator) translateOption(yso *YamlOption) *output.OptionData {
	option := &output.OptionData{
		Name:  yso.Name,
		Kind:  output.OptionValueKind_Boolean,
		Value: "true",
	}

	if yso.IntValue != "" {
		option.Kind = output.OptionValueKind_Integer
		option.Value = yso.IntValue
	} else if yso.StringValue != "" {
		option.Kind = output.OptionValueKind_String
		option.Value = yso.IntValue
	} else if yso.BoolValue != "" {
		option.Kind = output.OptionValueKind_Boolean
		option.Value = yso.IntValue
	}

	return option
}

func (ft *frontendTranslator) translateTransition(yt *YamlTransition) *output.TransitionData {
	ft.transitionIndex++

	return &output.TransitionData{
		From:      yt.From,
		To:        yt.To,
		Trigger:   yt.Trigger,
		Condition: yt.Condition,
		Action:    yt.Action,
	}
}
