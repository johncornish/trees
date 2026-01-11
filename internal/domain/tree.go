package domain

// TaskNode represents a single node in a task tree.
type TaskNode struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	Children    []TaskNode `json:"children,omitempty"`
}

// TaskTree represents a complete task tree for a project.
type TaskTree struct {
	Root TaskNode `json:"root"`
}
