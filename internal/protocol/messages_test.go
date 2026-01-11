package protocol

import (
	"encoding/json"
	"testing"
	"trees/internal/domain"
)

func TestSubscribeMessage_Marshal(t *testing.T) {
	msg := SubscribeMessage{
		Type:       "subscribe",
		ProjectKey: "abc",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SubscribeMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != "subscribe" {
		t.Errorf("expected type %q, got %q", "subscribe", decoded.Type)
	}
	if decoded.ProjectKey != "abc" {
		t.Errorf("expected projectKey %q, got %q", "abc", decoded.ProjectKey)
	}
}

func TestSubscribedMessage_Marshal(t *testing.T) {
	msg := SubscribedMessage{
		Type:       "subscribed",
		ProjectKey: "abc",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SubscribedMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != "subscribed" {
		t.Errorf("expected type %q, got %q", "subscribed", decoded.Type)
	}
	if decoded.ProjectKey != "abc" {
		t.Errorf("expected projectKey %q, got %q", "abc", decoded.ProjectKey)
	}
}

func TestPublishTreeMessage_Marshal(t *testing.T) {
	tree := domain.TaskTree{
		Root: domain.TaskNode{
			ID:    "root",
			Title: "Test",
		},
	}

	msg := PublishTreeMessage{
		Type:       "publishTree",
		ProjectKey: "abc",
		Tree:       tree,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded PublishTreeMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != "publishTree" {
		t.Errorf("expected type %q, got %q", "publishTree", decoded.Type)
	}
	if decoded.ProjectKey != "abc" {
		t.Errorf("expected projectKey %q, got %q", "abc", decoded.ProjectKey)
	}
	if decoded.Tree.Root.ID != "root" {
		t.Errorf("expected root ID %q, got %q", "root", decoded.Tree.Root.ID)
	}
}

func TestTreeAddedMessage_Marshal(t *testing.T) {
	tree := domain.TaskTree{
		Root: domain.TaskNode{
			ID:    "root",
			Title: "Test",
		},
	}

	msg := TreeAddedMessage{
		Type:       "treeAdded",
		ProjectKey: "abc",
		Tree:       tree,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded TreeAddedMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != "treeAdded" {
		t.Errorf("expected type %q, got %q", "treeAdded", decoded.Type)
	}
	if decoded.ProjectKey != "abc" {
		t.Errorf("expected projectKey %q, got %q", "abc", decoded.ProjectKey)
	}
	if decoded.Tree.Root.ID != "root" {
		t.Errorf("expected root ID %q, got %q", "root", decoded.Tree.Root.ID)
	}
}

func TestParseMessage_Subscribe(t *testing.T) {
	data := []byte(`{"type":"subscribe","projectKey":"abc"}`)

	msgType, err := ParseMessageType(data)
	if err != nil {
		t.Fatalf("failed to parse message type: %v", err)
	}

	if msgType != "subscribe" {
		t.Errorf("expected type %q, got %q", "subscribe", msgType)
	}
}

func TestParseMessage_PublishTree(t *testing.T) {
	data := []byte(`{"type":"publishTree","projectKey":"abc","tree":{"root":{"id":"root","title":"Test"}}}`)

	msgType, err := ParseMessageType(data)
	if err != nil {
		t.Fatalf("failed to parse message type: %v", err)
	}

	if msgType != "publishTree" {
		t.Errorf("expected type %q, got %q", "publishTree", msgType)
	}
}
