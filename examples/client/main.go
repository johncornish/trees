package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"trees/internal/domain"
	"trees/internal/protocol"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go subscribe <projectKey>")
		fmt.Println("  go run main.go publish <projectKey>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "subscribe":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run main.go subscribe <projectKey>")
		}
		runSubscriber(os.Args[2])
	case "publish":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run main.go publish <projectKey>")
		}
		runPublisher(os.Args[2])
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func runSubscriber(projectKey string) {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to server. Subscribing to project: %s\n", projectKey)

	// Subscribe
	subMsg := protocol.SubscribeMessage{
		Type:       "subscribe",
		ProjectKey: projectKey,
	}
	if err := writeJSON(conn, subMsg); err != nil {
		log.Fatalf("Failed to send subscribe: %v", err)
	}

	// Read messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Received: %s\n", line)

		// Parse message type
		msgType, err := protocol.ParseMessageType([]byte(line))
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		switch msgType {
		case protocol.TypeSubscribed:
			var msg protocol.SubscribedMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				log.Printf("Failed to unmarshal subscribed: %v", err)
				continue
			}
			fmt.Printf("✓ Successfully subscribed to: %s\n", msg.ProjectKey)

		case protocol.TypeTreeAdded:
			var msg protocol.TreeAddedMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				log.Printf("Failed to unmarshal treeAdded: %v", err)
				continue
			}
			fmt.Printf("✓ Tree added for project: %s\n", msg.ProjectKey)
			fmt.Printf("  Root: %s (%s)\n", msg.Tree.Root.Title, msg.Tree.Root.Status)
			printTree(msg.Tree.Root, "  ")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Connection error: %v", err)
	}
}

func runPublisher(projectKey string) {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to server. Publishing tree to project: %s\n", projectKey)

	// Create example tree
	tree := domain.TaskTree{
		Root: domain.TaskNode{
			ID:          "root",
			Title:       "Build Feature X",
			Description: "Complete implementation of Feature X",
			Status:      "in_progress",
			Children: []domain.TaskNode{
				{
					ID:     "task-1",
					Title:  "Design API",
					Status: "completed",
				},
				{
					ID:     "task-2",
					Title:  "Implement backend",
					Status: "in_progress",
					Children: []domain.TaskNode{
						{
							ID:     "task-2-1",
							Title:  "Write tests",
							Status: "completed",
						},
						{
							ID:     "task-2-2",
							Title:  "Write implementation",
							Status: "in_progress",
						},
					},
				},
				{
					ID:     "task-3",
					Title:  "Deploy to staging",
					Status: "pending",
				},
			},
		},
	}

	// Publish tree
	pubMsg := protocol.PublishTreeMessage{
		Type:       "publishTree",
		ProjectKey: projectKey,
		Tree:       tree,
	}

	if err := writeJSON(conn, pubMsg); err != nil {
		log.Fatalf("Failed to publish tree: %v", err)
	}

	fmt.Println("✓ Tree published successfully!")
	fmt.Printf("  Root: %s (%s)\n", tree.Root.Title, tree.Root.Status)
	printTree(tree.Root, "  ")
}

func writeJSON(conn net.Conn, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(conn, "%s\n", data)
	return err
}

func printTree(node domain.TaskNode, indent string) {
	for _, child := range node.Children {
		fmt.Printf("%s└─ %s (%s)\n", indent, child.Title, child.Status)
		if len(child.Children) > 0 {
			printTree(child, indent+"   ")
		}
	}
}
