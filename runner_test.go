package trees

import (
	"context"
	"testing"
	"time"
)

func TestStubRunnerExecutesTask(t *testing.T) {
	runner := NewStubRunner(10 * time.Millisecond)
	task := TaskNode{
		ID:          "task-1",
		Description: "Test task",
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, task)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TaskID != "task-1" {
		t.Errorf("expected TaskID 'task-1', got %q", result.TaskID)
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
}

func TestStubRunnerRespectsContext(t *testing.T) {
	runner := NewStubRunner(100 * time.Millisecond)
	task := TaskNode{
		ID:          "task-1",
		Description: "Test task",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := runner.Run(ctx, task)

	if err == nil {
		t.Error("expected context deadline exceeded error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestStubRunnerRecordsMetrics(t *testing.T) {
	runner := NewStubRunner(5 * time.Millisecond)
	task := TaskNode{
		ID:          "task-1",
		Description: "Test task",
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, task)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Duration <= 0 {
		t.Errorf("expected positive duration, got %v", result.Duration)
	}

	// Duration should be at least the sleep time
	if result.Duration < 5*time.Millisecond {
		t.Errorf("expected duration >= 5ms, got %v", result.Duration)
	}
}
