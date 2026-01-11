package server

import (
	"encoding/json"
	"sync"
	"testing"
	"trees/internal/domain"
	"trees/internal/protocol"
	"trees/internal/store"
)

// MockConnection implements Connection interface for testing.
type MockConnection struct {
	mu       sync.Mutex
	messages []string
	closed   bool
}

func (m *MockConnection) ReadLine() (string, error) {
	return "", nil
}

func (m *MockConnection) WriteLine(line string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, line)
	return nil
}

func (m *MockConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockConnection) GetMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.messages))
	copy(result, m.messages)
	return result
}

func (m *MockConnection) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestServer_Subscribe(t *testing.T) {
	s := NewServer(store.NewStore())
	conn := &MockConnection{}

	msg := protocol.SubscribeMessage{
		Type:       "subscribe",
		ProjectKey: "project1",
	}
	data, _ := json.Marshal(msg)

	if err := s.HandleMessage(conn, string(data)); err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	messages := conn.GetMessages()
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	var response protocol.SubscribedMessage
	if err := json.Unmarshal([]byte(messages[0]), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Type != "subscribed" {
		t.Errorf("expected type %q, got %q", "subscribed", response.Type)
	}
	if response.ProjectKey != "project1" {
		t.Errorf("expected projectKey %q, got %q", "project1", response.ProjectKey)
	}

	// Verify subscription was registered
	subs := s.GetSubscribers("project1")
	if len(subs) != 1 {
		t.Errorf("expected 1 subscriber, got %d", len(subs))
	}
}

func TestServer_PublishTree(t *testing.T) {
	s := NewServer(store.NewStore())

	tree := domain.TaskTree{
		Root: domain.TaskNode{
			ID:    "root",
			Title: "Test Task",
		},
	}

	msg := protocol.PublishTreeMessage{
		Type:       "publishTree",
		ProjectKey: "project1",
		Tree:       tree,
	}
	data, _ := json.Marshal(msg)

	conn := &MockConnection{}
	if err := s.HandleMessage(conn, string(data)); err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// Verify tree was stored
	storedTree, exists := s.store.Get("project1")
	if !exists {
		t.Error("expected tree to be stored")
	}
	if storedTree.Root.ID != "root" {
		t.Errorf("expected root ID %q, got %q", "root", storedTree.Root.ID)
	}
}

func TestServer_PublishAndBroadcast(t *testing.T) {
	s := NewServer(store.NewStore())

	// Create two subscribers
	sub1 := &MockConnection{}
	sub2 := &MockConnection{}

	// Subscribe both clients
	subMsg := protocol.SubscribeMessage{
		Type:       "subscribe",
		ProjectKey: "project1",
	}
	subData, _ := json.Marshal(subMsg)

	s.HandleMessage(sub1, string(subData))
	s.HandleMessage(sub2, string(subData))

	// Clear the subscription confirmation messages
	sub1.messages = nil
	sub2.messages = nil

	// Publish a tree
	tree := domain.TaskTree{
		Root: domain.TaskNode{
			ID:    "root",
			Title: "Broadcast Test",
		},
	}

	pubMsg := protocol.PublishTreeMessage{
		Type:       "publishTree",
		ProjectKey: "project1",
		Tree:       tree,
	}
	pubData, _ := json.Marshal(pubMsg)

	publisher := &MockConnection{}
	if err := s.HandleMessage(publisher, string(pubData)); err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// Verify both subscribers received the treeAdded message
	sub1Messages := sub1.GetMessages()
	sub2Messages := sub2.GetMessages()

	if len(sub1Messages) != 1 {
		t.Errorf("subscriber 1 expected 1 message, got %d", len(sub1Messages))
	}
	if len(sub2Messages) != 1 {
		t.Errorf("subscriber 2 expected 1 message, got %d", len(sub2Messages))
	}

	// Verify message content
	var broadcast protocol.TreeAddedMessage
	if err := json.Unmarshal([]byte(sub1Messages[0]), &broadcast); err != nil {
		t.Fatalf("failed to unmarshal broadcast: %v", err)
	}

	if broadcast.Type != "treeAdded" {
		t.Errorf("expected type %q, got %q", "treeAdded", broadcast.Type)
	}
	if broadcast.ProjectKey != "project1" {
		t.Errorf("expected projectKey %q, got %q", "project1", broadcast.ProjectKey)
	}
	if broadcast.Tree.Root.Title != "Broadcast Test" {
		t.Errorf("expected title %q, got %q", "Broadcast Test", broadcast.Tree.Root.Title)
	}
}

func TestServer_IsolatedProjects(t *testing.T) {
	s := NewServer(store.NewStore())

	// Subscribe to different projects
	sub1 := &MockConnection{}
	sub2 := &MockConnection{}

	subMsg1 := protocol.SubscribeMessage{Type: "subscribe", ProjectKey: "project1"}
	subMsg2 := protocol.SubscribeMessage{Type: "subscribe", ProjectKey: "project2"}

	subData1, _ := json.Marshal(subMsg1)
	subData2, _ := json.Marshal(subMsg2)

	s.HandleMessage(sub1, string(subData1))
	s.HandleMessage(sub2, string(subData2))

	// Clear subscription confirmations
	sub1.messages = nil
	sub2.messages = nil

	// Publish to project1
	tree := domain.TaskTree{
		Root: domain.TaskNode{ID: "root", Title: "Project 1 Tree"},
	}

	pubMsg := protocol.PublishTreeMessage{
		Type:       "publishTree",
		ProjectKey: "project1",
		Tree:       tree,
	}
	pubData, _ := json.Marshal(pubMsg)

	publisher := &MockConnection{}
	s.HandleMessage(publisher, string(pubData))

	// Only sub1 should receive the message
	sub1Messages := sub1.GetMessages()
	sub2Messages := sub2.GetMessages()

	if len(sub1Messages) != 1 {
		t.Errorf("subscriber 1 expected 1 message, got %d", len(sub1Messages))
	}
	if len(sub2Messages) != 0 {
		t.Errorf("subscriber 2 expected 0 messages, got %d", len(sub2Messages))
	}
}

func TestServer_Unsubscribe(t *testing.T) {
	s := NewServer(store.NewStore())

	conn := &MockConnection{}

	// Subscribe
	subMsg := protocol.SubscribeMessage{Type: "subscribe", ProjectKey: "project1"}
	subData, _ := json.Marshal(subMsg)
	s.HandleMessage(conn, string(subData))

	// Verify subscription
	if len(s.GetSubscribers("project1")) != 1 {
		t.Error("expected 1 subscriber before unsubscribe")
	}

	// Unsubscribe
	s.Unsubscribe(conn)

	// Verify subscription was removed
	if len(s.GetSubscribers("project1")) != 0 {
		t.Error("expected 0 subscribers after unsubscribe")
	}
}
