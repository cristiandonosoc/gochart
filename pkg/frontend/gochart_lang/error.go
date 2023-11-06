package gochart_lang

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
