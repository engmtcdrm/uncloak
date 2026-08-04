package analyzer

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
	clearTasks = ansi.RestoreCursorPos + ansi.ClearFromCursorToEndScreen
)

type taskManager struct {
	mu          sync.RWMutex
	out         *os.File
	tasks       []*task
	refreshRate time.Duration
	stopChan    chan struct{}
}

func NewTaskManager() *taskManager {
	out := os.Stdout

	return &taskManager{
		out:         out,
		tasks:       []*task{},
		refreshRate: 100 * time.Millisecond,
		stopChan:    make(chan struct{}),
	}
}

func (tm *taskManager) AddTask(task *task) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.tasks {
		if t.id == task.id {
			return
		}
	}

	tm.tasks = append(tm.tasks, task)
}

func (tm *taskManager) Start() {
	fmt.Fprint(tm.out, ansi.SaveCursorPos)

	var buf bytes.Buffer

	taskTimeFormat := "%s [%s]\n"

	go func() {
		for {
			select {
			case <-tm.stopChan:
				return
			default:
				fmt.Fprint(&buf, clearTasks)

				for _, t := range tm.tasks {
					switch {
					case t.Status == taskFinished:
						fmt.Fprintf(&buf, taskTimeFormat, t.Message, pp.Green(t.Duration()))
						continue
					default:
						fmt.Fprintf(&buf, taskTimeFormat, t.Message, t.Duration())
					}
				}

				fmt.Fprint(tm.out, buf.String())
				buf.Reset()
				time.Sleep(tm.refreshRate)
			}
		}
	}()

}

func (tm *taskManager) Finish() {

	close(tm.stopChan)
	var buf bytes.Buffer
	// fmt.Fprint(tm.out, clearTasks)
	taskTimeFormat := "%s %s [%s]\n"

	fmt.Fprint(&buf, clearTasks)

	for _, t := range tm.tasks {
		switch {
		case t.Status == taskFinished:
			fmt.Fprintf(&buf, taskTimeFormat, colors.Green("✓"), t.Message, pp.Green(t.Duration()))
			continue
		default:
			fmt.Fprintf(&buf, taskTimeFormat, " ", t.Message, t.Duration())
		}
	}

	fmt.Fprint(tm.out, buf.String())
}

func (tm *taskManager) Stop() {
	tm.Finish()
}
