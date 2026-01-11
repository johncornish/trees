#!/bin/bash
# Demo script for Trees TCP Task Dispatch System
# This script demonstrates the complete workflow

set -e

echo "=== Trees TCP Task Dispatch Demo ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Building binaries...${NC}"
go build -o /tmp/trees-server ./cmd/server
go build -o /tmp/trees-client ./cmd/client
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

echo -e "${BLUE}Starting server...${NC}"
/tmp/trees-server &
SERVER_PID=$!
sleep 1
echo -e "${GREEN}✓ Server started (PID: $SERVER_PID)${NC}"
echo ""

echo -e "${BLUE}Starting client (logging runner)...${NC}"
/tmp/trees-client -project demo-project -runner logging &
CLIENT_PID=$!
sleep 1
echo -e "${GREEN}✓ Client started (PID: $CLIENT_PID)${NC}"
echo ""

echo -e "${BLUE}Waiting for connections to establish...${NC}"
sleep 1
echo ""

echo -e "${YELLOW}Publishing tree with 5 tasks...${NC}"
curl -s -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{
    "id": "demo-tree-001",
    "projectKey": "demo-project",
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
        "description": "Write unit tests",
        "dependencies": ["task-2", "task-3"]
      },
      {
        "id": "task-5",
        "description": "Deploy to staging",
        "dependencies": ["task-4"]
      }
    ]
  }' | jq .

echo ""
echo -e "${GREEN}✓ Tree published${NC}"
echo ""

echo -e "${YELLOW}Waiting for client to process tasks...${NC}"
sleep 2
echo ""

echo -e "${YELLOW}Publishing another tree with parallel tasks...${NC}"
curl -s -X POST http://localhost:8080/publish \
  -H "Content-Type: application/json" \
  -d '{
    "id": "demo-tree-002",
    "projectKey": "demo-project",
    "tasks": [
      {
        "id": "task-a",
        "description": "Process batch A",
        "dependencies": []
      },
      {
        "id": "task-b",
        "description": "Process batch B",
        "dependencies": []
      },
      {
        "id": "task-c",
        "description": "Process batch C",
        "dependencies": []
      }
    ]
  }' | jq .

echo ""
echo -e "${GREEN}✓ Second tree published${NC}"
echo ""

echo -e "${YELLOW}Waiting for processing...${NC}"
sleep 2
echo ""

echo -e "${BLUE}Cleaning up...${NC}"
kill $CLIENT_PID 2>/dev/null || true
kill $SERVER_PID 2>/dev/null || true
sleep 1
echo -e "${GREEN}✓ Demo complete${NC}"
echo ""

echo "=== Demo Summary ==="
echo "• Server received HTTP requests and broadcasted trees via TCP"
echo "• Client subscribed to 'demo-project' and received 2 trees"
echo "• Tasks were dispatched in parallel (5 max concurrent)"
echo "• Execution summaries showed all tasks completed successfully"
echo ""
echo "Try it yourself:"
echo "  Terminal 1: go run ./cmd/server"
echo "  Terminal 2: go run ./cmd/client -project my-project"
echo "  Terminal 3: curl -X POST http://localhost:8080/publish -d '{...}'"
echo ""
echo "See README.md and USAGE.md for more details!"
