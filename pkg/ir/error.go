package ir

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/files"
)

const (
	IrErrorContextSize = 2
)

type irError struct {
	InternalError error
	pos           *files.LoadedFilePosition
}

var _ error = (*irError)(nil)

func (ie *irError) Error() string {
	return ie.InternalError.Error()
}

func (ie *irError) Render() string {
	if ie.pos == nil {
		lines := []string{
			"<no file position found>",
			ie.InternalError.Error(),
		}
		return strings.Join(lines, "\n")
	}

	unwrappedError, errorCount := unwrapError(ie.InternalError)

	header := fmt.Sprintf("semantic error:")
	msg := unwrappedError.Error()
	footer := ""
	if errorCount > 1 {
		footer = fmt.Sprintf("Full context: %s", ie.InternalError.Error())
	}

	return ie.pos.Render(header, msg, footer, IrErrorContextSize)
}

func irErrorf(pos *files.LoadedFilePosition, format string, args ...interface{}) error {
	return &irError{
		InternalError: fmt.Errorf(format, args...),
		pos:           pos,
	}
}

type irErrorContainer struct {
	Errors []error
}

var _ error = (*irErrorContainer)(nil)

func newIrErrorContainer(errors []error) *irErrorContainer {
	return &irErrorContainer{
		Errors: errors,
	}
}

func (iec *irErrorContainer) Error() string {
	errorStrings := make([]string, 0, len(iec.Errors))

	for _, err := range iec.Errors {
		var irerr *irError
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
// TODO(cdc): This function is repeated with frontend. We should have a common error library.
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
