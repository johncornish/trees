package trees

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestDispatcherRunsSingleTask(t *testing.T) {
	runner := NewStubRunner(10 * time.Millisecond)
	dispatcher := NewDispatcher(runner, 5)

	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks: []TaskNode{
			{ID: "task-1", Description: "First task"},
		},
	}

	ctx := context.Background()
	summary := dispatcher.Dispatch(ctx, tree)

	if summary.TotalTasks != 1 {
		t.Errorf("expected 1 total task, got %d", summary.TotalTasks)
	}

	if summary.Successes != 1 {
		t.Errorf("expected 1 success, got %d", summary.Successes)
	}

	if summary.Failures != 0 {
		t.Errorf("expected 0 failures, got %d", summary.Failures)
	}
}

func TestDispatcherRunsMultipleTasksInParallel(t *testing.T) {
	runner := NewStubRunner(50 * time.Millisecond)
	dispatcher := NewDispatcher(runner, 10)

	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks: []TaskNode{
			{ID: "task-1", Description: "First task"},
			{ID: "task-2", Description: "Second task"},
			{ID: "task-3", Description: "Third task"},
		},
	}

	ctx := context.Background()
	start := time.Now()
	summary := dispatcher.Dispatch(ctx, tree)
	duration := time.Since(start)

	if summary.TotalTasks != 3 {
		t.Errorf("expected 3 total tasks, got %d", summary.TotalTasks)
	}

	if summary.Successes != 3 {
		t.Errorf("expected 3 successes, got %d", summary.Successes)
	}

	// If tasks run in parallel, total time should be close to 50ms, not 150ms
	// Allow some overhead
	if duration > 100*time.Millisecond {
		t.Errorf("tasks appear to have run sequentially (took %v), expected parallel execution", duration)
	}
}

func TestDispatcherRespectsConcurrencyLimit(t *testing.T) {
	// This test validates that the dispatcher respects the concurrency limit
	// We'll use a custom tracking runner to monitor concurrent execution

	var concurrentCount int32
	var maxConcurrent int32

	// trackingRunner wraps an AgentRunner and tracks concurrency
	type trackingRunner struct {
		sleepDuration time.Duration
	}

	runner := &trackingRunner{sleepDuration: 50 * time.Millisecond}

	// Implement the AgentRunner interface
	runFunc := func(ctx context.Context, task TaskNode) (TaskResult, error) {
		current := atomic.AddInt32(&concurrentCount, 1)
		defer atomic.AddInt32(&concurrentCount, -1)

		// Track max concurrent
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max {
				break
			}
			if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}

		time.Sleep(runner.sleepDuration)

		return TaskResult{
			TaskID:   task.ID,
			Success:  true,
			Duration: runner.sleepDuration,
		}, nil
	}

	// Create a mock runner
	mockRunner := &mockAgentRunner{runFunc: runFunc}

	// Create dispatcher with concurrency limit of 2
	dispatcher := NewDispatcher(mockRunner, 2)

	// Create 5 tasks
	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks: []TaskNode{
			{ID: "task-1", Description: "Task 1"},
			{ID: "task-2", Description: "Task 2"},
			{ID: "task-3", Description: "Task 3"},
			{ID: "task-4", Description: "Task 4"},
			{ID: "task-5", Description: "Task 5"},
		},
	}

	ctx := context.Background()
	summary := dispatcher.Dispatch(ctx, tree)

	if summary.TotalTasks != 5 {
		t.Errorf("expected 5 total tasks, got %d", summary.TotalTasks)
	}

	if summary.Successes != 5 {
		t.Errorf("expected 5 successes, got %d", summary.Successes)
	}

	maxObserved := atomic.LoadInt32(&maxConcurrent)
	if maxObserved > 2 {
		t.Errorf("concurrency limit violated: max concurrent was %d, expected <= 2", maxObserved)
	}

	if maxObserved < 1 {
		t.Error("expected at least 1 concurrent task")
	}
}

// mockAgentRunner is a test helper that implements AgentRunner
type mockAgentRunner struct {
	runFunc func(ctx context.Context, task TaskNode) (TaskResult, error)
}

func (m *mockAgentRunner) Run(ctx context.Context, task TaskNode) (TaskResult, error) {
	return m.runFunc(ctx, task)
}

func TestDispatcherHandlesEmptyTree(t *testing.T) {
	runner := NewStubRunner(10 * time.Millisecond)
	dispatcher := NewDispatcher(runner, 5)

	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks:      []TaskNode{},
	}

	ctx := context.Background()
	summary := dispatcher.Dispatch(ctx, tree)

	if summary.TotalTasks != 0 {
		t.Errorf("expected 0 total tasks, got %d", summary.TotalTasks)
	}

	if summary.Successes != 0 {
		t.Errorf("expected 0 successes, got %d", summary.Successes)
	}
}

func TestDispatcherRespectsContext(t *testing.T) {
	runner := NewStubRunner(100 * time.Millisecond)
	dispatcher := NewDispatcher(runner, 5)

	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks: []TaskNode{
			{ID: "task-1", Description: "First task"},
			{ID: "task-2", Description: "Second task"},
			{ID: "task-3", Description: "Third task"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	summary := dispatcher.Dispatch(ctx, tree)

	// Some tasks should fail due to context timeout
	if summary.Failures == 0 {
		t.Error("expected some failures due to context timeout")
	}

	if summary.TotalTasks != 3 {
		t.Errorf("expected 3 total tasks, got %d", summary.TotalTasks)
	}
}
