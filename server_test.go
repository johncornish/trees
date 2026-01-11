package trees

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestServerAcceptsConnection(t *testing.T) {
	server := NewServer(":0") // Use random port
	go server.Start()
	defer server.Stop()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("tcp", server.Address())
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()
}

func TestServerHandlesSubscribe(t *testing.T) {
	server := NewServer(":0")
	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("tcp", server.Address())
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send subscribe message
	msg := Message{
		Type:       "subscribe",
		ProjectKey: "test-project",
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("failed to send subscribe message: %v", err)
	}

	// Server should accept the subscription (no response expected)
	// If we can read without error, subscription worked
	time.Sleep(50 * time.Millisecond)
}

func TestServerPublishesTreeToSubscribers(t *testing.T) {
	server := NewServer(":0")
	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Connect and subscribe
	conn, err := net.Dial("tcp", server.Address())
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send subscribe message
	subscribeMsg := Message{
		Type:       "subscribe",
		ProjectKey: "test-project",
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(subscribeMsg); err != nil {
		t.Fatalf("failed to send subscribe message: %v", err)
	}

	// Give time for subscription to register
	time.Sleep(50 * time.Millisecond)

	// Publish a tree
	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "test-project",
		Tasks: []TaskNode{
			{ID: "task-1", Description: "Test task"},
		},
	}

	server.PublishTree(tree)

	// Read the treeAdded message
	decoder := json.NewDecoder(conn)
	var receivedMsg Message

	// Set a timeout for reading
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	if err := decoder.Decode(&receivedMsg); err != nil {
		t.Fatalf("failed to receive treeAdded message: %v", err)
	}

	if receivedMsg.Type != "treeAdded" {
		t.Errorf("expected message type 'treeAdded', got %q", receivedMsg.Type)
	}

	if receivedMsg.ProjectKey != "test-project" {
		t.Errorf("expected projectKey 'test-project', got %q", receivedMsg.ProjectKey)
	}

	if receivedMsg.Tree == nil {
		t.Fatal("expected tree to be non-nil")
	}

	if receivedMsg.Tree.ID != "tree-1" {
		t.Errorf("expected tree ID 'tree-1', got %q", receivedMsg.Tree.ID)
	}
}

func TestServerOnlyNotifiesMatchingSubscribers(t *testing.T) {
	server := NewServer(":0")
	go server.Start()
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// Connect two clients with different subscriptions
	conn1, err := net.Dial("tcp", server.Address())
	if err != nil {
		t.Fatalf("failed to connect client 1: %v", err)
	}
	defer conn1.Close()

	conn2, err := net.Dial("tcp", server.Address())
	if err != nil {
		t.Fatalf("failed to connect client 2: %v", err)
	}
	defer conn2.Close()

	// Subscribe client 1 to "project-alpha"
	encoder1 := json.NewEncoder(conn1)
	subscribeMsg1 := Message{
		Type:       "subscribe",
		ProjectKey: "project-alpha",
	}
	if err := encoder1.Encode(subscribeMsg1); err != nil {
		t.Fatalf("failed to subscribe client 1: %v", err)
	}

	// Subscribe client 2 to "project-beta"
	encoder2 := json.NewEncoder(conn2)
	subscribeMsg2 := Message{
		Type:       "subscribe",
		ProjectKey: "project-beta",
	}
	if err := encoder2.Encode(subscribeMsg2); err != nil {
		t.Fatalf("failed to subscribe client 2: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Publish tree to "project-alpha"
	tree := Tree{
		ID:         "tree-1",
		ProjectKey: "project-alpha",
		Tasks:      []TaskNode{{ID: "task-1", Description: "Test"}},
	}
	server.PublishTree(tree)

	// Client 1 should receive the message
	decoder1 := json.NewDecoder(conn1)
	var msg1 Message
	conn1.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if err := decoder1.Decode(&msg1); err != nil {
		t.Fatalf("client 1 should have received message: %v", err)
	}

	if msg1.Type != "treeAdded" {
		t.Errorf("client 1 expected 'treeAdded', got %q", msg1.Type)
	}

	// Client 2 should NOT receive the message (timeout expected)
	decoder2 := json.NewDecoder(bufio.NewReader(conn2))
	var msg2 Message
	conn2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	err = decoder2.Decode(&msg2)
	if err == nil {
		t.Error("client 2 should not have received message for different project")
	}
}
