package task

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/engmtcdrm/go-ansi"
	pp "github.com/engmtcdrm/go-prettyprint"
	"github.com/engmtcdrm/uncloak/internal/colors"
	"github.com/engmtcdrm/uncloak/internal/testing/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for [NewManager] function.
func Test_NewManager(t *testing.T) {
	t.Run("should create a new Manager instance with default values", func(t *testing.T) {
		manager := NewManager()
		assert.Equal(t, os.Stdout, manager.Out)
	})
}

// Tests for [Manager.AddTask] function.
func Test_Manager_AddTask(t *testing.T) {
	t.Run("should add a new task to the manager if it does not already exist", func(t *testing.T) {
		manager := NewManager()
		task := &Task{id: "task1"}
		manager.AddTask(task)

		require.Len(t, manager.Tasks, 1)
		assert.Equal(t, task, manager.Tasks[0])
	})

	t.Run("should not add a task if it already exists in the manager", func(t *testing.T) {
		manager := NewManager()
		task := &Task{id: "task1"}
		manager.AddTask(task)
		manager.AddTask(task)

		require.Len(t, manager.Tasks, 1)
		assert.Equal(t, task, manager.Tasks[0])
	})
}

// Tests for [Manager.Finish] function.
func Test_Manager_Finish(t *testing.T) {
	t.Run("should start manager and stop when Finish is called", func(t *testing.T) {
		t.Parallel()

		master, slave := testutils.CreatePTYWithSize(t, 80, 30)
		manager := NewManager()
		manager.Out = slave

		task1 := NewTask("task1", "Test task")
		manager.AddTask(task1)
		task1.Start()

		task2 := NewTask("task2", "Test task")
		manager.AddTask(task2)
		task2.Start()

		terminalCh := make(chan struct {
			contents string
		}, 1)

		// Setup a goroutine to close the stop channel after a short delay so
		// that monitorTasks will both exit and have time to run at least once.
		go func() {
			time.Sleep(200 * time.Millisecond)
			manager.Finish()
			_ = slave.Close()
			contents := testutils.ReadPTYOutput(t, master, 1024)

			terminalCh <- struct {
				contents string
			}{contents}
		}()

		manager.Start()

		terminalOutput := <-terminalCh
		assert.Contains(t, terminalOutput.contents, "Test task")
	})
}

// Tests for [Manager.Start] function.
func Test_Manager_Start(t *testing.T) {
	t.Run("should start manager and stop when stop channel is closed", func(t *testing.T) {
		t.Parallel()

		master, slave := testutils.CreatePTYWithSize(t, 80, 30)
		manager := NewManager()
		manager.Out = slave

		task1 := NewTask("task1", "Test task")
		manager.AddTask(task1)
		task1.Start()

		task2 := NewTask("task2", "Test task")
		manager.AddTask(task2)
		task2.Start()

		terminalCh := make(chan struct {
			contents string
		}, 1)

		// Setup a goroutine to close the stop channel after a short delay so
		// that monitorTasks will both exit and have time to run at least once.
		go func() {
			time.Sleep(200 * time.Millisecond)
			close(manager.stopChan)
			_ = slave.Close()
			contents := testutils.ReadPTYOutput(t, master, 1024)

			terminalCh <- struct {
				contents string
			}{contents}
		}()

		manager.Start()

		terminalOutput := <-terminalCh
		assert.Contains(t, terminalOutput.contents, "Test task")
	})

	t.Run("should not start a new monitor if manager has already started", func(t *testing.T) {
		t.Parallel()

		master, slave := testutils.CreatePTYWithSize(t, 80, 30)
		manager := NewManager()
		manager.Out = slave

		task1 := NewTask("task1", "Test task")
		manager.AddTask(task1)
		task1.Start()

		task2 := NewTask("task2", "Test task")
		manager.AddTask(task2)
		task2.Start()

		terminalCh := make(chan struct {
			contents string
		}, 1)

		// Setup a goroutine to close the stop channel after a short delay so
		// that monitorTasks will both exit and have time to run at least once.
		go func() {
			time.Sleep(200 * time.Millisecond)
			close(manager.stopChan)
			_ = slave.Close()
			contents := testutils.ReadPTYOutput(t, master, 1024)

			terminalCh <- struct {
				contents string
			}{contents}
		}()

		manager.Start()
		manager.Start() // Attempt to start the manager again, should not start a new monitor.

		terminalOutput := <-terminalCh
		assert.Contains(t, terminalOutput.contents, "Test task")
	})

	t.Run("should start manager, but not write out tasks if the output is not a terminal", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)
		manager := NewManager()
		manager.Out = stdoutFile

		task1 := NewTask("task1", "Test task")
		manager.AddTask(task1)
		task1.Start()

		task2 := NewTask("task2", "Test task")
		manager.AddTask(task2)
		task2.Start()

		manager.Start()

		contents, err := os.ReadFile(stdoutFile.Name())
		require.NoError(t, err)
		assert.Equal(t, "", string(contents))
	})
}

