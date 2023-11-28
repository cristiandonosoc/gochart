package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/files"
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

	return ge.pos.Render(header, msg, footer, kContextSize)
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
