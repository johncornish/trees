# Trees TCP Client MVP - Usage Guide

## Overview

This is a TCP client MVP for the Trees task dispatch system. It implements:

- **TCP Server**: Accepts client connections, manages subscriptions, and broadcasts task trees
- **TCP Client**: Connects to the server, subscribes to a project, and dispatches tasks in parallel
- **Pluggable Agent Runner**: Interface for executing tasks (MVP includes stub and logging implementations)

## Architecture

```
┌─────────────────┐         TCP          ┌──────────────────┐
│   Trees Server  │◄────────────────────►│  Trees Client    │
│                 │                       │                  │
│  - TCP (9000)   │   Subscribe(project) │  - Dispatcher    │
│  - HTTP (8080)  │   ───────────────►   │  - AgentRunner   │
│                 │                       │  - Concurrency   │
│  Publish API    │   TreeAdded(tree)    │    Control       │
│  POST /publish  │   ◄───────────────   │                  │
└─────────────────┘                       └──────────────────┘
                                                    │
                                                    ▼
                                          ┌──────────────────┐
                                          │  Task Execution  │
                                          │  (Parallel)      │
                                          │                  │
                                          │  • Max 5 workers │
                                          │  • Context-aware │
                                          │  • Result summary│
                                          └──────────────────┘
```

## Components

### 1. Core Data Structures

**TaskNode** - Represents a single task
```go
type TaskNode struct {
    ID           string   `json:"id"`
    Description  string   `json:"description"`
    Dependencies []string `json:"dependencies"`
}
```

**Tree** - A collection of tasks for a project
```go
type Tree struct {
    ID         string     `json:"id"`
    ProjectKey string     `json:"projectKey"`
    Tasks      []TaskNode `json:"tasks"`
}
```

**Message** - Protocol message (subscribe, treeAdded)
```go
type Message struct {
    Type       string `json:"type"`
    ProjectKey string `json:"projectKey,omitempty"`
    Tree       *Tree  `json:"tree,omitempty"`
}
```

### 2. AgentRunner Interface

The AgentRunner interface is the pluggable component where "real agent launching" can be implemented:

```go
type AgentRunner interface {
    Run(ctx context.Context, task TaskNode) (TaskResult, error)
}
```

**MVP Implementations:**

1. **StubRunner** - Simulates work by sleeping
   - Configurable sleep duration
   - Respects context cancellation
   - Logs task execution

2. **LoggingRunner** - Immediately logs task details
   - No actual work performed
   - Instant execution
   - Good for testing

**Future Implementation:**
Replace with a real agent launcher that:
- Spawns actual agent processes
- Calls external APIs (e.g., OpenAI)
- Manages agent lifecycle
- Handles authentication

See `runner.go` for the interface and stub implementations.

### 3. Dispatcher

The Dispatcher orchestrates parallel task execution:

- **Concurrency Control**: Limits concurrent tasks (default: 5)
- **Context-Aware**: Respects cancellation and timeouts
- **Result Aggregation**: Collects and summarizes all task results

### 4. TCP Protocol

**Subscribe Message** (Client → Server)
```json
{
  "type": "subscribe",
  "projectKey": "my-project"
}
```

**TreeAdded Message** (Server → Client)
```json
{
  "type": "treeAdded",
  "projectKey": "my-project",
  "tree": {
    "id": "tree-123",
    "projectKey": "my-project",
    "tasks": [
      {
        "id": "task-1",
        "description": "Build feature",
        "dependencies": []
      },
      {
        "id": "task-2",
        "description": "Write tests",
        "dependencies": ["task-1"]
      }
    ]
  }
}
```

## Building

```bash
# Build all binaries
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
go build -o bin/healthcheck ./cmd/healthcheck

# Or build individually
go build -o trees-server ./cmd/server
go build -o trees-client ./cmd/client
```

## Running

### Terminal 1: Start the Server

```bash
go run ./cmd/server
```

**Options:**
- `-tcp :9000` - TCP address for client connections (default: :9000)
- `-http :8080` - HTTP address for API and health check (default: :8080)

**Example:**
```bash
go run ./cmd/server -tcp :9000 -http :8080
```

**Output:**
```
Starting Trees Server
  TCP address: :9000 (for client connections)
  HTTP address: :8080 (for API and health check)
HTTP server listening on :8080
[SERVER] Listening on [::]:9000
```

### Terminal 2: Start the Client

```bash
go run ./cmd/client -project my-project
```

