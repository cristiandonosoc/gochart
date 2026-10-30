package ir

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/errors"
	frontend_common "github.com/cristiandonosoc/gochart/pkg/frontend/common"
)

type OptionDefinition struct {
	// Name is the actual value that needs to be written in the gochart.
	Name string
	// FlagName is the public "documentation" driven name of this option.
	FlagName    string
	Kind        frontend_common.OptionValueKind
	Description string
	Synopsis    string
	Handler     any

	TestingOption bool
}

var (
	// Gochart Options ----------------------------------------------------------------------------

	kGochartOption_Parallel  = "parallel"
	kGochartOption_Namespace = "namespace"
	kGochartOption_Language  = "language"
	kGochartOption_Derives   = "derives"

	kGochartOption_DelegateDefaults = "delegate_defaults"
	kGochartOption_DefaultActive    = "default_active"

	kGochartOptionDefinitions = addHiddenTestingOptions(map[string]*OptionDefinition{
		kGochartOption_Parallel: {
			Name:     kGochartOption_Parallel,
			FlagName: kGochartOption_Parallel,
			Kind:     frontend_common.OptionValueKind_Boolean,
			Description: `Whether the root states of the gochart should be considered running in parallel within each other.

When this option is defined, all child states will run as separate parallel gochart groups and the
gochart is considered to be in all those states simultaneously.

See https://github.com/cristiandonosoc/gochart/docs/concurrent.md for more details.`,
			Synopsis: `
gochart(parallel) {
	...
	state ParallelChild1 {
		...
	}

	state ParallelChild2 {
		...
	}
}`,
			Handler: GochartOptionHandler(parseGochartOption_Parallel),
		},
		kGochartOption_Namespace: {
			Name:     kGochartOption_Namespace,
			FlagName: kGochartOption_Namespace,
			Kind:     frontend_common.OptionValueKind_String,
			Description: `The name of the namespace to generate in |gochart::<NAMESPACE>|.
Otherwise the name of gochart will be used.`,
			Synopsis: `
gochart(namespace = "SomeOtherNamespace") {
	...
}
`,
			Handler: GochartOptionHandler(parseGochartOption_Namespace),
		},
		kGochartOption_Language: {
			Name:     kGochartOption_Language,
			FlagName: kGochartOption_Language,
			Kind:     frontend_common.OptionValueKind_String,
			Description: `The language the gochart is written for: "cpp" (default) or "rust".

It determines how trigger arguments are parsed and which backend can generate the gochart.
Rust gocharts use named callbacks ("if some_condition", "do some_action") instead of inlined code,
and support state data (see the state "data" option).

See https://github.com/cristiandonosoc/gochart/docs/rust.md for more details.`,
			Synopsis: `
gochart Hauler(language = "rust") {
	trigger Reserved(job: HaulJob)
	...
}
`,
			Handler: GochartOptionHandler(parseGochartOption_Language),
		},
		kGochartOption_Derives: {
			Name:     kGochartOption_Derives,
			FlagName: kGochartOption_Derives,
			Kind:     frontend_common.OptionValueKind_String,
			Description: `Extra derives for the generated Rust types, comma separated (Rust only).

The generated chart always derives Clone, Debug, PartialEq and Eq. This option adds more, such as
serde's, so the chart can be saved or hashed as part of its owner.`,
			Synopsis: `
gochart Hauler(language = "rust", derives = "serde::Serialize, serde::Deserialize") {
	...
}
`,
			Handler: GochartOptionHandler(parseGochartOption_Derives),
		},
		kGochartOption_DelegateDefaults: {
			Name:     kGochartOption_DelegateDefaults,
			FlagName: kGochartOption_DelegateDefaults,
			Kind:     frontend_common.OptionValueKind_Boolean,
			Description: `Gives every callback of the generated Delegate trait a default body that panics (Rust only).

A delegate then implements only the callbacks the triggers it sends can reach. Without this option,
the compiler checks that every callback is implemented; with it, a missing callback panics when it
is reached.`,
			Synopsis: `
gochart Hauler(language = "rust", delegate_defaults) {
	...
}
`,
			Handler: GochartOptionHandler(parseGochartOption_DelegateDefaults),
		},
		kGochartOption_DefaultActive: {
			Name:     kGochartOption_DefaultActive,
			FlagName: kGochartOption_DefaultActive,
			Kind:     frontend_common.OptionValueKind_Boolean,
			Description: `Makes the default chart start in its default states, instead of inactive (Rust only).

The chart can then be built with Default (or derived with its owner) without a delegate. Entering
the default states must not call any callback: no enter reactions, no null transitions on the
default leaves, and state data only through "data_default".`,
			Synopsis: `
gochart Hauler(language = "rust", default_active) {
	...
}
`,
			Handler: GochartOptionHandler(parseGochartOption_DefaultActive),
		},
	})

	// State Options ---------------------------------------------------------------------------------

	kStateOption_Default     = "default"
	kStateOption_Parallel    = "parallel"
	kStateOption_Data        = "data"
	kStateOption_DataDefault = "data_default"

	kStateOptionDefinitions = addHiddenTestingOptions(map[string]*OptionDefinition{
		kStateOption_Default: {
			Name:     kStateOption_Default,
			FlagName: kStateOption_Default,
			Kind:     frontend_common.OptionValueKind_Boolean,
			Description: `Whether this particular state should be considered the default entry point for the parent state for the parent state.

By default, the first child is considered to be the default entry point, but this can be override with this option.
There can only be one default child per state.

While valid, it is redundant to set the default flag for the first child.`,
			Synopsis: `
state Parent {
	state Child1 {
		...
	}

	state Child2(default) {
		...
	}
}
`,
			Handler: StateOptionHandler(parseStateOption_Default),
		},
		kStateOption_Parallel: {
			Name:     kStateOption_Parallel,
			FlagName: kStateOption_Parallel,
			Kind:     frontend_common.OptionValueKind_Boolean,
			Description: `Whether the child states of this state should be considered running in parallel within each other.

When this option is defined, all child states will run as separate parallel gochart groups and the
gochart is considered to be in all those states simultaneously.

See https://github.com/cristiandonosoc/gochart/docs/concurrent.md for more details.`,
			Synopsis: `
state SomeState(parallel) {
	...
	state ParallelChild1 {
		...
	}

	state ParallelChild2 {
		...
	}
}`,
			Handler: StateOptionHandler(parseStateOption_Parallel),
		},
		kStateOption_Data: {
			Name:     kStateOption_Data,
			FlagName: kStateOption_Data,
			Kind:     frontend_common.OptionValueKind_String,
			Description: `The type of the data that lives while this state is active (Rust only).

The data is created when the state is entered (by a callback, or with Default::default() if
"data_default" is set) and dropped when the state is exited. Callbacks of the state and its
children receive it.`,
			Synopsis: `
state Hauling(data = "HaulJob") {
	...
}
`,
			Handler: StateOptionHandler(parseStateOption_Data),
		},
		kStateOption_DataDefault: {
			Name:     kStateOption_DataDefault,
			FlagName: kStateOption_DataDefault,
			Kind:     frontend_common.OptionValueKind_Boolean,
			Description: `Create the state data with Default::default() instead of a constructor callback (Rust only).

Requires the "data" option.`,
			Synopsis: `
state Loading(data = "Transfer", data_default) {
	...
}
`,
			Handler: StateOptionHandler(parseStateOption_DataDefault),
		},
	})
)

