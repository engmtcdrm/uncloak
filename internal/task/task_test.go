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

		assert.Equal(t, name, task.Name(), "Expected task name to match")
		assert.Equal(t, message, task.Message(), "Expected task message to match")
		assert.NotEmpty(t, task.ID(), "Expected task ID to be generated")
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

	t.Run("should return zero duration for a task that has not started", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")

		duration := task.Duration()
		assert.Equal(t, int64(0), duration.Milliseconds(), "Expected duration to be zero for a task that has not started")
	})
}

// Tests for [Task.Error] function.
func Test_Task_Error(t *testing.T) {
	t.Run("should set the task's status to Error and record the end time", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()

		time.Sleep(10 * time.Millisecond)

		task.Error()

		assert.Equal(t, Error, task.Status(), "Expected status to be Error")
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

		assert.Equal(t, Finished, task.Status(), "Expected status to be Finished")
		assert.False(t, task.end.IsZero(), "Expected end time to be set")
	})
}

// Tests for [Task.ID] function.
func Test_Task_ID(t *testing.T) {
	t.Run("should return the unique identifier of the task", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")

		id := task.ID()
		assert.NotEmpty(t, id, "Expected ID to be non-empty")
	})
}

// Tests for [Task.Message] function.
func Test_Task_Message(t *testing.T) {
	t.Run("should return the message of the task", func(t *testing.T) {
		expectedMessage := "This is a test task"
		task := NewTask("Test Task", expectedMessage)

		message := task.Message()
		assert.Equal(t, expectedMessage, message, "Expected message to match the initial value")
	})
}

// Tests for [Task.Name] function.
func Test_Task_Name(t *testing.T) {
	t.Run("should return the name of the task", func(t *testing.T) {
		expectedName := "Test Task"
		task := NewTask(expectedName, "This is a test task")

		name := task.Name()
		assert.Equal(t, expectedName, name, "Expected name to match the initial value")
	})
}

// Tests for [Task.SetMessage] function.
func Test_Task_SetMessage(t *testing.T) {
	t.Run("should set the task's message to the given value", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		expectedUpdatedMessage := "Updated message"

		task.SetMessage(expectedUpdatedMessage)
		assert.Equal(t, expectedUpdatedMessage, task.Message(), "Expected message to be updated")
	})
}

// Tests for [Task.SetName] function.
func Test_Task_SetName(t *testing.T) {
	t.Run("should set the task's name to the given value", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		expectedUpdatedName := "Updated Task Name"

		task.SetName(expectedUpdatedName)
		assert.Equal(t, expectedUpdatedName, task.Name(), "Expected name to be updated")
	})
}

// Tests for [Task.SetStatus] function.
func Test_Task_SetStatus(t *testing.T) {
	t.Run("should set the task's status to the given value", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		expectedUpdatedStatus := Finished

		task.SetStatus(expectedUpdatedStatus)
		assert.Equal(t, expectedUpdatedStatus, task.Status(), "Expected status to be updated")
	})
}

// Tests for [Task.Start] function.
func Test_Task_Start(t *testing.T) {
	t.Run("should set the task's status to Started and start time to be set", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()

		time.Sleep(10 * time.Millisecond)

		assert.Equal(t, Started, task.Status(), "Expected status to be Started")
		assert.False(t, task.start.IsZero(), "Expected start time to be set")
		assert.True(t, task.end.IsZero(), "Expected end time to be zero")
	})
}

// Tests for [Task.Status] function.
func Test_Task_Status(t *testing.T) {
	t.Run("should return the current status of the task", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")

		expectedStatus := NotStarted
		status := task.Status()
		assert.Equal(t, expectedStatus, status, "Expected status to match the initial value")
	})
}

// Tests for [Task.Warning] function.
func Test_Task_Warning(t *testing.T) {
	t.Run("should set the task's status to Warning and record the end time", func(t *testing.T) {
		task := NewTask("Test Task", "This is a test task")
		task.Start()

		time.Sleep(10 * time.Millisecond)

		task.Warning()

		assert.Equal(t, Warning, task.Status(), "Expected status to be Warning")
		assert.False(t, task.end.IsZero(), "Expected end time to be set")
	})
}
