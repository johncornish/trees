# Trees - TCP Task Dispatch System

A TCP-based client-server system for distributing and executing task trees in parallel.

## Quick Start

**Terminal 1 - Start Server:**
```bash
go run ./cmd/server
```

**Terminal 2 - Start Client:**
```bash
go run ./cmd/client -project my-project
```

**Terminal 3 - Publish a Tree:**
```bash
curl -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{
    "id": "tree-1",
    "projectKey": "my-project",
    "tasks": [
      {"id": "task-1", "description": "First task", "dependencies": []},
      {"id": "task-2", "description": "Second task", "dependencies": []},
      {"id": "task-3", "description": "Third task", "dependencies": []}
    ]
  }'
```

**Client Output:**
```
[CLIENT] Received tree tree-1 (project: my-project) with 3 tasks
[DISPATCHER] Starting dispatch for tree tree-1 (project: my-project) with 3 tasks
[RUNNER] Starting task task-1: First task
[RUNNER] Starting task task-2: Second task
[RUNNER] Starting task task-3: Third task
[DISPATCHER] Completed dispatch in 10.5ms: 3 successes, 0 failures

=== EXECUTION SUMMARY ===
  Total Tasks: 3
  Successes: 3
  Failures: 0
  Duration: 10.5ms
=========================
```

## Features

✅ **TCP Client/Server** - Persistent connections, JSON protocol
✅ **Parallel Task Dispatch** - Configurable concurrency (default: 5)
✅ **Pluggable Agent Runner** - Easy to swap stub for real agent launching
✅ **Project Isolation** - Subscribe by projectKey, only receive matching trees
✅ **Comprehensive Tests** - 15+ test cases covering all components
✅ **Clean Architecture** - Interface-based, easy to extend

## Architecture

```
Server (TCP :9000, HTTP :8080)
  ├─ Accept client connections
  ├─ Manage subscriptions by projectKey
  └─ Broadcast trees to subscribers

Client
  ├─ Connect & subscribe to projectKey
  ├─ Receive treeAdded messages
  └─ Dispatch tasks in parallel
      └─ AgentRunner interface (pluggable!)
          ├─ StubRunner (MVP: simulates work)
          ├─ LoggingRunner (MVP: instant logs)
          └─ RealAgentRunner (YOUR implementation here)
```

## Components

- **`types.go`** - Core data structures (TaskNode, Tree, Message)
- **`runner.go`** - AgentRunner interface + stub implementations
- **`dispatcher.go`** - Parallel task execution with concurrency control
- **`server.go`** - TCP server for client connections
- **`client.go`** - TCP client for receiving and dispatching trees
- **`cmd/server/`** - Server CLI
- **`cmd/client/`** - Client CLI

## Building

```bash
# Build all
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Documentation

📖 **[USAGE.md](USAGE.md)** - Complete usage guide, examples, and API reference
📖 **[DEV.md](DEV.md)** - TDD workflow and development practices

## Where to Plug In Real Agent Launching

The system uses an `AgentRunner` interface in `runner.go`:

```go
type AgentRunner interface {
    Run(ctx context.Context, task TaskNode) (TaskResult, error)
}
```

**MVP implementations:**
- `StubRunner` - Simulates work with configurable sleep
- `LoggingRunner` - Instantly logs task details

**Your implementation:**
Create a new runner that calls your actual AI agent API (OpenAI, Anthropic, etc.):

```go
type RealAgentRunner struct {
    apiKey string
}

func (r *RealAgentRunner) Run(ctx context.Context, task TaskNode) (TaskResult, error) {
    // Call your agent API here
    // Process task.Description
    // Return results
}
```

Then update `cmd/client/main.go` to use it:
```go
runner = NewRealAgentRunner(os.Getenv("API_KEY"))
```

See [USAGE.md](USAGE.md) for detailed implementation guide.

## Project Status

**✅ MVP Complete**
- TCP client connectivity + protocol
- Task dispatch with concurrency control
- Pluggable agent runner interface
- Comprehensive test suite

**🚧 Future Work**
- Dependency resolution (currently all tasks run in parallel)
- Authentication/authorization
- Persistent storage
- Tree versioning/deduplication
- Real agent integration example

## Testing

All components have comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run specific suites
go test -v -run TestDispatcher
go test -v -run TestServer
go test -v -run TestClient

# With verbose output
go test -v ./...
```

## Example Use Case

1. **CI/CD Pipeline**: Distribute build/test tasks across multiple workers
2. **Agent Orchestration**: Coordinate AI agents working on related tasks
3. **Distributed Computing**: Farm out computations to multiple clients
4. **Task Queue**: Simple pub/sub task distribution system

## License

MIT
