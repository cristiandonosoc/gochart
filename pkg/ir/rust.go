package ir

import (
	"fmt"
	"strings"
	"unicode"
)

// ParseRustArguments parses trigger arguments written as in a Rust function declaration:
// "name: Type, other: Vec<(u32, Fx)>". References and lifetimes are not allowed, as the generated
// code stores the arguments in an owned event.
func ParseRustArguments(argString string) ([]*TriggerArgument, error) {
	splitArgs, err := splitRustArguments(argString)
	if err != nil {
		return nil, fmt.Errorf("splitting arguments for %q: %w", argString, err)
	}

	names := make(map[string]struct{})
	args := make([]*TriggerArgument, 0, len(splitArgs))
	for _, argStr := range splitArgs {
		arg, err := parseRustArgument(argStr)
		if err != nil {
			return nil, fmt.Errorf("parsing argument %q: %w", argStr, err)
		}

		if _, ok := names[arg.Name]; ok {
			return nil, fmt.Errorf("argument %q: argument is defined twice", arg.Name)
		}
		names[arg.Name] = struct{}{}

		args = append(args, arg)
	}

	return args, nil
}

// splitRustArguments separates the arguments on the commas that are not nested within generics,
// tuples or arrays. Eg:
// "a: HashMap<u32, Fx>, b: (u8, u8)" -> "a: HashMap<u32, Fx>", "b: (u8, u8)"
func splitRustArguments(argString string) ([]string, error) {
	var args []string

	// The stack of currently open delimiters, so we can check that they close in order.
	var open []rune
	closers := map[rune]rune{'>': '<', ')': '(', ']': '['}

	var current []rune
	for i, r := range argString {
		switch {
		case r == ',' && len(open) == 0:
			args = append(args, string(current))
			current = current[:0]
			continue
		case r == '<' || r == '(' || r == '[':
			open = append(open, r)
		case r == '>' || r == ')' || r == ']':
			if len(open) == 0 || open[len(open)-1] != closers[r] {
				return nil, fmt.Errorf("char %d: unmatched %q", i, r)
			}
			open = open[:len(open)-1]
		case r == '&' || r == '\'':
			return nil, fmt.Errorf("char %d: references and lifetimes are not allowed in trigger arguments (arguments are owned by the generated event)", i)
		case unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_: ;,\t", r):
		default:
			return nil, fmt.Errorf("char %d: invalid char %q", i, r)
		}

		current = append(current, r)
	}

	if len(open) != 0 {
		return nil, fmt.Errorf("unterminated %q", open[len(open)-1])
	}

	// A trailing comma is fine, as rustfmt would write it.
	if last := strings.TrimSpace(string(current)); last != "" || len(args) == 0 {
		args = append(args, string(current))
	}

	cleaned := make([]string, 0, len(args))
	for _, arg := range args {
		cleaned = append(cleaned, strings.Join(strings.Fields(arg), " "))
	}

	return cleaned, nil
}

// parseRustArgument parses a single "name: Type" argument.
func parseRustArgument(argString string) (*TriggerArgument, error) {
	name, argType, ok := strings.Cut(argString, ":")
	if !ok {
		return nil, fmt.Errorf("expected an argument in the form \"name: Type\"")
	}
	name = strings.TrimSpace(name)
	argType = strings.TrimSpace(argType)

	if !IsRustIdentifier(name) {
		return nil, fmt.Errorf("invalid argument name %q", name)
	}
	if argType == "" {
		return nil, fmt.Errorf("argument %q has no type", name)
	}
	if strings.HasPrefix(argType, ":") {
		return nil, fmt.Errorf("argument %q: expected \"name: Type\", got a path", name)
	}

	return &TriggerArgument{
		Type: argType,
		Name: name,
	}, nil
}

// IsRustIdentifier returns whether |s| is a plain (ASCII) Rust identifier.
func IsRustIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
		if r == '_' || unicode.IsLetter(r) {
			continue
		}
		if i > 0 && unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}
