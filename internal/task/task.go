package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Name    string
	Message string
	Status  Status

	id    string
	start time.Time
	end   time.Time
}

func NewTask(name, message string) *Task {
	id := uuid.New().String()
	return &Task{
		id:      id,
		Name:    name,
		Message: message,
	}
}

func (t *Task) Duration() time.Duration {
	switch t.Status {
	case Finished, Warning, Error:
		return t.end.Sub(t.start).Round(time.Millisecond)
	default:
		return time.Since(t.start).Round(time.Millisecond)
	}
}

func (t *Task) Error() {
	t.Status = Error
	t.end = time.Now()
}

func (t *Task) Finish() {
	t.Status = Finished
	t.end = time.Now()
}

func (t *Task) Start() {
	t.Status = Started
	t.start = time.Now()
}

func (t *Task) Warning() {
	t.Status = Warning
	t.end = time.Now()
}
