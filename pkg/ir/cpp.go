package ir

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/bradenaw/juniper/xslices"
)

func ParseCppArguments(argString string) ([]*TriggerArgument, error) {
	cpp := &cppArgumentParser{
		allowedChars: []rune{'_', '<', '>', ':', ' ', '&', '(', ')', ','},
	}
	return cpp.parseArguments(argString)
}

// cppArgumentParser is a helper struct to parse arguments of the C++ language.
type cppArgumentParser struct {
	// We ensure that we don't have weird characters. Either letter, number or one of these.
	// NOTE: This could be a source of annoyance, but for now we prefer to be more strict.
	allowedChars []rune
}

// parseArguments separates the string as it represents cpp arguments in a function declaration.
// We use some very simple heuristics to parse the input.
// IMPORTANT: THIS WILL NOT COVER ALL CASES!
func (cpp *cppArgumentParser) parseArguments(argString string) ([]*TriggerArgument, error) {
	splitArgs, err := cpp.splitArguments(argString)
	if err != nil {
		return nil, fmt.Errorf("splitting arguments for %q: %w", argString, err)
	}

	// For each argument, we define separate the the type from the name.
	names := make(map[string]struct{})
	args := make([]*TriggerArgument, 0, len(splitArgs))
	for _, argStr := range splitArgs {
		arg, err := cpp.parseArgument(argStr)
		if err != nil {
			return nil, fmt.Errorf("parsing argument %q: %w", argStr, err)
		}

		// We ensure arguments are not repeated.
		if _, ok := names[arg.Name]; ok {
			return nil, fmt.Errorf("argument %q: argument is defined twice", arg.Name)
		}
		names[arg.Name] = struct{}{}

		args = append(args, arg)
	}

	return args, nil
}

// splitArguments separates a string arguments into the separate arguments, that can be processed
// on their own. Eg:
// "const std::vector<Foo, int>& list, int count" -> "const std::vector<Foo, in>& list", "int count"
func (cpp *cppArgumentParser) splitArguments(argString string) ([]string, error) {
	var args []string

	templateCount := 0
	var current []rune
	for i, r := range argString {
		// If we find a comma, and we're not in a template, whatever we found is an argument.
		if r == ',' && templateCount == 0 {
			args = append(args, string(current))
			current = current[:0] // Clear.
			continue
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if xslices.Count(cpp.allowedChars, r) == 0 {
				return nil, fmt.Errorf("char %d: invalid char %q", i, r)
			}
		}

		// Otherwise, it means that we're accumulating the string for the argument.
		current = append(current, r)

		// The only rule that we need to track is whether we're opening templates ("<" character), which
		// can validly use commas.
		if r == '<' {
			templateCount += 1
		} else if r == '>' {
			templateCount -= 1
			if templateCount < 0 {
				return nil, fmt.Errorf("char %d: closing unopening template", i)
			}
		}
	}

	// See if we're in an unterminated template.
	if templateCount != 0 {
		return nil, fmt.Errorf("unterminated template")
	}

	// We add the remainder as the last argument.
	args = append(args, string(current))

	// We perform some cleanup over the arguments.
	cleaned := make([]string, 0, len(args))
	for _, arg := range args {
		// We split and join via spaces, thus removing any extra spaces.
		cleaned = append(cleaned, strings.Join(strings.Fields(arg), " "))
	}

	return cleaned, nil
}

// parseArgument takes a string representing a single C++ argument and parses into the ir type.
// The rule is simple: the last word is the argument name. It must have no reference or pointer
// qualifier (eg. *, &). Those must be embedded in the type.
func (cpp *cppArgumentParser) parseArgument(argString string) (*TriggerArgument, error) {
	fields := strings.Fields(argString)
	if len(fields) < 2 {
		return nil, fmt.Errorf("an argument requires at least 2 words: <type> <name>")
	}

	argType := strings.Join(fields[:len(fields)-1], " ")
	argName := fields[len(fields)-1]

	// Verify that the name has no weird stuff.
	for _, r := range argName {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return nil, fmt.Errorf("unexpected character in argument name %q", r)
		}
	}

	return &TriggerArgument{
		Type: argType,
		Name: argName,
	}, nil
}
