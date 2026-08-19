package task

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Task represents a unit of work with a name, message, status, and timing
// information.
type Task struct {
	id      string
	name    string
	message string
	status  Status
	start   time.Time
	end     time.Time
	mu      sync.RWMutex
}

// NewTask creates a new task with the given name and message, and generates a
// unique ID for it.
func NewTask(name, message string) *Task {
	id := uuid.New().String()

	return &Task{
		id:      id,
		name:    name,
		message: message,
		mu:      sync.RWMutex{},
	}
}

// Duration returns the duration of the task based on its status. If the task is
// finished, warning, or error, it calculates the duration from the start to the
// end time. If the task is still running, it calculates the duration from the
// start time to the current time.
func (t *Task) Duration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.start.IsZero() {
		return 0
	}

	switch t.status {
	case Finished, Warning, Error:
		return t.end.Sub(t.start).Round(time.Millisecond)
	default:
		return time.Since(t.start).Round(time.Millisecond)
	}
}

// Error sets the task's status to Error and records the end time.
func (t *Task) Error() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status = Error
	t.end = time.Now()
}

// Finish sets the task's status to Finished and records the end time.
func (t *Task) Finish() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status = Finished
	t.end = time.Now()
}

// ID returns the unique identifier of the task.
func (t *Task) ID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.id
}

// Message returns the message of the task.
func (t *Task) Message() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.message
}

// Name returns the name of the task.
func (t *Task) Name() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.name
}

// SetMessage sets the task's message to the given value.
func (t *Task) SetMessage(message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.message = message
}

// SetName sets the task's name to the given value.
func (t *Task) SetName(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.name = name
}

// SetStatus sets the task's status to the given value.
func (t *Task) SetStatus(status Status) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status = status
}

// Start sets the task's status to Started and records the start time.
func (t *Task) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status = Started
	t.start = time.Now()
}

// Status returns the current status of the task.
func (t *Task) Status() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.status
}

// Warning sets the task's status to Warning and records the end time.
func (t *Task) Warning() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status = Warning
	t.end = time.Now()
}
