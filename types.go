package trees

// TaskNode represents a single task in a task tree
type TaskNode struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
}

// Tree represents a collection of tasks for a project
type Tree struct {
	ID         string     `json:"id"`
	ProjectKey string     `json:"projectKey"`
	Tasks      []TaskNode `json:"tasks"`
}

// Message represents a protocol message exchanged between client and server
type Message struct {
	Type       string `json:"type"`
	ProjectKey string `json:"projectKey,omitempty"`
	Tree       *Tree  `json:"tree,omitempty"`
}
