package ir

import (
	"fmt"
	"strconv"

	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

type OptionDefinition struct {
	Name        string
	Scope       string
	Kind        output.OptionValueKind
	Description string
	Handler     any
}

var (
	kStateOptionDefinitions = map[string]*OptionDefinition{
		"default": {
			Name:        "default",
			Scope:       "state",
			Kind:        output.OptionValueKind_Boolean,
			Description: "Whether this particular state should be considered the default entry point for the parent state for the parent state.",
			Handler:     StateOptionHandler(parseStateOption_Default),
		},
	}
)

// Option represents the parsed option that comes from the frontend, which some getters for the
// types.
type Option struct {
	Definition *OptionDefinition

	stringValue string
	intValue    int64
	boolValue   bool

	frontendData *output.OptionData
}

func newOption(od *OptionDefinition, fo *output.OptionData) (*Option, error) {
	// Check that the expected kind is the same as the provided one.
	if fo.Kind != od.Kind {
		return nil, fmt.Errorf("expected option %q to be %q, got %q instead", fo.Name, od.Kind, fo.Kind)
	}

	option := &Option{
		Definition:   od,
		frontendData: fo,
	}

	switch od.Kind {
	case output.OptionValueKind_Boolean:
		value, err := strconv.ParseBool(fo.Value)
		if err != nil {
			return nil, fmt.Errorf("parsing bool value %q: %w", fo.Value, err)
		}
		option.boolValue = value
	case output.OptionValueKind_String:
		option.stringValue = fo.Value
	case output.OptionValueKind_Integer:
		value, err := strconv.ParseInt(fo.Value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parsing integer value %q: %w", fo.Value, err)
		}
		option.intValue = value
	default:
		panic(fmt.Sprintf("unknown OptionValueKind %v", od.Kind))
	}

	return option, nil
}

func (so *Option) Name() string {
	return so.Definition.Name
}

func (so *Option) StringValue() string {
	if so.Definition.Kind != output.OptionValueKind_String {
		panic(fmt.Sprintf("Accessing wrong type, wanted string, got %s", so.Definition.Kind))
	}
	return so.stringValue
}

func (so *Option) IntValue() int64 {
	if so.Definition.Kind != output.OptionValueKind_Integer {
		panic(fmt.Sprintf("Accessing wrong type, wanted int, got %s", so.Definition.Kind))
	}
	return so.intValue
}

func (so *Option) BoolValue() bool {
	if so.Definition.Kind != output.OptionValueKind_Boolean {
		panic(fmt.Sprintf("Accessing wrong type, wanted boolean, got %s", so.Definition.Kind))
	}
	return so.boolValue
}

// State Options -----------------------------------------------------------------------------------

type StateOptionHandler func(state *State, option *Option) error

func parseStateOptions(state *State, frontendOptions []*output.OptionData) error {
	set := map[string]struct{}{}

	var options []*Option

	for _, fo := range frontendOptions {
		// See if we can find the option definition.
		od, ok := kStateOptionDefinitions[fo.Name]
		if !ok {
			return fmt.Errorf("unrecognized state option %q", fo.Name)
		}

		// Check if we haven't parsed this option already.
		if _, ok := set[od.Name]; ok {
			return fmt.Errorf("option %q set more than once", od.Name)
		}

		// Try to parse the option.
		option, err := newOption(od, fo)
		if err != nil {
			return fmt.Errorf("parsing state option %q: %w", fo.Name, err)
		}

		// Cast to the handler
		handler, ok := od.Handler.(StateOptionHandler)
		if !ok {
			panic(fmt.Sprintf("invalid handler type for state option %q", fo.Name))
		}

		if err := handler(state, option); err != nil {
			return fmt.Errorf("handling state option %q: %w", fo.Name, err)
		}

		options = append(options, option)
	}

	state.Options = options
	return nil
}

func parseStateOption_Default(state *State, option *Option) error {
	state.Default = option.BoolValue()
	return nil
}
