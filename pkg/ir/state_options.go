package ir

import (
	"fmt"

	"github.com/cristiandonosoc/gochart/pkg/frontend"
)

//go:generate stringer -type=StateOptionKind
type StateOptionKind int

const (
	StateOptionKind_Default StateOptionKind = iota
)

type StateOption struct {
	Kind      StateOptionKind
	ValueKind frontend.StateOptionValueKind

	stringValue string
	intValue    int
	boolValue   bool
}

func (so *StateOption) Name() string {
	return so.Kind.String()
}

func (so *StateOption) StringValue() string {
	if so.ValueKind != frontend.StateOptionValueKind_String {
		panic(fmt.Sprintf("Accessing wrong type, wanted string, got %s", so.ValueKind))
	}

	return so.stringValue
}

func (so *StateOption) IntValue() int {
	if so.ValueKind != frontend.StateOptionValueKind_Integer {
		panic(fmt.Sprintf("Accessing wrong type, wanted int, got %s", so.ValueKind))
	}

	return so.intValue
}

func (so *StateOption) BoolValue() bool {
	if so.ValueKind != frontend.StateOptionValueKind_Boolean {
		panic(fmt.Sprintf("Accessing wrong type, wanted boolean, got %s", so.ValueKind))
	}

	return so.boolValue
}
