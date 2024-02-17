package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cristiandonosoc/golib/pkg/files"
)

const (
	kContextSize = 2
)

type gochartError struct {
	Kind          string
	InternalError error
	pos           *files.LoadedFilePosition
}

var _ error = (*gochartError)(nil)

func (ge *gochartError) Error() string {
	return ge.InternalError.Error()
}

func (ge *gochartError) Render() string {
	if ge.pos == nil {
		lines := []string{
			"<no file position found>",
			ge.InternalError.Error(),
		}
		return strings.Join(lines, "\n")
	}

	unwrappedError, errorCount := unwrapError(ge.InternalError)

	header := fmt.Sprintf("%s error:", ge.Kind)
	msg := unwrappedError.Error()
	footer := ""
	if errorCount > 1 {
		footer = fmt.Sprintf("Full context: %s", ge.InternalError.Error())
	}

	render, err := Render(ge.pos, header, msg, footer, kContextSize)
	if err != nil {
		return fmt.Sprintf("ERROR RENDERING: %v", err)
	}

	return render
}

func Render(lfp *files.LoadedFilePosition, header, msg, footer string, contextSize int) (string, error) {
	if lfp.File == nil {
		return "", fmt.Errorf("no loaded file")
	}

	var output []string

	// Start with an empty line.
	output = append(output, "")
	if header != "" {
		output = append(output, header)
	}

	// Show a couple lines of context.
	begin := lfp.Line - contextSize
	end := lfp.Line + contextSize

	// Just in case the printing didn't get to print the msg.
	printedMsg := false

	for i := begin; i <= end; i++ {
		// NOTE: Line as 1-based.
		arrayIndex := i - 1
		if arrayIndex < 0 {
			continue
		}

		lines, err := lfp.File.Lines()
		if err != nil {
			return "", fmt.Errorf("reading lines from file: %w", err)
		}
		if arrayIndex >= len(lines) {
			break
		}

		line := lines[arrayIndex]

		// We print the line.
		output = append(output, fmt.Sprintf("%4d: %s", i, line))

		// if this is the error line, we also add a caret on the location and the actual error.
		if i == lfp.Line {
			// We need to create the padding taking into account tabs! Fucking tabs!
			padding := make([]rune, 0, lfp.Char)
			for _, runeValue := range line {
				if len(padding) >= lfp.Char-1 {
					break
				}

				// if it is a tab, we want to add it as is into the padding.
				// Otherwise we just add a space (we assume all other characters are 1 char wide.
				if runeValue == '\t' {
					padding = append(padding, '\t')
					continue
				}

				padding = append(padding, ' ')
			}

			// We merge the padding with the known padding we added to account for line numbers.
			output = append(output, fmt.Sprintf("%*s%s^---> %s", 6, " ", string(padding), msg))
			printedMsg = true
		}
	}

	if !printedMsg && msg != "" {
		output = append(output, msg)
	}

	if footer != "" {
		output = append(output, "", footer)
	}

	// Trim the spaces.
	for i := 0; i < len(output); i++ {
		output[i] = strings.TrimRight(output[i], " \r\n\t")
	}

	return strings.Join(output, "\n"), nil
}

func Errorf(pos *files.LoadedFilePosition, kind, format string, args ...interface{}) error {
	return &gochartError{
		Kind:          kind,
		InternalError: fmt.Errorf(format, args...),
		pos:           pos,
	}
}

func InternalErrorf(format string, args ...interface{}) error {
	originalText := fmt.Sprintf(format, args...)
	finalText := fmt.Sprintf("%s\n*** This is most likely an internal error. Please file a bug with a repro. ***", originalText)

	return fmt.Errorf(finalText)
}

type gochartErrorContainer struct {
	Errors []error
}

var _ error = (*gochartErrorContainer)(nil)

func NewErrorContainer(errors []error) *gochartErrorContainer {
	return &gochartErrorContainer{
		Errors: errors,
	}
}

func (gec *gochartErrorContainer) Error() string {
	errorStrings := make([]string, 0, len(gec.Errors))

	for _, err := range gec.Errors {
		var irerr *gochartError
		if errors.As(err, &irerr) {
			errorStrings = append(errorStrings, irerr.Render())
		} else {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	return strings.Join(errorStrings, "\n")
}

// unwrapError returns the last wrapped error in an error chain and also returns the count of errors
// within the chain.
func unwrapError(err error) (error, int) {
	count := 1
	for {
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			err = unwrapped
			fmt.Println("ERROR:", count, err)
			count++
			continue
		}

		break
	}

	return err, count
}
