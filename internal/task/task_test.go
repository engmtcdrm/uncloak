package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests for [NewTask] function.
func Test_NewTask(t *testing.T) {
	t.Run("should create a new task with the given name and message", func(t *testing.T) {
		name := "Test Task"
		message := "This is a test task"
		task := NewTask(name, message)

		assert.Equal(t, name, task.Name, "Expected task name to match")
		assert.Equal(t, message, task.Message, "Expected task message to match")
		assert.NotEmpty(t, task.id, "Expected task ID to be generated")
	})
}

// Tests for [Task.Duration] function.
func Test_Task_Duration(t *testing.T) {
	t.Run("should return the correct duration for a finished task", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()
		time.Sleep(10 * time.Millisecond)
		task.Finish()

		duration := task.Duration()
		assert.GreaterOrEqual(t, duration.Milliseconds(), int64(10), "Expected duration to be at least 10 milliseconds")
	})

	t.Run("should return the correct duration for a running task", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()
		time.Sleep(10 * time.Millisecond)

		duration := task.Duration()
		assert.GreaterOrEqual(t, duration.Milliseconds(), int64(10), "Expected duration to be at least 10 milliseconds")
	})
}

// Tests for [Task.Error] function.
func Test_Task_Error(t *testing.T) {
	t.Run("should set the task's status to Error and record the end time", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()

		time.Sleep(10 * time.Millisecond)

		task.Error()

		assert.Equal(t, Error, task.Status, "Expected status to be Error")
		assert.False(t, task.end.IsZero(), "Expected end time to be set")
	})
}

// Tests for [Task.Finish] function.
func Test_Task_Finish(t *testing.T) {
	t.Run("should set the task's status to Finished and record the end time", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()

		time.Sleep(10 * time.Millisecond)

		task.Finish()

		assert.Equal(t, Finished, task.Status, "Expected status to be Finished")
		assert.False(t, task.end.IsZero(), "Expected end time to be set")
	})
}

// Tests for [Task.Start] function.
func Test_Task_Start(t *testing.T) {
	t.Run("should set the task's status to Started and start time to be set", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()

		time.Sleep(10 * time.Millisecond)

		assert.Equal(t, Started, task.Status, "Expected status to be Started")
		assert.False(t, task.start.IsZero(), "Expected start time to be set")
		assert.True(t, task.end.IsZero(), "Expected end time to be zero")
	})
}

// Tests for [Task.Warning] function.
func Test_Task_Warning(t *testing.T) {
	t.Run("should set the task's status to Warning and record the end time", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()

		time.Sleep(10 * time.Millisecond)

		task.Warning()

		assert.Equal(t, Warning, task.Status, "Expected status to be Warning")
		assert.False(t, task.end.IsZero(), "Expected end time to be set")
	})
}
