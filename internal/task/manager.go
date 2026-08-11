package task

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/engmtcdrm/go-ansi"
	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/engmtcdrm/uncloak/internal/colors"
	"golang.org/x/term"
)

const (
	taskTimeFormat = "%s %s %s\n"
)

type Manager struct {
	mu            *sync.RWMutex
	out           *os.File
	tasks         []*Task
	renderedLines int
	refreshRate   time.Duration
	stopChan      chan struct{}
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

	_, _ = fmt.Fprint(m.out, ansi.HideCursor)

	m.mu.Unlock()

	go func() {
		for {
			select {
			case <-m.stopChan:
				return
			default:
				m.mu.Lock()

				fmt.Fprint(m.out, renderTasks(m.tasks, m.renderedLines, m.terminalWidth()))
				m.renderedLines = len(m.tasks)
				m.mu.Unlock()
				time.Sleep(m.refreshRate)
			}
		}
	}()
}

func (m *Manager) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Fprint(m.out, renderTasks(m.tasks, m.renderedLines, m.terminalWidth()))
	fmt.Fprint(m.out, ansi.ShowCursor)

	close(m.stopChan)
}

func (m *Manager) isTerminal() bool {
	fd := m.out.Fd()
	return term.IsTerminal(int(fd))
}

func (m *Manager) terminalWidth() int {
	width, _, err := term.GetSize(int(m.out.Fd()))
	if err != nil {
		return 0
	}

	return width
}

func formatTaskStatus(task *Task, terminalWidth int) string {
	durationText := fmt.Sprint(pp.Bold(task.Duration()))
	statusText, styledStatus, styledDuration := styleTaskStatus(task.Status, durationText)
	message := truncateTaskMessage(task.Message, terminalWidth, statusText, durationText)

	return fmt.Sprintf(taskTimeFormat, styledStatus, message, styledDuration)
}

func renderTasks(tasks []*Task, previousLines, terminalWidth int) string {
	var buf bytes.Buffer

	if previousLines > 0 {
		fmt.Fprint(&buf, ansi.CursorUp(previousLines))
		fmt.Fprint(&buf, ansi.ClearFromCursorToEndScreen)
	}

	for _, t := range tasks {
		fmt.Fprint(&buf, formatTaskStatus(t, terminalWidth))
	}

	return buf.String()
}

func styleTaskStatus(status Status, durationText string) (statusText string, styledStatus string, styledDuration string) {
	switch status {
	case Finished:
		return "✓", pp.Bold(colors.Green("✓")), colors.Green(durationText)
	case Error:
		return "✗", pp.Bold(pp.Red("✗")), pp.Red(durationText)
	case Warning:
		return "!", pp.Bold(pp.Yellow("!")), pp.Yellow(durationText)
	default:
		return " ", " ", pp.Dim(durationText)
	}
}

func truncateTaskMessage(message string, terminalWidth int, statusText string, durationText string) string {
	if terminalWidth <= 0 {
		return message
	}

	visibleFixedWidth := utf8.RuneCountInString(statusText) + 2 + utf8.RuneCountInString(durationText)
	availableWidth := terminalWidth - visibleFixedWidth
	if availableWidth <= 0 {
		return ""
	}

	if utf8.RuneCountInString(message) <= availableWidth {
		return message
	}

	if availableWidth <= 3 {
		return strings.Repeat(".", availableWidth)
	}

	runes := []rune(message)
	return string(runes[:availableWidth-3]) + "..."
}
