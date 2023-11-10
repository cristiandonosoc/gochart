package gochart_lang

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
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
	Line          int
	Char          int
}

var _ error = (*frontendError)(nil)

func newFrontendError(kind string, err error, line, char int) *frontendError {
	return &frontendError{
		Kind:          kind,
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

	// Start with an empty line.
	errorLines = append(errorLines, "", fmt.Sprintf("%s error:", fe.Kind))

	// Show a couple lines of context.
	begin := fe.Line - FrontendErrorContextSize
	end := fe.Line + FrontendErrorContextSize

	unwrappedError, errorCount := unwrapError(fe.InternalError)

	// Just in case the line handling doesn't get to print the error line.
	printedError := false

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

		// if this is the error line, we also add a caret on the location and the actual error.
		if i == fe.Line {
			errorLines = append(errorLines, fmt.Sprintf("%*c---> %s", 6+fe.Char, '^', unwrappedError.Error()))
			printedError = true
		}
	}

	if !printedError {
		errorLines = append(errorLines, fe.InternalError.Error())
	}

	// Add the full context if there was more than one wrapped error.
	if errorCount > 1 {
		errorLines = append(errorLines, "")
		errorLines = append(errorLines, fmt.Sprintf("Full context: %s", fe.InternalError.Error()))
	}

	// Trim the spaces.
	for i := 0; i < len(errorLines); i++ {
		errorLines[i] = strings.TrimRight(errorLines[i], " \r\n\t")
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
