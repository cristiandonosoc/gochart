package gochart_lang

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cristiandonosoc/gochart/internal/files"
)

const (
	FrontendErrorContextSize = 2
)

func internalErrorf(format string, args ...interface{}) error {
	originalText := fmt.Sprintf(format, args...)
	finalText := fmt.Sprintf("%s\n*** This is most likely an internal error. Please file a bug with a repro. ***", originalText)

	return fmt.Errorf(finalText)
}

type frontendError struct {
	Kind          string
	InternalError error
	pos           *files.LoadedFilePosition
}

var _ error = (*frontendError)(nil)

func newFrontendError(kind string, err error, pos *files.LoadedFilePosition) *frontendError {
	return &frontendError{
		Kind:          kind,
		InternalError: err,
		pos:           pos,
	}
}

func (fe *frontendError) Error() string {
	return fe.InternalError.Error()
}

func (fe *frontendError) Render(lines []string) string {
	if fe.pos == nil {
		lines := []string{
			"<no file position found>",
			fe.InternalError.Error(),
		}
		return strings.Join(lines, "\n")
	}

	unwrappedError, errorCount := unwrapError(fe.InternalError)

	header := fmt.Sprintf("%s error:", fe.Kind)
	msg := unwrappedError.Error()
	footer := ""
	if errorCount > 1 {
		footer = fmt.Sprintf("Full context: %s", fe.InternalError.Error())
	}

	return fe.pos.Render(header, msg, footer, FrontendErrorContextSize)
}

type FrontendErrorContainer struct {
	Errors []error
	lf     *files.LoadedFile
}

var _ error = (*FrontendErrorContainer)(nil)

func NewFrontendErrorContainer(lf *files.LoadedFile, errors []error) *FrontendErrorContainer {
	return &FrontendErrorContainer{
		Errors: errors,
		lf:     lf,
	}
}

func (fec *FrontendErrorContainer) Error() string {
	errorStrings := make([]string, 0, len(fec.Errors))
	for _, err := range fec.Errors {
		if fe, ok := err.(*frontendError); ok {
			errorStrings = append(errorStrings, fe.Render(fec.lf.Lines))
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

func tokenToPosition(lf *files.LoadedFile, token *Token) *files.LoadedFilePosition {
	if token == nil {
		return nil
	}

	return newLoadedFilePosition(lf, token.line, token.char)
}

func newLoadedFilePosition(lf *files.LoadedFile, line, char int) *files.LoadedFilePosition {
	return &files.LoadedFilePosition{
		File: lf,
		Line: line,
		Char: char,
	}
}
