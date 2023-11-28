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
	InternalError error
	pos           *files.LoadedFilePosition
}

var _ error = (*gochartError)(nil)

func (ie *gochartError) Error() string {
	return ie.InternalError.Error()
}

func (ie *gochartError) Render() string {
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

	return ie.pos.Render(header, msg, footer, kContextSize)
}

func Errorf(pos *files.LoadedFilePosition, format string, args ...interface{}) error {
	return &gochartError{
		InternalError: fmt.Errorf(format, args...),
		pos:           pos,
	}
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

func (iec *gochartErrorContainer) Error() string {
	errorStrings := make([]string, 0, len(iec.Errors))

	for _, err := range iec.Errors {
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