// Tests for -[Manager.isTerminal] function.
func Test_Manager_isTerminal(t *testing.T) {
	t.Run("should return true if the output is a terminal", func(t *testing.T) {
		_, mockTTY := testutils.CreatePTY(t)

		manager := NewManager()
		manager.Out = mockTTY
		assert.True(t, manager.isTerminal())
	})

	t.Run("should return false if the output is not a terminal", func(t *testing.T) {
		// Although os.Stdout is used, we are technically not in a terminal here
		manager := NewManager()
		assert.False(t, manager.isTerminal())
	})
}

// Tests for [Manager.monitorTasks] function.
func Test_Manager_monitorTasks(t *testing.T) {
	t.Run("should monitor tasks and return when stop channel is closed", func(t *testing.T) {
		stdoutFile := testutils.SetStdout(t)
		manager := NewManager()
		manager.Out = stdoutFile

		task1 := NewTask("task1", "Test task")
		manager.AddTask(task1)
		task1.Start()

		task2 := NewTask("task2", "Test task")
		manager.AddTask(task2)
		task2.Start()

		terminalCh := make(chan struct {
			contents string
		}, 1)

		// Setup a goroutine to close the stop channel after a short delay so
		// that monitorTasks will both exit and have time to run at least once.
		go func() {
			time.Sleep(200 * time.Millisecond)
			close(manager.stopChan)

			contents, err := os.ReadFile(stdoutFile.Name())
			require.NoError(t, err)

			terminalCh <- struct {
				contents string
			}{string(contents)}
		}()
		manager.monitorTasks()

		terminalOutput := <-terminalCh
		assert.Contains(t, terminalOutput.contents, "Test task")
	})
}

// Tests for [Manager.terminalWidth] function.
func Test_Manager_terminalWidth(t *testing.T) {
	t.Run("should return the width of the terminal", func(t *testing.T) {
		mockPTY, mockTTY := testutils.CreatePTYWithSize(t, 80, 24)
		t.Cleanup(func() {
			require.NoError(t, mockPTY.Close(), "Failed to close master pty")
			require.NoError(t, mockTTY.Close(), "Failed to close slave pty")
		})

		manager := NewManager()
		manager.Out = mockTTY
		assert.Equal(t, 80, manager.terminalWidth())
	})

	t.Run("should return a 0 width if the output is not a terminal", func(t *testing.T) {
		manager := NewManager()
		assert.Equal(t, 0, manager.terminalWidth())
	})
}

// Tests for [formatTaskStatus] function.
func Test_formatTaskStatus(t *testing.T) {
	t.Run("should format the task status correctly when terminal has enough width", func(t *testing.T) {
		task := NewTask("task1", "Test task")
		task.Start()
		task.Finish()

		expectedStatus := fmt.Sprintf("%s %s %s\n", pp.Bold(colors.Green("✓")), "Test task", colors.Green(pp.Bold(task.Duration())))
		formattedStatus := formatTaskStatus(task, 80)
		assert.Equal(t, expectedStatus, formattedStatus)
	})

	t.Run("should truncate the task message if it exceeds the terminal width", func(t *testing.T) {
		task := NewTask("task1", "This is a very long test task message that should be truncated")
		task.Start()
		task.Finish()

		formattedStatus := formatTaskStatus(task, 30)
		strippedFormattedStatus := strings.TrimSpace(ansi.Strip(formattedStatus))
		assert.LessOrEqual(t, len(strippedFormattedStatus), 30)
	})
}

// Tests for [renderTasks] function.
func Test_renderTasks(t *testing.T) {
	t.Run("should return empty string if no previous lines or tasks are provided", func(t *testing.T) {
		rendered := renderTasks([]*Task{}, 0, 80)
		assert.Equal(t, "", rendered)
	})

	t.Run("should render tasks correctly with previous lines and terminal width", func(t *testing.T) {
		task1 := NewTask("task1", "Test task 1")
		task1.Start()
		task1.Finish()

		task2 := NewTask("task2", "Test task 2")
		task2.Start()
		task2.Finish()

		tasks := []*Task{task1, task2}
		rendered := renderTasks(tasks, 2, 80)

		expectedStatus1 := fmt.Sprintf("%s %s %s\n", pp.Bold(colors.Green("✓")), "Test task 1", colors.Green(pp.Bold(task1.Duration())))
		expectedStatus2 := fmt.Sprintf("%s %s %s\n", pp.Bold(colors.Green("✓")), "Test task 2", colors.Green(pp.Bold(task2.Duration())))
		expectedRendered := ansi.CursorUp(2) + ansi.ClearFromCursorToEndScreen + expectedStatus1 + expectedStatus2

		assert.Equal(t, expectedRendered, rendered)
	})
}

