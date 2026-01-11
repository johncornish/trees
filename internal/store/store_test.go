package store

import (
	"testing"
	"trees/internal/domain"
)

func TestStore_SetAndGet(t *testing.T) {
	s := NewStore()

	tree := domain.TaskTree{
		Root: domain.TaskNode{
			ID:    "root",
			Title: "Test Project",
		},
	}

	s.Set("project1", tree)

	retrieved, exists := s.Get("project1")
	if !exists {
		t.Error("expected tree to exist for project1")
	}
	if retrieved.Root.ID != "root" {
		t.Errorf("expected root ID %q, got %q", "root", retrieved.Root.ID)
	}
}

func TestStore_GetNonExistent(t *testing.T) {
	s := NewStore()

	_, exists := s.Get("nonexistent")
	if exists {
		t.Error("expected tree not to exist for nonexistent project")
	}
}

func TestStore_OverwriteExisting(t *testing.T) {
	s := NewStore()

	tree1 := domain.TaskTree{
		Root: domain.TaskNode{ID: "root1", Title: "Version 1"},
	}
	s.Set("project1", tree1)

	tree2 := domain.TaskTree{
		Root: domain.TaskNode{ID: "root2", Title: "Version 2"},
	}
	s.Set("project1", tree2)

	retrieved, _ := s.Get("project1")
	if retrieved.Root.ID != "root2" {
		t.Errorf("expected root ID %q, got %q", "root2", retrieved.Root.ID)
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	s := NewStore()

	done := make(chan bool)

	// Multiple goroutines writing
	for i := 0; i < 10; i++ {
		go func(id int) {
			tree := domain.TaskTree{
				Root: domain.TaskNode{
					ID:    "root",
					Title: "Concurrent Test",
				},
			}
			s.Set("project1", tree)
			done <- true
		}(i)
	}

	// Multiple goroutines reading
	for i := 0; i < 10; i++ {
		go func(id int) {
			s.Get("project1")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic and should have a tree stored
	_, exists := s.Get("project1")
	if !exists {
		t.Error("expected tree to exist after concurrent access")
	}
}
