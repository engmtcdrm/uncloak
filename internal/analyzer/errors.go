package analyzer

import "fmt"

// taskCanceledError represents an error indicating that a task was canceled.
type taskCanceledError struct {
	s string
	e error
}

// newTaskCanceledError creates a new taskCanceledError with the given error.
func newTaskCanceledError(err error) taskCanceledError {
	return taskCanceledError{
		s: "task canceled",
		e: err,
	}
}

// Error returns the error message.
func (t taskCanceledError) Error() string {
	if t.e == nil {
		return t.s
	}

	return fmt.Sprintf("%s: %v", t.s, t.e)
}

// Unwrap returns the underlying error.
func (t taskCanceledError) Unwrap() error {
	return t.e
}
