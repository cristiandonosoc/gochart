package ir

import (
	"fmt"
	"strconv"

	"github.com/cristiandonosoc/gochart/internal/errors"
	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

type OptionDefinition struct {
	Name        string
	Kind        frontend_common.OptionValueKind
	Description string
	Handler     any

	TestingOption bool
}

var (
	// Gochart Options ----------------------------------------------------------------------------

	kGochartOption_Concurrent = "concurrent"
	kGochartOption_Namespace  = "namespace"

	kGochartOptionDefinitions = addHiddenTestingOptions(map[string]*OptionDefinition{
		kGochartOption_Concurrent: {
			Name:        kGochartOption_Concurrent,
			Kind:        frontend_common.OptionValueKind_Boolean,
			Description: "Whether the root states of the gochart should be considered concurrent within each other.",
			Handler:     GochartOptionHandler(parseGochartOption_Concurrent),
		},
		kGochartOption_Namespace: {
			Name:        kGochartOption_Namespace,
			Kind:        frontend_common.OptionValueKind_String,
			Description: "The name of the namespace to generate. Otherwise the name of gochart will be used.",
			Handler:     GochartOptionHandler(parseGochartOption_Namespace),
		},
	})

	// State Options ---------------------------------------------------------------------------------

	kStateOption_Default    = "default"
	kStateOption_Concurrent = "concurrent"

	kStateOptionDefinitions = addHiddenTestingOptions(map[string]*OptionDefinition{
		kStateOption_Default: {
			Name:        kStateOption_Default,
			Kind:        frontend_common.OptionValueKind_Boolean,
			Description: "Whether this particular state should be considered the default entry point for the parent state for the parent state.",
			Handler:     StateOptionHandler(parseStateOption_Default),
		},
		kStateOption_Concurrent: {
			Name:        kStateOption_Concurrent,
			Kind:        frontend_common.OptionValueKind_Boolean,
			Description: "Whether this state is to be considered a concurrent one.",
			Handler:     StateOptionHandler(parseStateOption_Concurrent),
		},
	})
)

func addHiddenTestingOptions(options map[string]*OptionDefinition) map[string]*OptionDefinition {
	options["test_bool"] = &OptionDefinition{
		Name:          "test_bool",
		Kind:          frontend_common.OptionValueKind_Boolean,
		Description:   "Option added for testing purposes.",
		TestingOption: true,
	}
	options["test_int"] = &OptionDefinition{
		Name:          "test_int",
		Kind:          frontend_common.OptionValueKind_Integer,
		Description:   "Option added for testing purposes.",
		TestingOption: true,
	}
	options["test_string"] = &OptionDefinition{
		Name:          "test_string",
		Kind:          frontend_common.OptionValueKind_String,
		Description:   "Option added for testing purposes.",
		TestingOption: true,
	}

	return options
}

// Option represents the parsed option that comes from the frontend, which some getters for the
// types.
type Option struct {
	Definition *OptionDefinition

	stringValue string
	intValue    int64
	boolValue   bool
	// implicitBool refers whether the boolean value was implicit or not (simply writing the name,
	// rather than an explicit "= true"). This doesn't change the logic of the option, but it's useful
	// for downstream systems (like formatters) to know in which way the option was written.
	implicitBool bool

	frontendData *frontend_common.OptionData
}

func newOption(od *OptionDefinition, fo *frontend_common.OptionData) (*Option, error) {
	// Check that the expected kind is the same as the provided one.
	if fo.Kind != od.Kind {
		return nil, errors.Errorf(fo.Frontend_NamePos, "Semantic", "expected option %q to be %q, got %q instead", fo.Name, od.Kind, fo.Kind)
	}

	option := &Option{
		Definition:   od,
		frontendData: fo,
	}

	switch od.Kind {
	case frontend_common.OptionValueKind_Boolean:
		implicit := true
		value := true
		if fo.Value != "" {
			v, err := strconv.ParseBool(fo.Value)
			if err != nil {
				return nil, errors.Errorf(fo.Frontend_NamePos, "Semantic", "parsing bool value %q: %w", fo.Value, err)
			}

			value = v
			implicit = false
		}
		option.boolValue = value
		option.implicitBool = implicit
	case frontend_common.OptionValueKind_String:
		option.stringValue = fo.Value
	case frontend_common.OptionValueKind_Integer:
		value, err := strconv.ParseInt(fo.Value, 10, 32)
		if err != nil {
			return nil, errors.Errorf(fo.Frontend_NamePos, "Semantic", "parsing integer value %q: %w", fo.Value, err)
		}
		option.intValue = value
	default:
		return nil, errors.InternalErrorf("unknown OptionValueKind %v", od.Kind)
	}

	return option, nil
}

func collectOptionsFromDefinitions(level string,
	definitions map[string]*OptionDefinition,
	frontendOptions []*frontend_common.OptionData) ([]*Option, error) {

	set := map[string]struct{}{}

	var options []*Option

	for _, fo := range frontendOptions {
		// See if we can find the option definition.
		od, ok := definitions[fo.Name]
		if !ok {
			return nil, errors.Errorf(fo.Frontend_NamePos, "Semantic", "unrecognized %s option %q", level, fo.Name)
		}

		// Check if we haven't parsed this option already.
		if _, ok := set[od.Name]; ok {
			return nil, errors.Errorf(fo.Frontend_NamePos, "Semantic", "option %q set more than once", od.Name)
		}
		set[od.Name] = struct{}{}

		// Try to parse the option.
		option, err := newOption(od, fo)
		if err != nil {
			return nil, fmt.Errorf("parsing state option %q: %w", fo.Name, err)
		}

		// If it is a testing option, we just need to parse it. No need to call the handler.
		if od.TestingOption {
			continue
		}

		options = append(options, option)
	}

	return options, nil
}

func (so *Option) Name() string {
	return so.Definition.Name
}

func (so *Option) StringValue() string {
	if so.Definition.Kind != frontend_common.OptionValueKind_String {
		panic(fmt.Sprintf("Accessing wrong type, wanted string, got %s", so.Definition.Kind))
	}
	return so.stringValue
}

func (so *Option) IntValue() int64 {
	if so.Definition.Kind != frontend_common.OptionValueKind_Integer {
		panic(fmt.Sprintf("Accessing wrong type, wanted int, got %s", so.Definition.Kind))
	}
	return so.intValue
}

func (so *Option) BoolValue() bool {
	if so.Definition.Kind != frontend_common.OptionValueKind_Boolean {
		panic(fmt.Sprintf("Accessing wrong type, wanted boolean, got %s", so.Definition.Kind))
	}
	return so.boolValue
}

func (so *Option) IsImplicitBool() bool {
	if so.Definition.Kind != frontend_common.OptionValueKind_Boolean {
		panic(fmt.Sprintf("Accessing wrong type, wanted boolean, got %s", so.Definition.Kind))
	}
	return so.implicitBool
}

// Gochart Options ------------------------------------------------------------------------------

type GochartOptionHandler func(gc *Gochart, option *Option) error

func parseGochartOptions(gc *Gochart, frontendOptions []*frontend_common.OptionData) error {
	options, err := collectOptionsFromDefinitions("gochart", kGochartOptionDefinitions, frontendOptions)
	if err != nil {
		return fmt.Errorf("collecting gochart options from definitions: %w", err)
	}

	for _, option := range options {
		// Cast the handler.
		handler, ok := option.Definition.Handler.(GochartOptionHandler)
		if !ok {
			return errors.InternalErrorf("invalid handler type for gochart option %q", option.Definition.Name)
		}

		if err := handler(gc, option); err != nil {
			return fmt.Errorf("handling gochart option %q: %w", option.Definition.Name, err)
		}
	}

	gc.Options = options
	return nil
}

func parseGochartOption_Concurrent(gc *Gochart, option *Option) error {
	gc.Root.Concurrent = option.BoolValue()
	return nil
}

func parseGochartOption_Namespace(gc *Gochart, option *Option) error {
	gc.Namespace = option.StringValue()
	return nil
}

// State Options -----------------------------------------------------------------------------------

type StateOptionHandler func(state *State, option *Option) error

func parseStateOptions(state *State, frontendOptions []*frontend_common.OptionData) error {
	options, err := collectOptionsFromDefinitions("state", kStateOptionDefinitions, frontendOptions)
	if err != nil {
		return fmt.Errorf("collecting state options from definitions: %w", err)
	}

	for _, option := range options {
		// Cast the handler.
		handler, ok := option.Definition.Handler.(StateOptionHandler)
		if !ok {
			return errors.InternalErrorf("invalid handler type for state option %q", option.Definition.Name)
		}

		if err := handler(state, option); err != nil {
			return fmt.Errorf("handling state option %q: %w", option.Definition.Name, err)
		}
	}

	state.Options = options
	return nil
}

func parseStateOption_Default(state *State, option *Option) error {
	state.Default = option.BoolValue()
	return nil
}

func parseStateOption_Concurrent(state *State, option *Option) error {
	state.Concurrent = option.BoolValue()
	return nil
}