func collectOptionDefinitions(definitions map[string]*OptionDefinition) []*OptionDefinition {
	options := make([]*OptionDefinition, 0, len(definitions))

	for _, definition := range definitions {
		if definition.TestingOption {
			continue
		}

		options = append(options, definition)
	}

	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})

	return options
}

// GochartOptionDefinitions returns a sorted list of options definitions.
func GochartOptionDefinitions() []*OptionDefinition {
	return collectOptionDefinitions(kGochartOptionDefinitions)
}

// StateOptionDefinitions returns a sorted list of options definitions.
func StateOptionDefinitions() []*OptionDefinition {
	return collectOptionDefinitions(kStateOptionDefinitions)
}

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

func parseGochartOption_Parallel(gc *Gochart, option *Option) error {
	gc.Root.Parallel = option.BoolValue()
	return nil
}

func parseGochartOption_Namespace(gc *Gochart, option *Option) error {
	gc.Namespace = option.StringValue()
	return nil
}

func parseGochartOption_Language(gc *Gochart, option *Option) error {
	switch language := Language(option.StringValue()); language {
	case Language_Cpp, Language_Rust:
		gc.Language = language
		return nil
	default:
		return errors.Errorf(option.frontendData.Frontend_NamePos, "Semantic",
			"unknown language %q, expected %q or %q", option.StringValue(), Language_Cpp, Language_Rust)
	}
}

func parseGochartOption_DelegateDefaults(gc *Gochart, option *Option) error {
	gc.DelegateDefaults = option.BoolValue()
	return nil
}

func parseGochartOption_DefaultActive(gc *Gochart, option *Option) error {
	gc.DefaultActive = option.BoolValue()
	return nil
}

func parseGochartOption_Derives(gc *Gochart, option *Option) error {
	gc.Derives = nil
	for _, derive := range strings.Split(option.StringValue(), ",") {
		if derive = strings.TrimSpace(derive); derive != "" {
			gc.Derives = append(gc.Derives, derive)
		}
	}
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

func parseStateOption_Parallel(state *State, option *Option) error {
	state.Parallel = option.BoolValue()
	return nil
}

func parseStateOption_Data(state *State, option *Option) error {
	state.DataType = strings.TrimSpace(option.StringValue())
	if state.DataType == "" {
		return errors.Errorf(option.frontendData.Frontend_NamePos, "Semantic", "state %q: empty data type", state.Name)
	}
	return nil
}

func parseStateOption_DataDefault(state *State, option *Option) error {
	state.DataDefault = option.BoolValue()
	return nil
}

// FindOption returns the option with the given name, or nil.
func FindOption(options []*Option, name string) *Option {
	for _, option := range options {
		if option.Name() == name {
			return option
		}
	}
	return nil
}
