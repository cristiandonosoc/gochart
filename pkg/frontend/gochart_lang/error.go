package gochart_lang

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	FrontendErrorContextSize = 2
)

type frontendError struct {
	InternalError error
	Line          int
	Char          int
}

var _ error = (*frontendError)(nil)

func newFrontendError(err error, line, char int) *frontendError {
	return &frontendError{
		InternalError: err,
		Line:          line,
		Char:          char,
	}
}

func (fe *frontendError) Error() string {
	return fe.InternalError.Error()
}

func (fe *frontendError) Render(lines []string) string {
	var errorLines []string

	// Show a couple lines of context.
	begin := fe.Line - FrontendErrorContextSize
	end := fe.Line + FrontendErrorContextSize

	for i := begin; i <= end; i++ {
		// NOTE: Line as 1-based.
		arrayIndex := i - 1
		if arrayIndex < 0 {
			continue
		}

		if arrayIndex >= len(lines) {
			break
		}

		line := lines[arrayIndex]

		// We print the line.
		errorLines = append(errorLines, fmt.Sprintf("%4d: %s", i, line))

		// if this is the error line, we also add a caret on the error.
		if i == fe.Line {
			errorLines = append(errorLines, fmt.Sprintf("%*c---> %s", 6+fe.Char, '^', fe.InternalError.Error()))
		}
	}

	return strings.Join(errorLines, "\n")
}

type FrontendErrorContainer struct {
	Lines  []string
	Errors []error
}

var _ error = (*FrontendErrorContainer)(nil)

func NewFrontendErrorContainer(r io.Reader, errors []error) *FrontendErrorContainer {
	var lines []string
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if scanner.Err() != nil {
		panic(scanner.Err())
	}

	return &FrontendErrorContainer{
		Lines:  lines,
		Errors: errors,
	}
}

func (fec *FrontendErrorContainer) Error() string {
	errorStrings := make([]string, 0, len(fec.Errors))
	for _, err := range fec.Errors {
		if fe, ok := err.(*frontendError); ok {
			errorStrings = append(errorStrings, fe.Render(fec.Lines))
		} else {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	return strings.Join(errorStrings, "\n")
}