**Options:**
- `-server localhost:9000` - TCP server address (default: localhost:9000)
- `-project <key>` - Project key to subscribe to (REQUIRED)
- `-concurrency 5` - Max concurrent tasks (default: 5)
- `-runner stub` - Runner type: stub or logging (default: stub)
- `-sleep 100ms` - Sleep duration for stub runner (default: 100ms)

**Example:**
```bash
go run ./cmd/client -project my-project -concurrency 10 -runner stub -sleep 50ms
```

**Output:**
```
Starting Trees Client
  Server: localhost:9000
  Project: my-project
  Max Concurrency: 10
  Runner Type: stub
  Sleep Duration: 50ms
[CLIENT] Connecting to server at localhost:9000
[CLIENT] Connected to server
[CLIENT] Subscribed to project "my-project"
```

### Terminal 3: Publish a Tree

Use curl to publish a task tree via the HTTP API:

```bash
curl -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{
    "id": "tree-001",
    "projectKey": "my-project",
    "tasks": [
      {
        "id": "task-1",
        "description": "Initialize project structure",
        "dependencies": []
      },
      {
        "id": "task-2",
        "description": "Setup database schema",
        "dependencies": ["task-1"]
      },
      {
        "id": "task-3",
        "description": "Create API endpoints",
        "dependencies": ["task-1"]
      },
      {
        "id": "task-4",
        "description": "Write integration tests",
        "dependencies": ["task-2", "task-3"]
      }
    ]
  }'
```

## Example: Complete Workflow

### 1. Start Server
```bash
$ go run ./cmd/server
Starting Trees Server
  TCP address: :9000 (for client connections)
  HTTP address: :8080 (for API and health check)
[SERVER] Listening on [::]:9000
HTTP server listening on :8080
```

### 2. Start Client (Logging Runner)
```bash
$ go run ./cmd/client -project demo-project -runner logging
Starting Trees Client
  Server: localhost:9000
  Project: demo-project
  Max Concurrency: 5
  Runner Type: logging
[CLIENT] Connecting to server at localhost:9000
[CLIENT] Connected to server
[CLIENT] Subscribed to project "demo-project"
```

### 3. Publish Tree
```bash
$ curl -X POST http://localhost:8080/publish -H "Content-Type: application/json" -d '{
  "id": "demo-tree-1",
  "projectKey": "demo-project",
  "tasks": [
    {"id": "task-1", "description": "First task", "dependencies": []},
    {"id": "task-2", "description": "Second task", "dependencies": []},
    {"id": "task-3", "description": "Third task", "dependencies": []}
  ]
}'
```

**Response:**
```json
{"status":"published","treeId":"demo-tree-1"}
```

### 4. Client Output
```
[CLIENT] Received tree demo-tree-1 (project: demo-project) with 3 tasks
[DISPATCHER] Starting dispatch for tree demo-tree-1 (project: demo-project) with 3 tasks
[RUNNER] Task task-1: First task (dependencies: [])
[RUNNER] Task task-2: Second task (dependencies: [])
[RUNNER] Task task-3: Third task (dependencies: [])
[DISPATCHER] Completed dispatch in 245.7µs: 3 successes, 0 failures

=== EXECUTION SUMMARY ===
  Total Tasks: 3
  Successes: 3
  Failures: 0
  Duration: 245.7µs
=========================
```

## Testing

Run all tests:
```bash
go test ./...
```

Run specific test suites:
```bash
go test -v -run TestDispatcher  # Test dispatcher
go test -v -run TestServer      # Test TCP server
go test -v -run TestClient      # Test TCP client
go test -v -run TestRunner      # Test agent runners
```

Run with coverage:
```bash
go test -cover ./...
```

## Key Features Demonstrated

### ✅ TCP Client Connectivity
- Client connects to server via TCP
- Sends subscribe message with projectKey
- Maintains persistent connection

### ✅ Message Protocol
- JSON-based protocol
- Subscribe and TreeAdded message types
- Clean separation of concerns

### ✅ Task Dispatch
- Parallel execution of tasks
- Configurable concurrency limit (default: 5)
- Uses semaphore pattern to limit goroutines

### ✅ Pluggable Agent Runner
- Interface-based design
- Easy to swap implementations
- MVP includes stub and logging runners

### ✅ Project Isolation
- Clients only receive trees for their subscribed project
- Server maintains separate subscriber lists per project

### ✅ Execution Summaries
- Total tasks, successes, failures
- Duration tracking
- Clear console output

### ✅ Context-Aware Execution
- Respects cancellation signals
- Graceful shutdown
- No leaked goroutines

## Where to Plug In "Real Agent Launching"

The system is designed to make it easy to add real agent execution. Here's how:

### 1. Implement AgentRunner Interface

