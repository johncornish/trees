package trees

import (
	"context"
	"testing"
	"time"
)

func TestClientConnectsAndSubscribes(t *testing.T) {
	// Start a server
	server := NewServer(":0")
	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Create and start client
	runner := NewLoggingRunner()
	dispatcher := NewDispatcher(runner, 5)
	client := NewClient(server.Address(), "test-project", dispatcher)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect(ctx)
	}()

	// Give client time to connect and subscribe
	time.Sleep(100 * time.Millisecond)

	// Verify connection by publishing a tree
	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks: []TaskNode{
			{ID: "task-1", Description: "Test task"},
		},
	}

	server.PublishTree(tree)

	// Give client time to process
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop client
	cancel()

	// Wait for client to finish
	err := <-errChan
	if err != nil && err != context.Canceled {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClientDispatchesTasks(t *testing.T) {
	// Start a server
	server := NewServer(":0")
	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Create client with stub runner
	runner := NewStubRunner(10 * time.Millisecond)
	dispatcher := NewDispatcher(runner, 5)
	client := NewClient(server.Address(), "test-project", dispatcher)

	// Track execution summaries
	summaries := make(chan ExecutionSummary, 1)
	client.OnTreeReceived = func(summary ExecutionSummary) {
		summaries <- summary
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect in background
	go client.Connect(ctx)

	// Give client time to connect
	time.Sleep(100 * time.Millisecond)

	// Publish a tree with multiple tasks
	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks: []TaskNode{
			{ID: "task-1", Description: "First task"},
			{ID: "task-2", Description: "Second task"},
			{ID: "task-3", Description: "Third task"},
		},
	}

	server.PublishTree(tree)

	// Wait for execution summary
	select {
	case summary := <-summaries:
		if summary.TotalTasks != 3 {
			t.Errorf("expected 3 tasks, got %d", summary.TotalTasks)
		}
		if summary.Successes != 3 {
			t.Errorf("expected 3 successes, got %d", summary.Successes)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for execution summary")
	}

	cancel()
}

func TestClientHandlesMultipleTrees(t *testing.T) {
	// Start a server
	server := NewServer(":0")
	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Create client
	runner := NewLoggingRunner()
	dispatcher := NewDispatcher(runner, 5)
	client := NewClient(server.Address(), "test-project", dispatcher)

	// Track number of trees received
	treeCount := 0
	client.OnTreeReceived = func(summary ExecutionSummary) {
		treeCount++
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect in background
	go client.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	// Publish multiple trees
	for i := 0; i < 3; i++ {
		tree := Tree{
			ID:         "tree-" + string(rune('1'+i)),
			ProjectKey: "test-project",
			Tasks: []TaskNode{
				{ID: "task-1", Description: "Task"},
			},
		}
		server.PublishTree(tree)
		time.Sleep(50 * time.Millisecond)
	}

	// Give time to process
	time.Sleep(100 * time.Millisecond)

	if treeCount != 3 {
		t.Errorf("expected 3 trees received, got %d", treeCount)
	}

	cancel()
}

func TestClientOnlyReceivesMatchingProject(t *testing.T) {
	// Start a server
	server := NewServer(":0")
	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Create client subscribed to "project-alpha"
	runner := NewLoggingRunner()
	dispatcher := NewDispatcher(runner, 5)
	client := NewClient(server.Address(), "project-alpha", dispatcher)

	treeCount := 0
	client.OnTreeReceived = func(summary ExecutionSummary) {
		treeCount++
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go client.Connect(ctx)

	time.Sleep(100 * time.Millisecond)

	// Publish tree for "project-beta" (should NOT be received)
	tree1 := Tree{
		ID:         "tree-1",
		ProjectKey: "project-beta",
		Tasks:      []TaskNode{{ID: "task-1", Description: "Task"}},
	}
	server.PublishTree(tree1)

	time.Sleep(100 * time.Millisecond)

	// Publish tree for "project-alpha" (SHOULD be received)
	tree2 := Tree{
		ID:         "tree-2",
		ProjectKey: "project-alpha",
		Tasks:      []TaskNode{{ID: "task-2", Description: "Task"}},
	}
	server.PublishTree(tree2)

	time.Sleep(100 * time.Millisecond)

	if treeCount != 1 {
		t.Errorf("expected 1 tree received (only matching project), got %d", treeCount)
	}

	cancel()
}
