package trees

import (
	"context"
	"fmt"
	"log"
	"time"
)

// TaskResult represents the outcome of running a task
type TaskResult struct {
	TaskID   string
	Success  bool
	Error    error
	Duration time.Duration
}

// AgentRunner is the interface for executing tasks
// This is the pluggable interface where "real agent launching" can be implemented later
type AgentRunner interface {
	Run(ctx context.Context, task TaskNode) (TaskResult, error)
}

// StubRunner is a simple implementation that simulates work
type StubRunner struct {
	sleepDuration time.Duration
}

// NewStubRunner creates a new stub runner with the specified sleep duration
func NewStubRunner(sleepDuration time.Duration) *StubRunner {
	return &StubRunner{
		sleepDuration: sleepDuration,
	}
}

// Run executes a task by simulating work (sleeping)
func (r *StubRunner) Run(ctx context.Context, task TaskNode) (TaskResult, error) {
	start := time.Now()

	log.Printf("[RUNNER] Starting task %s: %s", task.ID, task.Description)

	// Simulate work with a sleep that respects context cancellation
	select {
	case <-time.After(r.sleepDuration):
		// Work completed
		duration := time.Since(start)
		log.Printf("[RUNNER] Completed task %s in %v", task.ID, duration)

		return TaskResult{
			TaskID:   task.ID,
			Success:  true,
			Duration: duration,
		}, nil

	case <-ctx.Done():
		// Context cancelled or deadline exceeded
		duration := time.Since(start)
		log.Printf("[RUNNER] Task %s cancelled after %v", task.ID, duration)

		return TaskResult{
			TaskID:   task.ID,
			Success:  false,
			Error:    ctx.Err(),
			Duration: duration,
		}, ctx.Err()
	}
}

// LoggingRunner is an even simpler implementation that just logs immediately
type LoggingRunner struct{}

// NewLoggingRunner creates a new logging runner
func NewLoggingRunner() *LoggingRunner {
	return &LoggingRunner{}
}

// Run executes a task by logging it immediately
func (r *LoggingRunner) Run(ctx context.Context, task TaskNode) (TaskResult, error) {
	start := time.Now()

	// Check if context is already done
	select {
	case <-ctx.Done():
		return TaskResult{
			TaskID:   task.ID,
			Success:  false,
			Error:    ctx.Err(),
			Duration: time.Since(start),
		}, ctx.Err()
	default:
	}

	log.Printf("[RUNNER] Task %s: %s (dependencies: %v)", task.ID, task.Description, task.Dependencies)

	duration := time.Since(start)
	return TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Duration: duration,
	}, nil
}

// ExecutionSummary tracks the results of running multiple tasks
type ExecutionSummary struct {
	TotalTasks int
	Successes  int
	Failures   int
	Duration   time.Duration
	Results    []TaskResult
}

// String formats the summary for display
func (s *ExecutionSummary) String() string {
	return fmt.Sprintf(
		"Execution Summary:\n"+
			"  Total Tasks: %d\n"+
			"  Successes: %d\n"+
			"  Failures: %d\n"+
			"  Duration: %v\n",
		s.TotalTasks,
		s.Successes,
		s.Failures,
		s.Duration,
	)
}
