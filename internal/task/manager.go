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
	"golang.org/x/term"
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

func (m *Manager) AddTask(task *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range m.tasks {
		if t.id == task.id {
			return
		}
	}

	m.tasks = append(m.tasks, task)
}

func (m *Manager) Start() {
	m.mu.Lock()

	// Do not bother outputting tasks statuses if we are not in a terminal.
	// Finish function will write the final task statuses out.
	if !m.isTerminal() {
		m.mu.Unlock()
		return
	}

	_, _ = fmt.Fprint(m.out, ansi.SaveCursorPos+ansi.HideCursor)

	m.mu.Unlock()

	go func() {
		for {
			select {
			case <-m.stopChan:
				return
			default:
				m.mu.Lock()

				var buf bytes.Buffer

				fmt.Fprint(&buf, clearTasks)

				for _, t := range m.tasks {
					fmt.Fprint(&buf, printTaskStatus(t))
				}

				fmt.Fprint(m.out, buf.String())
				m.mu.Unlock()
				time.Sleep(m.refreshRate)
			}
		}
	}()
}

func (m *Manager) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var buf bytes.Buffer

	fmt.Fprint(&buf, clearTasks)

	for _, t := range m.tasks {
		fmt.Fprint(&buf, printTaskStatus(t))
	}

	fmt.Fprint(&buf, ansi.ShowCursor)
	fmt.Fprint(m.out, buf.String())

	close(m.stopChan)
}

func (m *Manager) isTerminal() bool {
	fd := m.out.Fd()
	return term.IsTerminal(int(fd))
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