// Tests for [styleTaskStatus] function.
func Test_styleTaskStatus(t *testing.T) {
	t.Run("should return default when status is started", func(t *testing.T) {
		task := NewTask("task1", "Test task")
		task.Start()

		// Purposely finishing and resetting to Start so we can compare styledDuration output.
		task.Finish()
		task.Status = Started

		expectedStatus := " "
		expectedDuration := pp.Dim(pp.Bold(task.Duration()))
		styledStatus, styledDuration := styleTaskStatus(task)
		assert.Equal(t, expectedStatus, styledStatus)
		assert.Equal(t, styledDuration, expectedDuration)
	})

	t.Run("should return checkmark and green when status is finished", func(t *testing.T) {
		task := NewTask("task1", "Test task")
		task.Start()
		task.Finish()

		expectedStatus := pp.Bold(colors.Green("✓"))
		expectedDuration := colors.Green(pp.Bold(task.Duration()))
		styledStatus, styledDuration := styleTaskStatus(task)
		assert.Equal(t, styledStatus, expectedStatus)
		assert.Equal(t, styledDuration, expectedDuration)
	})

	t.Run("should return cross and red when status is error", func(t *testing.T) {
		task := NewTask("task1", "Test task")
		task.Start()
		task.Error()

		expectedStatus := pp.Bold(pp.Red("✗"))
		expectedDuration := pp.Red(pp.Bold(task.Duration()))
		styledStatus, styledDuration := styleTaskStatus(task)
		assert.Equal(t, styledStatus, expectedStatus)
		assert.Equal(t, styledDuration, expectedDuration)
	})

	t.Run("should return exclamation and yellow when status is warning", func(t *testing.T) {
		task := NewTask("task1", "Test task")
		task.Start()
		task.Warning()

		expectedStatus := pp.Bold(pp.Yellow("!"))
		expectedDuration := pp.Yellow(pp.Bold(task.Duration()))
		styledStatus, styledDuration := styleTaskStatus(task)
		assert.Equal(t, styledStatus, expectedStatus)
		assert.Equal(t, styledDuration, expectedDuration)
	})
}

// Tests for [truncateTaskMessage] function.
func Test_truncateTaskMessage(t *testing.T) {
	// Add ANSI escape sequences so we can make sure they are stripped out
	status := pp.Dim(" ")
	duration := pp.Dim("00:00:01")

	t.Run("should return the original message if terminalWidth is less than or equal to 0", func(t *testing.T) {
		const expectedMessage = "This is a test message"
		const terminalWidth = 0

		truncatedMessage := truncateTaskMessage([]rune(expectedMessage), terminalWidth, status, duration)
		assert.Equal(t, expectedMessage, truncatedMessage)
	})

	t.Run("should truncate the message if it exceeds the available width", func(t *testing.T) {
		const expectedMessage = "This is a test message that exceeds the available width"
		const terminalWidth = 30

		truncatedMessage := truncateTaskMessage([]rune(expectedMessage), terminalWidth, status, duration)
		assert.LessOrEqual(t, len(truncatedMessage), terminalWidth)
	})

	t.Run("should return the original message if it fits within the available width", func(t *testing.T) {
		const expectedMessage = "Short message"
		const terminalWidth = 30

		truncatedMessage := truncateTaskMessage([]rune(expectedMessage), terminalWidth, status, duration)
		assert.Equal(t, expectedMessage, truncatedMessage)
	})

	t.Run("should return empty string if available width is less than or equal to 0", func(t *testing.T) {
		const expectedMessage = "This is a test message"
		const terminalWidth = 10

		truncatedMessage := truncateTaskMessage([]rune(expectedMessage), terminalWidth, status, duration)
		assert.Equal(t, "", truncatedMessage)
	})

	t.Run("should return ellipsis if available width is less than or equal to 3", func(t *testing.T) {
		const expectedMessage = "This is a test message"
		const terminalWidth = 14

		truncatedMessage := truncateTaskMessage([]rune(expectedMessage), terminalWidth, status, duration)
		assert.Equal(t, "...", truncatedMessage)
	})
}
