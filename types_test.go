package trees

import (
	"encoding/json"
	"testing"
)

func TestTaskNodeUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "task-1",
		"description": "Build feature",
		"dependencies": ["task-0"]
	}`

	var task TaskNode
	err := json.Unmarshal([]byte(jsonData), &task)

	if err != nil {
		t.Fatalf("failed to unmarshal task: %v", err)
	}

	if task.ID != "task-1" {
		t.Errorf("expected ID 'task-1', got %q", task.ID)
	}

	if task.Description != "Build feature" {
		t.Errorf("expected description 'Build feature', got %q", task.Description)
	}

	if len(task.Dependencies) != 1 || task.Dependencies[0] != "task-0" {
		t.Errorf("expected dependencies ['task-0'], got %v", task.Dependencies)
	}
}

func TestTreeUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "tree-123",
		"projectKey": "project-alpha",
		"tasks": [
			{
				"id": "task-1",
				"description": "First task",
				"dependencies": []
			},
			{
				"id": "task-2",
				"description": "Second task",
				"dependencies": ["task-1"]
			}
		]
	}`

	var tree Tree
	err := json.Unmarshal([]byte(jsonData), &tree)

	if err != nil {
		t.Fatalf("failed to unmarshal tree: %v", err)
	}

	if tree.ID != "tree-123" {
		t.Errorf("expected ID 'tree-123', got %q", tree.ID)
	}

	if tree.ProjectKey != "project-alpha" {
		t.Errorf("expected projectKey 'project-alpha', got %q", tree.ProjectKey)
	}

	if len(tree.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tree.Tasks))
	}
}

func TestSubscribeMessageMarshal(t *testing.T) {
	msg := Message{
		Type:       "subscribe",
		ProjectKey: "project-alpha",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	expected := `{"type":"subscribe","projectKey":"project-alpha"}`
	if string(data) != expected {
		t.Errorf("expected JSON %q, got %q", expected, string(data))
	}
}

func TestTreeAddedMessageUnmarshal(t *testing.T) {
	jsonData := `{
		"type": "treeAdded",
		"projectKey": "project-alpha",
		"tree": {
			"id": "tree-123",
			"projectKey": "project-alpha",
			"tasks": []
		}
	}`

	var msg Message
	err := json.Unmarshal([]byte(jsonData), &msg)

	if err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if msg.Type != "treeAdded" {
		t.Errorf("expected type 'treeAdded', got %q", msg.Type)
	}

	if msg.ProjectKey != "project-alpha" {
		t.Errorf("expected projectKey 'project-alpha', got %q", msg.ProjectKey)
	}

	if msg.Tree == nil {
		t.Fatal("expected tree to be non-nil")
	}

	if msg.Tree.ID != "tree-123" {
		t.Errorf("expected tree ID 'tree-123', got %q", msg.Tree.ID)
	}
}
