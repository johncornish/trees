package trees

import (
	"encoding/json"
	"log"
	"net"
	"sync"
)

// Server manages TCP connections and message routing
type Server struct {
	address      string
	listener     net.Listener
	subscribers  map[string][]*json.Encoder // projectKey -> list of client encoders
	subscriberMu sync.RWMutex
	stopChan     chan struct{}
}

// NewServer creates a new TCP server
func NewServer(address string) *Server {
	return &Server{
		address:     address,
		subscribers: make(map[string][]*json.Encoder),
		stopChan:    make(chan struct{}),
	}
}

// Start begins accepting TCP connections
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.listener = listener
	log.Printf("[SERVER] Listening on %s", s.listener.Addr().String())

	for {
		select {
		case <-s.stopChan:
			log.Printf("[SERVER] Stopping server")
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			// Check if we're stopping
			select {
			case <-s.stopChan:
				return nil
			default:
				log.Printf("[SERVER] Error accepting connection: %v", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	close(s.stopChan)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Address returns the server's listening address
func (s *Server) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.address
}

// handleConnection processes messages from a connected client
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("[SERVER] New connection from %s", conn.RemoteAddr())

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("[SERVER] Error decoding message from %s: %v", conn.RemoteAddr(), err)
			return
		}

		switch msg.Type {
		case "subscribe":
			s.handleSubscribe(msg.ProjectKey, encoder)
			log.Printf("[SERVER] Client %s subscribed to project %q", conn.RemoteAddr(), msg.ProjectKey)

		default:
			log.Printf("[SERVER] Unknown message type from %s: %q", conn.RemoteAddr(), msg.Type)
		}
	}
}

// handleSubscribe registers a client for a project
func (s *Server) handleSubscribe(projectKey string, encoder *json.Encoder) {
	s.subscriberMu.Lock()
	defer s.subscriberMu.Unlock()

	s.subscribers[projectKey] = append(s.subscribers[projectKey], encoder)
}

// PublishTree sends a tree to all subscribers of the project
func (s *Server) PublishTree(tree Tree) {
	s.subscriberMu.RLock()
	defer s.subscriberMu.RUnlock()

	subscribers := s.subscribers[tree.ProjectKey]
	if len(subscribers) == 0 {
		log.Printf("[SERVER] No subscribers for project %q", tree.ProjectKey)
		return
	}

	msg := Message{
		Type:       "treeAdded",
		ProjectKey: tree.ProjectKey,
		Tree:       &tree,
	}

	log.Printf("[SERVER] Publishing tree %s to %d subscribers of project %q",
		tree.ID, len(subscribers), tree.ProjectKey)

	// Send to all subscribers
	// Note: In production, we'd handle send failures and remove dead connections
	for _, encoder := range subscribers {
		if err := encoder.Encode(msg); err != nil {
			log.Printf("[SERVER] Error sending to subscriber: %v", err)
		}
	}
}
