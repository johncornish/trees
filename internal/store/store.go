package store

import (
	"sync"
	"trees/internal/domain"
)

// Store holds task trees in memory, keyed by project key.
type Store struct {
	mu    sync.RWMutex
	trees map[string]domain.TaskTree
}

// NewStore creates a new in-memory store.
func NewStore() *Store {
	return &Store{
		trees: make(map[string]domain.TaskTree),
	}
}

// Set stores a task tree for the given project key.
func (s *Store) Set(projectKey string, tree domain.TaskTree) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trees[projectKey] = tree
}

// Get retrieves a task tree for the given project key.
// Returns the tree and a boolean indicating whether the tree exists.
func (s *Store) Get(projectKey string) (domain.TaskTree, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tree, exists := s.trees[projectKey]
	return tree, exists
}
