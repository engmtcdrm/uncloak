package task

import (
	"os"
	"testing"

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

// Tests for [Manager.Start] function.
func Test_Manager_Start(t *testing.T) {

}

// Tests for -[Manager.isTerminal] function.
func Test_Manager_isTerminal(t *testing.T) {
	t.Run("should return true if the output is a terminal", func(t *testing.T) {
		mockPTY, mockTTY := testutils.CreatePTY(t)
		t.Cleanup(func() {
			mockPTY.Close()
			mockTTY.Close()
		})

		manager := NewManager()
		manager.Out = mockTTY
		assert.True(t, manager.isTerminal())
	})

	t.Run("should return false if the output is not a terminal", func(t *testing.T) {
		// Althought os.Stdout is used, we are technically not in a terminal here
		manager := NewManager()
		assert.False(t, manager.isTerminal())
	})
}

// Tests for [Manager.terminalWidth] function.
func Test_Manager_terminalWidth(t *testing.T) {
	t.Run("should return the width of the terminal", func(t *testing.T) {
		mockPTY, mockTTY := testutils.CreatePTYWithSize(t, 80, 24)
		t.Cleanup(func() {
			mockPTY.Close()
			mockTTY.Close()
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
