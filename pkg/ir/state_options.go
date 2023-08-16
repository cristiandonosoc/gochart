package ir

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/pkg/frontend/output"
)

type StateOptionKind string

const (
	StateOptionKind_Default StateOptionKind = "default"
)

type StateOption struct {
	Kind      StateOptionKind
	ValueKind output.StateOptionValueKind

	stringValue string
	intValue    int
	boolValue   bool
}

func (so *StateOption) String() string {
	return string(so.Kind)
}

func (so *StateOption) StringValue() string {
	if so.ValueKind != output.StateOptionValueKind_String {
		panic(fmt.Sprintf("Accessing wrong type, wanted string, got %s", so.ValueKind))
	}

	return so.stringValue
}

func (so *StateOption) IntValue() int {
	if so.ValueKind != output.StateOptionValueKind_Integer {
		panic(fmt.Sprintf("Accessing wrong type, wanted int, got %s", so.ValueKind))
	}

	return so.intValue
}

func (so *StateOption) BoolValue() bool {
	if so.ValueKind != output.StateOptionValueKind_Boolean {
		panic(fmt.Sprintf("Accessing wrong type, wanted boolean, got %s", so.ValueKind))
	}

	return so.boolValue
}

// Option parser -----------------------------------------------------------------------------------

type optionParser struct {
	state *State
}

func newOptionParser(state *State) *optionParser {
	return &optionParser{
		state: state,
	}
}

func (op *optionParser) parseOptions(frontendOptions []*output.StateOptionData) error {
	for _, fo := range frontendOptions {
		switch fo.Name {
		case string(StateOptionKind_Default):
			if err := op.parseDefaultOption(fo); err != nil {
				return fmt.Errorf("parsing default option: %w", err)
			}
		default:
			return fmt.Errorf("unrecognized state option %q", fo.Name)
		}
	}

	return nil
}

func (op *optionParser) parseDefaultOption(fo *output.StateOptionData) error {
	if op.state.Default {
		return fmt.Errorf("default option already found")
	}

	if fo.ValueKind != output.StateOptionValueKind_None {
		return fmt.Errorf("default option has no value associated, found %q", fo.ValueKind)
	}

	op.state.Default = true

	return nil
}
