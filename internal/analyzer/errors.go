package analyzer

// taskCanceledError represents an error indicating that a task was canceled.
type taskCanceledError struct {
	e error
}

// newTaskCanceledError creates a new taskCanceledError with the given error.
func newTaskCanceledError(err error) *taskCanceledError {
	return &taskCanceledError{e: err}
}

// Error returns the error message of the underlying error.
func (t *taskCanceledError) Error() string {
	return t.e.Error()
}

// Unwrap returns the underlying error.
func (t *taskCanceledError) Unwrap() error {
	return t.e
}
