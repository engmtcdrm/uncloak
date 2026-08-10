package task

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/engmtcdrm/go-ansi"
	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/engmtcdrm/uncloak/internal/colors"
)

const (
	taskTimeFormat = "%s %s %s\n"
	clearTasks     = ansi.RestoreCursorPos + ansi.ClearFromCursorToEndScreen
)

type Manager struct {
	mu          *sync.RWMutex
	out         *os.File
	tasks       []*Task
	refreshRate time.Duration
	active      bool
	stopChan    chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		out:         os.Stdout,
		tasks:       []*Task{},
		refreshRate: 100 * time.Millisecond,
		stopChan:    make(chan struct{}, 1),
		mu:          &sync.RWMutex{},
	}
}

func (tm *Manager) AddTask(task *Task) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.tasks {
		if t.id == task.id {
			return
		}
	}

	tm.tasks = append(tm.tasks, task)
}

func (tm *Manager) Start() {
	tm.mu.Lock()
	tm.active = true
	_, _ = fmt.Fprint(tm.out, ansi.SaveCursorPos+ansi.HideCursor)

	tm.mu.Unlock()

	go func() {
		for {
			select {
			case <-tm.stopChan:
				return
			default:
				tm.mu.Lock()

				if !tm.active {
					tm.mu.Unlock()
					return
				}

				var buf bytes.Buffer

				fmt.Fprint(&buf, clearTasks)

				for _, t := range tm.tasks {
					fmt.Fprint(&buf, printTaskStatus(t))
				}

				fmt.Fprint(tm.out, buf.String())
				tm.mu.Unlock()
				time.Sleep(tm.refreshRate)
			}
		}
	}()
}

func (tm *Manager) Finish() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.active = false

	var buf bytes.Buffer

	fmt.Fprint(&buf, clearTasks)

	for _, t := range tm.tasks {
		fmt.Fprint(&buf, printTaskStatus(t))
	}

	fmt.Fprint(&buf, ansi.ShowCursor)
	fmt.Fprint(tm.out, buf.String())

	close(tm.stopChan)
}

func printTaskStatus(task *Task) string {
	duration := pp.Bold(task.Duration())

	switch task.Status {
	case Finished:
		return fmt.Sprintf(taskTimeFormat, pp.Bold(colors.Green("✓")), task.Message, colors.Green(duration))
	case Error:
		return fmt.Sprintf(taskTimeFormat, pp.Bold(pp.Red("✗")), task.Message, pp.Red(duration))
	case Warning:
		return fmt.Sprintf(taskTimeFormat, pp.Bold(pp.Yellow("!")), task.Message, pp.Yellow(duration))
	default:
		return fmt.Sprintf(taskTimeFormat, " ", task.Message, pp.Bold(pp.Dim(duration)))
	}
}
