package store

import (
	"os"
	"path/filepath"
	"testing"
	"trees/graph"
)

func TestNewStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	s, err := New(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	if s.Graph() == nil {
		t.Fatal("expected non-nil graph")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	s, _ := New(path)
	g := s.Graph()

	claim := g.AddClaim("test claim")
	ev := g.AddEvidence("/home/user/file.go", "1-10", "abc123")
	g.LinkEvidence(claim.ID, ev.ID)

	if err := s.Save(); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected data file to exist")
	}

	s2, err := New(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	g2 := s2.Graph()

	if len(g2.Claims) != 1 {
		t.Fatalf("expected 1 claim, got %d", len(g2.Claims))
	}
	if len(g2.Evidence) != 1 {
		t.Fatalf("expected 1 evidence, got %d", len(g2.Evidence))
	}
	if len(g2.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g2.Edges))
	}

	loadedClaim := g2.GetClaim(claim.ID)
	if loadedClaim == nil {
		t.Fatal("expected to find claim after load")
	}
	if loadedClaim.Content != "test claim" {
		t.Errorf("expected content %q, got %q", "test claim", loadedClaim.Content)
	}

	loadedEv := g2.GetEvidence(ev.ID)
	if loadedEv == nil {
		t.Fatal("expected to find evidence after load")
	}
	if loadedEv.FilePath != "/home/user/file.go" {
		t.Errorf("expected file path %q, got %q", "/home/user/file.go", loadedEv.FilePath)
	}
}

func TestLoadCreatesDirectoryIfNeeded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "data.json")

	s, err := New(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s.Graph().AddClaim("test")
	if err := s.Save(); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected nested data file to exist")
	}
}

func TestConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	s, _ := New(path)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			s.WithGraph(func(g *graph.Graph) {
				g.AddClaim("concurrent claim")
			})
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	g := s.Graph()
	if len(g.Claims) != 10 {
		t.Errorf("expected 10 claims, got %d", len(g.Claims))
	}
}
