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
	Out         *os.File
	Tasks       []*Task
	RefreshRate time.Duration

	mu            *sync.RWMutex
	renderedLines int
	stopChan      chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		Out:         os.Stdout,
		Tasks:       []*Task{},
		RefreshRate: 100 * time.Millisecond,
		stopChan:    make(chan struct{}, 1),
		mu:          &sync.RWMutex{},
	}
}

func (m *Manager) AddTask(task *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range m.Tasks {
		if t.id == task.id {
			return
		}
	}

	m.Tasks = append(m.Tasks, task)
}

func (m *Manager) Start() {
	m.mu.Lock()

	// Do not bother outputting tasks statuses if we are not in a terminal.
	// Finish function will write the final task statuses out.
	if !m.isTerminal() {
		m.mu.Unlock()
		return
	}

	fmt.Fprint(m.Out, ansi.HideCursor) //nolint:errcheck

	m.mu.Unlock()

	go func() {
		for {
			select {
			case <-m.stopChan:
				return
			default:
				m.mu.Lock()

				fmt.Fprint(m.Out, renderTasks(m.Tasks, m.renderedLines, m.terminalWidth())) //nolint:errcheck
				m.renderedLines = len(m.Tasks)
				m.mu.Unlock()
				time.Sleep(m.RefreshRate)
			}
		}
	}()
}

func (m *Manager) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Fprint(m.Out, renderTasks(m.Tasks, m.renderedLines, m.terminalWidth())) //nolint:errcheck
	fmt.Fprint(m.Out, ansi.ShowCursor)                                          //nolint:errcheck

	close(m.stopChan)
}

func (m *Manager) isTerminal() bool {
	fd := m.Out.Fd()

	return term.IsTerminal(int(fd))
}

func (m *Manager) terminalWidth() int {
	width, _, err := term.GetSize(int(m.Out.Fd()))
	if err != nil {
		return 0
	}

	return width
}

func formatTaskStatus(task *Task, terminalWidth int) string {
	styledStatus, styledDuration := styleTaskStatus(task)
	message := truncateTaskMessage([]rune(task.Message), terminalWidth, styledStatus, styledDuration)

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

func styleTaskStatus(task *Task) (styledStatus string, styledDuration string) {
	styledDuration = fmt.Sprint(pp.Bold(task.Duration()))

	switch task.Status {
	case Finished:
		return pp.Bold(colors.Green("✓")), colors.Green(styledDuration)
	case Error:
		return pp.Bold(pp.Red("✗")), pp.Red(styledDuration)
	case Warning:
		return pp.Bold(pp.Yellow("!")), pp.Yellow(styledDuration)
	default:
		return " ", pp.Dim(styledDuration)
	}
}

func truncateTaskMessage(message []rune, terminalWidth int, status string, duration string) string {
	if terminalWidth <= 0 {
		return string(message)
	}

	// Need to strip ANSI escape sequences to get proper width count.
	noANSIStatus := ansi.Strip(status)
	noANSIDuration := ansi.Strip(duration)

	const messagePadding = 2 // 1 space before and after the message.

	visibleFixedWidth := utf8.RuneCountInString(noANSIStatus) + messagePadding + utf8.RuneCountInString(noANSIDuration)
	availableWidth := terminalWidth - visibleFixedWidth
	if availableWidth <= 0 {
		return ""
	}

	if len(message) <= availableWidth {
		return string(message)
	}

	if availableWidth <= 3 {
		return strings.Repeat(".", availableWidth)
	}

	return string(message[:availableWidth-3]) + "..."
}