Create a new file `real_agent_runner.go`:

```go
package trees

import (
    "context"
    "time"
)

// RealAgentRunner launches actual AI agents
type RealAgentRunner struct {
    apiKey    string
    baseURL   string
    timeout   time.Duration
}

func NewRealAgentRunner(apiKey string) *RealAgentRunner {
    return &RealAgentRunner{
        apiKey:  apiKey,
        baseURL: "https://api.openai.com/v1",
        timeout: 5 * time.Minute,
    }
}

func (r *RealAgentRunner) Run(ctx context.Context, task TaskNode) (TaskResult, error) {
    start := time.Now()

    // 1. Create agent prompt from task description
    prompt := fmt.Sprintf("Execute this task: %s", task.Description)

    // 2. Call AI API (OpenAI, Anthropic, etc.)
    response, err := r.callAgentAPI(ctx, prompt)
    if err != nil {
        return TaskResult{
            TaskID:   task.ID,
            Success:  false,
            Error:    err,
            Duration: time.Since(start),
        }, err
    }

    // 3. Process agent response
    // ... implementation details ...

    return TaskResult{
        TaskID:   task.ID,
        Success:  true,
        Duration: time.Since(start),
    }, nil
}

func (r *RealAgentRunner) callAgentAPI(ctx context.Context, prompt string) (string, error) {
    // Implement actual API call here
    // Use context for timeouts and cancellation
    return "", nil
}
```

### 2. Update Client CLI

Modify `cmd/client/main.go` to support the new runner:

```go
case "real":
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY environment variable required for real runner")
    }
    runner = trees.NewRealAgentRunner(apiKey)
```

### 3. Use It

```bash
export OPENAI_API_KEY="sk-..."
go run ./cmd/client -project my-project -runner real
```

## Project Structure

```
trees/
├── cmd/
│   ├── client/          # TCP client CLI
│   │   └── main.go
│   ├── server/          # TCP server CLI
│   │   └── main.go
│   └── healthcheck/     # Original health check server
│       ├── main.go
│       └── main_test.go
├── types.go             # Core data structures
├── types_test.go
├── runner.go            # AgentRunner interface + implementations
├── runner_test.go
├── dispatcher.go        # Task dispatch with concurrency
├── dispatcher_test.go
├── server.go            # TCP server
├── server_test.go
├── client.go            # TCP client
├── client_test.go
├── go.mod
├── DEV.md              # TDD workflow guide
└── USAGE.md            # This file
```

## Notes on MVP Design

### ✅ What's Included
- Full TCP client/server implementation
- JSON protocol for subscribe/treeAdded
- Parallel task dispatch with concurrency limits
- Pluggable agent runner interface
- Comprehensive tests (15+ test cases)
- CLI programs ready to use

### ⚠️ What's NOT Included (Future Work)
- Authentication/authorization
- Persistent storage of trees or results
- Dependency resolution (tasks run in parallel regardless of dependencies)
- Tree versioning/deduplication
- Connection pooling or load balancing
- Real AI agent integration
- Result storage or retrieval API

### 🔌 Extension Points

1. **AgentRunner**: Swap stub for real agent launcher
2. **Dispatcher**: Add dependency resolution logic
3. **Server**: Add persistence layer (DB, Redis, etc.)
4. **Protocol**: Add authentication messages
5. **Client**: Add result reporting back to server

## Concurrency and Performance

### Current Implementation
- **Max Concurrency**: Configurable (default: 5)
- **Semaphore Pattern**: Limits goroutine spawning
- **No Dependencies**: All tasks run in parallel
- **Context-Aware**: Respects cancellation

### Performance Characteristics
- With 5 concurrent workers and 50ms stub delay:
  - 10 tasks complete in ~100ms (2 batches)
  - 100 tasks complete in ~1s (20 batches)

### Future Optimizations
- Add dependency graph resolution
- Implement topological sort for task ordering
- Add priority queues for task scheduling
- Support streaming results back to server

## Troubleshooting

### Client can't connect
- Ensure server is running: `curl http://localhost:8080/health`
- Check firewall settings
- Verify TCP port 9000 is not in use: `lsof -i :9000`

### No trees received
- Verify projectKey matches: client subscription vs. published tree
- Check server logs for publishing confirmation
- Ensure tree JSON is valid

### Tasks not executing in parallel
- Check `-concurrency` flag (must be > 1)
- Verify runner is being called (check logs)
- Look for context cancellation errors

### Tests failing
- Run `go test -v ./...` for detailed output
- Check for port conflicts (tests use random ports)
- Ensure no firewall blocking localhost connections

## License

MIT
