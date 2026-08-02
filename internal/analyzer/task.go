package analyzer

import (
	"time"
)

type taskStatus int

const (
	started taskStatus = iota
	running
	finished
)

type task struct {
	id       string
	Name     string
	Message  string
	Status   taskStatus
	Started  time.Time
	Finished time.Time
}

func (t *task) Start() {
	t.Status = running
	t.Started = time.Now()
}

func (t *task) Finish() {
	t.Status = finished
	t.Finished = time.Now()
}

func (t *task) Duration() time.Duration {
	if t.Status == finished {
		return t.Finished.Sub(t.Started)
	}

	return time.Since(t.Started)
}
