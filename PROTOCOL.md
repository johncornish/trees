# Trees Server Protocol

This document describes the TCP protocol for the Trees server MVP.

## Overview

The Trees server uses a line-delimited JSON protocol over TCP. Each message is a single line of JSON terminated by a newline character (`\n`).

## Connection

- **Host:** localhost (or server IP)
- **Port:** 9090 (default)
- **Protocol:** TCP
- **Encoding:** UTF-8
- **Framing:** Newline-delimited (`\n`)

## Message Format

All messages are JSON objects with a `type` field that identifies the message type.

### Message Types

#### 1. SUBSCRIBE (Client → Server)

Subscribe to receive notifications for a specific project.

```json
{
  "type": "subscribe",
  "projectKey": "my-project"
}
```

**Fields:**
- `type` (string): Must be `"subscribe"`
- `projectKey` (string): The project key to subscribe to

**Response:**
The server responds with a SUBSCRIBED message.

---

#### 2. SUBSCRIBED (Server → Client)

Confirmation that subscription was successful.

```json
{
  "type": "subscribed",
  "projectKey": "my-project"
}
```

**Fields:**
- `type` (string): Always `"subscribed"`
- `projectKey` (string): The project key that was subscribed to

---

#### 3. PUBLISH TREE (Client → Server)

Publish a new task tree for a project.

```json
{
  "type": "publishTree",
  "projectKey": "my-project",
  "tree": {
    "root": {
      "id": "task-1",
      "title": "Root Task",
      "description": "Optional description",
      "status": "pending",
      "children": [
        {
          "id": "task-2",
          "title": "Subtask 1",
          "status": "in_progress",
          "children": []
        }
      ]
    }
  }
}
```

**Fields:**
- `type` (string): Must be `"publishTree"`
- `projectKey` (string): The project key to publish to
- `tree` (object): The task tree
  - `root` (TaskNode): The root node of the tree

**TaskNode structure:**
- `id` (string, required): Unique identifier for the node
- `title` (string, required): Human-readable title
- `description` (string, optional): Longer description
- `status` (string, required): Task status (e.g., "pending", "in_progress", "completed")
- `children` (array of TaskNode, optional): Child nodes

**Behavior:**
- The tree is stored in memory, replacing any existing tree for that project
- All subscribers to the project receive a TREE ADDED notification

---

#### 4. TREE ADDED (Server → Subscribers)

Notification that a new tree was published.

```json
{
  "type": "treeAdded",
  "projectKey": "my-project",
  "tree": {
    "root": {
      "id": "task-1",
      "title": "Root Task",
      "status": "pending",
      "children": []
    }
  }
}
```

**Fields:**
- `type` (string): Always `"treeAdded"`
- `projectKey` (string): The project key
- `tree` (object): The complete task tree that was published

**When sent:**
- Immediately after a PUBLISH TREE message is processed
- Only sent to clients subscribed to the specific project

---

## Protocol Flow

### Typical Session

```
Client                          Server
  |                               |
  |--- SUBSCRIBE (project1) ----->|
  |<--- SUBSCRIBED (project1) ----|
  |                               |
  |--- PUBLISH TREE (project1) -->|
  |<--- TREE ADDED (project1) ----|
  |                               |
  |     (disconnect)              |
```

### Multiple Subscribers

```
Client A                Server                Client B
  |                       |                       |
  |--- SUBSCRIBE -------->|<------- SUBSCRIBE ----|
  |<--- SUBSCRIBED -------|-------- SUBSCRIBED -->|
  |                       |                       |
  |--- PUBLISH TREE ----->|                       |
  |<--- TREE ADDED -------|-------- TREE ADDED -->|
  |                       |                       |
```

## Subscription Management

- **Multiple subscriptions:** A client can subscribe to multiple projects by sending multiple SUBSCRIBE messages
- **Project isolation:** Subscribers only receive notifications for projects they've subscribed to
- **Disconnection:** When a client disconnects, all their subscriptions are automatically removed

## Error Handling

- If a message cannot be parsed, the connection is closed
- If an unknown message type is received, the connection is closed
- Connection errors are logged server-side

## Example Raw TCP Session

Using `nc` (netcat):

```bash
# Terminal 1 - Subscriber
$ nc localhost 9090
{"type":"subscribe","projectKey":"test-project"}
{"type":"subscribed","projectKey":"test-project"}
{"type":"treeAdded","projectKey":"test-project","tree":{"root":{"id":"1","title":"New Task","status":"pending","children":[]}}}

# Terminal 2 - Publisher
$ nc localhost 9090
{"type":"publishTree","projectKey":"test-project","tree":{"root":{"id":"1","title":"New Task","status":"pending","children":[]}}}
```

## Implementation Notes

- All messages must be terminated with a newline (`\n`)
- The server uses newline framing: each line is one complete JSON message
- There is no message size limit currently (use reasonable sizes)
- No authentication or authorization in the MVP
- All data is stored in memory only (not persisted)
