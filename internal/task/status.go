package task

// Status represents the current state of a task in the task manager.
type Status int

const (
	// NotStarted represents a task that has not started yet.
	NotStarted Status = iota

	// Started represents a task that has started.
	Started

	// Finished represents a task that has finished successfully.
	Finished

	// Warning represents a task that has finished with a warning.
	Warning

	// Error represents a task that has finished with an error.
	Error
)
