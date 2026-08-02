package analyzer

import (
	"fmt"
	"os"
	"time"

	"github.com/engmtcdrm/go-ansi"
	"github.com/google/uuid"
)

const (
	clearTasks = ansi.RestoreCursorPos + ansi.ClearFromCursorToEndScreen
)

type taskManager struct {
	out         *os.File
	tasks       []task
	refreshRate time.Duration
}

func NewTaskManager() *taskManager {
	out := os.Stdout

	return &taskManager{
		out:         out,
		tasks:       []task{},
		refreshRate: 100 * time.Millisecond,
	}
}

func (tm *taskManager) AddTask(name string, message string) string {
	id := uuid.New().String()
	tm.tasks = append(tm.tasks, task{
		id:      id,
		Name:    name,
		Message: message,
		Status:  started,
	})

	return id
}

func (tm *taskManager) UpdateTask(id string, message string, status taskStatus) {
	for i, t := range tm.tasks {
		if t.id == id {
			tm.tasks[i].Message = message
			tm.tasks[i].Status = status
			break
		}
	}
}

func (tm *taskManager) Start() {
	fmt.Fprint(tm.out, ansi.SaveCursorPos)
}

func (tm *taskManager) Update() {
	fmt.Fprint(tm.out, clearTasks)
}

func (tm *taskManager) Finish() {
	fmt.Fprint(tm.out, clearTasks)
}
