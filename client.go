package trees

import (
	"context"
	"encoding/json"
	"log"
	"net"
)

// Client connects to a Trees TCP server and dispatches tasks
type Client struct {
	serverAddress string
	projectKey    string
	dispatcher    *Dispatcher

	// Optional callback when a tree is received and processed
	OnTreeReceived func(ExecutionSummary)
}

// NewClient creates a new TCP client
func NewClient(serverAddress string, projectKey string, dispatcher *Dispatcher) *Client {
	return &Client{
		serverAddress: serverAddress,
		projectKey:    projectKey,
		dispatcher:    dispatcher,
	}
}

// Connect establishes a connection to the server, subscribes, and listens for trees
func (c *Client) Connect(ctx context.Context) error {
	log.Printf("[CLIENT] Connecting to server at %s", c.serverAddress)

	// Connect to server
	conn, err := net.Dial("tcp", c.serverAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("[CLIENT] Connected to server")

	// Send subscribe message
	encoder := json.NewEncoder(conn)
	subscribeMsg := Message{
		Type:       "subscribe",
		ProjectKey: c.projectKey,
	}

	if err := encoder.Encode(subscribeMsg); err != nil {
		return err
	}

	log.Printf("[CLIENT] Subscribed to project %q", c.projectKey)

	// Listen for messages
	decoder := json.NewDecoder(conn)

	// Create channel to handle decoding in a goroutine
	type decodeResult struct {
		msg Message
		err error
	}
	decodeChan := make(chan decodeResult, 1)

	for {
		// Start decoding in a goroutine so we can select on context
		go func() {
			var msg Message
			err := decoder.Decode(&msg)
			decodeChan <- decodeResult{msg: msg, err: err}
		}()

		select {
		case <-ctx.Done():
			log.Printf("[CLIENT] Context cancelled, disconnecting")
			return ctx.Err()

		case result := <-decodeChan:
			if result.err != nil {
				log.Printf("[CLIENT] Error decoding message: %v", result.err)
				return result.err
			}

			switch result.msg.Type {
			case "treeAdded":
				c.handleTreeAdded(ctx, result.msg)
			default:
				log.Printf("[CLIENT] Unknown message type: %q", result.msg.Type)
			}
		}
	}
}

// handleTreeAdded processes a treeAdded message
func (c *Client) handleTreeAdded(ctx context.Context, msg Message) {
	if msg.Tree == nil {
		log.Printf("[CLIENT] Received treeAdded with nil tree")
		return
	}

	tree := *msg.Tree

	log.Printf("[CLIENT] Received tree %s (project: %s) with %d tasks",
		tree.ID, tree.ProjectKey, len(tree.Tasks))

	// Dispatch tasks
	summary := c.dispatcher.Dispatch(ctx, tree)

	// Print summary
	log.Printf("[CLIENT] %s", summary.String())

	// Call callback if set
	if c.OnTreeReceived != nil {
		c.OnTreeReceived(summary)
	}
}
