// Package yaml is a simple frontend that reads yaml. Mostly used to quickly test the whole pipeline
// instead of requiring a custom language/parser.
package yaml

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/pkg/frontend/output"

	"gopkg.in/yaml.v2"
)

func NewYamlFrontend() *yamlFrontend {
	return &yamlFrontend{}
}

type yamlFrontend struct {
}

func (yf *yamlFrontend) Process(r io.Reader) (*output.StatechartData, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input reader: %w", err)
	}

	var yamlData YamlStatechart
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
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
	index := ft.triggerIndex
	ft.triggerIndex += 1
	return &output.TriggerData{
		Name:       yt.Name,
		ArgsString: yt.ArgsString,
		Index:      index,
	}
}

func (ft *frontendTranslator) translateState(ys *YamlState) *output.StateData {
	index := ft.stateIndex
	ft.stateIndex++
	state := &output.StateData{
		Name:   ys.Name,
		Parent: ys.Parent,
		Index:  index,
	}

	for _, opt := range ys.Options {
		state.Options = append(state.Options, ft.translateStateOption(opt))
	}

	for _, er := range ys.EnterReactions {
		state.EnterReactions = append(state.EnterReactions, er)
	}

	for _, er := range ys.ExitReactions {
		state.ExitReactions = append(state.ExitReactions, er)
	}

	return state
}

func (ft *frontendTranslator) translateStateOption(yso *YamlStateOption) *output.StateOptionData {
	option := &output.StateOptionData{
		Name:      yso.Name,
		ValueKind: output.StateOptionValueKind_None,
	}

	if yso.IntValue != "" {
		option.ValueKind = output.StateOptionValueKind_Integer
		option.Value = yso.IntValue
	} else if yso.StringValue != "" {
		option.ValueKind = output.StateOptionValueKind_String
		option.Value = yso.IntValue
	} else if yso.BoolValue != "" {
		option.ValueKind = output.StateOptionValueKind_Boolean
		option.Value = yso.IntValue
	}

	return option
}

func (ft *frontendTranslator) translateTransition(yt *YamlTransition) *output.TransitionData {
	index := ft.transitionIndex
	ft.transitionIndex++

	return &output.TransitionData{
		From:    yt.From,
		To:      yt.To,
		Trigger: yt.Trigger,
		Index:   index,
	}
}
