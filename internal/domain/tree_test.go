package domain

import (
	"testing"
)

func TestTaskNode_Creation(t *testing.T) {
	node := TaskNode{
		ID:          "task-1",
		Title:       "Test Task",
		Description: "A test task",
		Status:      "pending",
		Children:    []TaskNode{},
	}

	if node.ID != "task-1" {
		t.Errorf("expected ID %q, got %q", "task-1", node.ID)
	}
	if node.Title != "Test Task" {
		t.Errorf("expected Title %q, got %q", "Test Task", node.Title)
	}
	if node.Status != "pending" {
		t.Errorf("expected Status %q, got %q", "pending", node.Status)
	}
}

func TestTaskTree_Creation(t *testing.T) {
	tree := TaskTree{
		Root: TaskNode{
			ID:    "root",
			Title: "Root Task",
		},
	}

	if tree.Root.ID != "root" {
		t.Errorf("expected root ID %q, got %q", "root", tree.Root.ID)
	}
}

func TestTaskTree_WithChildren(t *testing.T) {
	tree := TaskTree{
		Root: TaskNode{
			ID:    "root",
			Title: "Root",
			Children: []TaskNode{
				{ID: "child-1", Title: "Child 1"},
				{ID: "child-2", Title: "Child 2"},
			},
		},
	}

	if len(tree.Root.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(tree.Root.Children))
	}
	if tree.Root.Children[0].ID != "child-1" {
		t.Errorf("expected first child ID %q, got %q", "child-1", tree.Root.Children[0].ID)
	}
}
