package server

import (
	"encoding/json"
	"fmt"
	"sync"
	"trees/internal/protocol"
	"trees/internal/store"
)

// Connection represents a client connection (humble interface for testing).
type Connection interface {
	ReadLine() (string, error)
	WriteLine(line string) error
	Close() error
}

// Server manages subscriptions and broadcasts tree updates.
type Server struct {
	store         *store.Store
	mu            sync.RWMutex
	subscriptions map[string][]Connection // projectKey -> list of connections
}

// NewServer creates a new server instance.
func NewServer(store *store.Store) *Server {
	return &Server{
		store:         store,
		subscriptions: make(map[string][]Connection),
	}
}

// HandleMessage processes a message from a client connection.
func (s *Server) HandleMessage(conn Connection, line string) error {
	msgType, err := protocol.ParseMessageType([]byte(line))
	if err != nil {
		return fmt.Errorf("failed to parse message type: %w", err)
	}

	switch msgType {
	case protocol.TypeSubscribe:
		return s.handleSubscribe(conn, line)
	case protocol.TypePublishTree:
		return s.handlePublishTree(conn, line)
	default:
		return fmt.Errorf("unknown message type: %s", msgType)
	}
}

func (s *Server) handleSubscribe(conn Connection, line string) error {
	var msg protocol.SubscribeMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return fmt.Errorf("failed to unmarshal subscribe message: %w", err)
	}

	// Register subscription
	s.mu.Lock()
	s.subscriptions[msg.ProjectKey] = append(s.subscriptions[msg.ProjectKey], conn)
	s.mu.Unlock()

	// Send confirmation
	response := protocol.SubscribedMessage{
		Type:       protocol.TypeSubscribed,
		ProjectKey: msg.ProjectKey,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal subscribed message: %w", err)
	}

	return conn.WriteLine(string(data))
}

func (s *Server) handlePublishTree(conn Connection, line string) error {
	var msg protocol.PublishTreeMessage
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return fmt.Errorf("failed to unmarshal publishTree message: %w", err)
	}

	// Store the tree
	s.store.Set(msg.ProjectKey, msg.Tree)

	// Broadcast to all subscribers
	broadcast := protocol.TreeAddedMessage{
		Type:       protocol.TypeTreeAdded,
		ProjectKey: msg.ProjectKey,
		Tree:       msg.Tree,
	}

	data, err := json.Marshal(broadcast)
	if err != nil {
		return fmt.Errorf("failed to marshal treeAdded message: %w", err)
	}

	s.broadcastToProject(msg.ProjectKey, string(data))

	return nil
}

func (s *Server) broadcastToProject(projectKey string, message string) {
	s.mu.RLock()
	subscribers := s.subscriptions[projectKey]
	s.mu.RUnlock()

	for _, conn := range subscribers {
		// Send to each subscriber (ignore errors for now)
		conn.WriteLine(message)
	}
}

// Unsubscribe removes a connection from all subscriptions.
func (s *Server) Unsubscribe(conn Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for projectKey, subscribers := range s.subscriptions {
		filtered := make([]Connection, 0, len(subscribers))
		for _, sub := range subscribers {
			if sub != conn {
				filtered = append(filtered, sub)
			}
		}
		s.subscriptions[projectKey] = filtered
	}
}

// GetSubscribers returns the list of subscribers for a project (for testing).
func (s *Server) GetSubscribers(projectKey string) []Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subscriptions[projectKey]
}
