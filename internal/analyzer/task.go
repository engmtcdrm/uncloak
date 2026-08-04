package analyzer

import (
	"time"

	"github.com/google/uuid"
)

type taskStatus int

const (
	taskStarted taskStatus = iota
	taskFinished
	taskWarning
	taskError
)

type task struct {
	id       string
	Name     string
	Message  string
	Status   taskStatus
	Started  time.Time
	Finished time.Time
}

func NewTask(name, message string) *task {
	uuid := uuid.New().String()
	return &task{
		id:      uuid,
		Name:    name,
		Message: message,
	}
}

func (t *task) Start() {
	t.Status = taskStarted
	t.Started = time.Now()
}

func (t *task) Finish() {
	t.Status = taskFinished
	t.Finished = time.Now()
}

func (t *task) Duration() time.Duration {
	if t.Status == taskFinished {
		return t.Finished.Sub(t.Started).Round(time.Millisecond)
	}

	return time.Since(t.Started).Round(time.Millisecond)
}
