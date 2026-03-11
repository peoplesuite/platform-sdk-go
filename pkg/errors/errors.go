package errors

import "fmt"

// Error is the canonical platform error.
type Error struct {
	status  int
	message string
	kind    Kind
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Status() int {
	return e.status
}

func (e *Error) Kind() Kind {
	return e.kind
}

func (e *Error) Message() string {
	return e.message
}

func mk(status int, kind Kind, message string, cause error) *Error {
	return &Error{
		status:  status,
		kind:    kind,
		message: message,
		cause:   cause,
	}
}
