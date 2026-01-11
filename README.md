# Trees Server MVP

An in-memory task tree server that supports pub/sub over TCP. Clients can subscribe to project keys and receive real-time notifications when task trees are published.

## Features

- **In-memory task trees**: Store hierarchical task structures per project
- **TCP pub/sub**: Subscribe to projects and receive real-time updates
- **No authentication**: Simple MVP with no auth layer
- **Concurrent-safe**: Thread-safe operations with proper locking
- **Newline-delimited JSON**: Simple, text-based protocol

## Quick Start

### Start the Server

```bash
go run main.go
```

The server listens on:
- **TCP port 9090** - Trees pub/sub protocol
- **HTTP port 8080** - Health check endpoint (`/health`)

### Example: Subscribe to a Project

Terminal 1 (Subscriber):
```bash
cd examples/client
go run main.go subscribe my-project
```

Terminal 2 (Publisher):
```bash
cd examples/client
go run main.go publish my-project
```

The subscriber will receive the published tree in real-time!

### Using Raw TCP (netcat)

Terminal 1 (Subscriber):
```bash
nc localhost 9090
{"type":"subscribe","projectKey":"test"}
# Server responds: {"type":"subscribed","projectKey":"test"}
# Wait for tree notifications...
```

Terminal 2 (Publisher):
```bash
nc localhost 9090
{"type":"publishTree","projectKey":"test","tree":{"root":{"id":"1","title":"My Task","status":"pending","children":[]}}}
```

## Architecture

The codebase follows TDD principles and the Humble Object pattern:

```
.
├── internal/
│   ├── domain/          # Task tree data structures
│   ├── store/           # In-memory storage
│   ├── protocol/        # Message types and parsing
│   └── server/          # Server logic and TCP handling
├── examples/
│   └── client/          # Example Go client
├── main.go              # Entry point
├── DEV.md               # Development guide (TDD workflow)
└── PROTOCOL.md          # Protocol specification
```

### Key Design Decisions

1. **Humble Object Pattern**: TCP I/O is separated from business logic, making the core testable without real network connections
2. **Thread-safe**: All operations use proper locking for concurrent access
3. **Newline framing**: Simple, debuggable protocol using JSON lines
4. **Project isolation**: Subscribers only receive updates for their subscribed projects

## Protocol

See [PROTOCOL.md](./PROTOCOL.md) for complete protocol documentation.

### Message Types

1. **subscribe** - Subscribe to a project
2. **subscribed** - Confirmation of subscription
3. **publishTree** - Publish a new task tree
4. **treeAdded** - Broadcast to subscribers when tree is published

### Task Tree Structure

```json
{
  "root": {
    "id": "task-1",
    "title": "Root Task",
    "description": "Optional description",
    "status": "pending",
    "children": [
      {
        "id": "task-2",
        "title": "Subtask",
        "status": "in_progress",
        "children": []
      }
    ]
  }
}
```

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Run tests for a specific package:
```bash
go test ./internal/server/
```

### Test Coverage

- **Domain**: Task tree data structures
- **Store**: Concurrent in-memory storage
- **Protocol**: Message serialization/deserialization
- **Server**: Subscription management and broadcast logic

The server logic is fully tested using mock connections (no real TCP required).

## Development

This project follows Test-Driven Development (TDD) principles. See [DEV.md](./DEV.md) for:

- The Red-Green-Refactor cycle
- The Transformation Priority Premise
- The Humble Object pattern
- Testing guidelines

### Adding Features

1. Write tests first (Red)
2. Make them pass with simplest code (Green)
3. Refactor while keeping tests green
4. Run full test suite: `go test ./...`

## API Examples

### Go Client

See [examples/client/main.go](./examples/client/main.go) for a complete example.

```go
// Connect
conn, _ := net.Dial("tcp", "localhost:9090")

// Subscribe
subMsg := protocol.SubscribeMessage{
    Type:       "subscribe",
    ProjectKey: "my-project",
}
json.NewEncoder(conn).Encode(subMsg)

// Publish
tree := domain.TaskTree{
    Root: domain.TaskNode{
        ID:     "root",
        Title:  "My Task",
        Status: "pending",
    },
}
pubMsg := protocol.PublishTreeMessage{
    Type:       "publishTree",
    ProjectKey: "my-project",
    Tree:       tree,
}
json.NewEncoder(conn).Encode(pubMsg)
```

## Limitations (MVP)

- **No persistence**: All data is stored in memory only
- **No authentication**: Anyone can connect and publish/subscribe
- **No TLS**: Communication is unencrypted
- **No message replay**: Subscribers only receive new trees published after they subscribe
- **No tree updates**: Only full tree replacement is supported

## Future Enhancements

- Persistent storage (database)
- Authentication and authorization
- TLS encryption
- Tree updates/patches (not just full replacement)
- Message history/replay for new subscribers
- WebSocket support
- REST API alongside TCP

## License

See [LICENSE](./LICENSE) for details.