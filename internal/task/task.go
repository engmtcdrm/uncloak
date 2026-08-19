package task

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a unit of work with a name, message, status, and timing
// information.
type Task struct {
	Name    string
	Message string
	Status  Status

	id    string
	start time.Time
	end   time.Time
}

// NewTask creates a new task with the given name and message, and generates a
// unique ID for it.
func NewTask(name, message string) *Task {
	id := uuid.New().String()
	return &Task{
		id:      id,
		Name:    name,
		Message: message,
	}
}

// Duration returns the duration of the task based on its status. If the task is
// finished, warning, or error, it calculates the duration from the start to the
// end time. If the task is still running, it calculates the duration from the
// start time to the current time.
func (t *Task) Duration() time.Duration {
	switch t.Status {
	case Finished, Warning, Error:
		return t.end.Sub(t.start).Round(time.Millisecond)
	default:
		return time.Since(t.start).Round(time.Millisecond)
	}
}

// Error sets the task's status to Error and records the end time.
func (t *Task) Error() {
	t.Status = Error
	t.end = time.Now()
}

// Finish sets the task's status to Finished and records the end time.
func (t *Task) Finish() {
	t.Status = Finished
	t.end = time.Now()
}

// Start sets the task's status to Started and records the start time.
func (t *Task) Start() {
	t.Status = Started
	t.start = time.Now()
}

// Warning sets the task's status to Warning and records the end time.
func (t *Task) Warning() {
	t.Status = Warning
	t.end = time.Now()
}
