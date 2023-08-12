// Package yaml is a simple frontend that reads yaml. Mostly used to quickly test the whole pipeline
// instead of requiring a custom language/parser.
package yaml

import (
	"fmt"
	"io"

	"github.com/cristiandonosoc/gochart/pkg/frontend"

	"gopkg.in/yaml.v2"
)

func NewYamlFrontend() *yamlFrontend {
	return &yamlFrontend{}
}

var _ frontend.GochartFrontend = (*yamlFrontend)(nil)

type yamlFrontend struct {
}

func (yf *yamlFrontend) Process(r io.Reader) (*frontend.StatechartData, error) {
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

func (ft *frontendTranslator) translateStatechart(ysc *YamlStatechart) *frontend.StatechartData {
	// Collect Triggers.
	triggers := make([]*frontend.TriggerData, 0, len(ysc.Triggers))
	for _, yt := range ysc.Triggers {
		triggers = append(triggers, ft.translateTrigger(yt))
	}

	// Collect states.
	states := make([]*frontend.StateData, 0, len(ysc.States))
	for _, ys := range ysc.States {
		states = append(states, ft.translateState(ys))
	}

	transitions := make([]*frontend.TransitionData, 0, len(ysc.Transitions))
	for _, yt := range ysc.Transitions {
		transitions = append(transitions, ft.translateTransition(yt))
	}

	return &frontend.StatechartData{
		Name:        ysc.Name,
		Triggers:    triggers,
		States:      states,
		Transitions: transitions,
	}
}

func (ft *frontendTranslator) translateTrigger(yt *YamlTrigger) *frontend.TriggerData {
	index := ft.triggerIndex
	ft.triggerIndex += 1
	return &frontend.TriggerData{
		Name:       yt.Name,
		ArgsString: yt.ArgsString,
		Index:      index,
	}
}

func (ft *frontendTranslator) translateState(ys *YamlState) *frontend.StateData {
	index := ft.stateIndex
	ft.stateIndex++
	state := &frontend.StateData{
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

func (ft *frontendTranslator) translateStateOption(yso *YamlStateOption) *frontend.StateOptionData {
	option := &frontend.StateOptionData{
		Name:      yso.Name,
		ValueKind: frontend.StateOptionValueKind_None,
	}

	if yso.IntValue != "" {
		option.ValueKind = frontend.StateOptionValueKind_Integer
		option.Value = yso.IntValue
	} else if yso.StringValue != "" {
		option.ValueKind = frontend.StateOptionValueKind_String
		option.Value = yso.IntValue
	} else if yso.BoolValue != "" {
		option.ValueKind = frontend.StateOptionValueKind_Boolean
		option.Value = yso.IntValue
	}

	return option
}

func (ft *frontendTranslator) translateTransition(yt *YamlTransition) *frontend.TransitionData {
	index := ft.transitionIndex
	ft.transitionIndex++

	return &frontend.TransitionData{
		From:    yt.From,
		To:      yt.To,
		Trigger: yt.Trigger,
		Index:   index,
	}
}
