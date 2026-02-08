package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"trees/graph"
)

type Store struct {
	path string
	g    *graph.Graph
	mu   sync.RWMutex
}

func New(path string) (*Store, error) {
	s := &Store{
		path: path,
		g:    graph.New(),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Graph() *graph.Graph {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.g
}

func (s *Store) WithGraph(fn func(g *graph.Graph)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.g)
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s.g)
}
