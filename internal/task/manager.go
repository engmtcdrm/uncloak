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

type Manager struct {
	Out         *os.File
	Tasks       []*Task
	RefreshRate time.Duration

	mu            *sync.RWMutex
	renderedLines int
	stopChan      chan struct{}
}

// NewManager creates a new instance of [Manager] with default values.
func NewManager() *Manager {
	return &Manager{
		Out:         os.Stdout,
		Tasks:       []*Task{},
		RefreshRate: 100 * time.Millisecond,
		stopChan:    make(chan struct{}, 1),
		mu:          &sync.RWMutex{},
	}
}

// AddTask adds a new task to the manager if it does not already exist.
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

// Start begins the task manager's output loop, which periodically renders the
// status of all tasks to the output. It only runs if the output is a terminal.
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

	go m.monitorTasks()
}

// Finish stops the task manager's output loop and renders the final status of
// all tasks.
func (m *Manager) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var buf bytes.Buffer

	fmt.Fprint(&buf, renderTasks(m.Tasks, m.renderedLines, m.terminalWidth()))
	fmt.Fprint(&buf, ansi.ShowCursor)

	fmt.Fprint(m.Out, buf.String()) //nolint:errcheck

	close(m.stopChan)
}

// isTerminal checks if [Manager.Out] is a terminal.
func (m *Manager) isTerminal() bool {
	fd := m.Out.Fd()

	return term.IsTerminal(int(fd))
}

// monitorTasks continuously monitors the tasks and updates the terminal output
// with their current status until [Manager.stopChan] is closed.
func (m *Manager) monitorTasks() {
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
}

// terminalWidth returns the width of the terminal in characters. If the output
// is not a terminal, it returns 0.
func (m *Manager) terminalWidth() int {
	width, _, err := term.GetSize(int(m.Out.Fd()))
	if err != nil {
		return 0
	}

	return width
}

// formatTaskStatus formats the status of a task for display in the terminal.
func formatTaskStatus(task *Task, terminalWidth int) string {
	styledStatus, styledDuration := styleTaskStatus(task)
	message := truncateTaskMessage([]rune(task.Message), terminalWidth, styledStatus, styledDuration)

	return fmt.Sprintf("%s %s %s\n", styledStatus, message, styledDuration)
}

// renderTasks generates the string representation of all tasks, including
// cursor movement and clearing commands to update the terminal display.
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

// styleTaskStatus returns the styled status and duration of a task based on its
// current status.
func styleTaskStatus(task *Task) (styledStatus string, styledDuration string) {
	styledDuration = pp.Bold(task.Duration())

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

// truncateTaskMessage truncates the task message to fit within the available
// terminal width.
func truncateTaskMessage(message []rune, terminalWidth int, status string, duration string) string {
	if terminalWidth <= 0 {
		return string(message)
	}

	// Need to strip ANSI escape sequences to get proper width count.
	noANSIStatus := ansi.Strip(status)
	noANSIDuration := ansi.Strip(duration)

	const messagePadding = 2 // 1 space before and after the message.
	const ellipsis = "..."

	visibleFixedWidth := utf8.RuneCountInString(noANSIStatus) + messagePadding + utf8.RuneCountInString(noANSIDuration)
	availableWidth := terminalWidth - visibleFixedWidth
	if availableWidth <= 0 {
		return ""
	}

	if len(message) <= availableWidth {
		return string(message)
	}

	ellipseLength := len(ellipsis)

	if availableWidth <= ellipseLength {
		return strings.Repeat(".", availableWidth)
	}

	return string(message[:availableWidth-messagePadding-ellipseLength]) + ellipsis
}
