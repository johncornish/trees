package graph

import (
	"fmt"
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Evidence) != 0 {
		t.Errorf("expected empty evidence map, got %d", len(g.Evidence))
	}
	if len(g.Claims) != 0 {
		t.Errorf("expected empty claims map, got %d", len(g.Claims))
	}
	if len(g.Edges) != 0 {
		t.Errorf("expected empty edges, got %d", len(g.Edges))
	}
}

func TestAddEvidence(t *testing.T) {
	g := New()
	ev := g.AddEvidence("/home/user/project/main.go", "1-3,7,13-70", "abc123def456")

	if ev.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if ev.FilePath != "/home/user/project/main.go" {
		t.Errorf("expected file path %q, got %q", "/home/user/project/main.go", ev.FilePath)
	}
	if ev.LineRef != "1-3,7,13-70" {
		t.Errorf("expected line ref %q, got %q", "1-3,7,13-70", ev.LineRef)
	}
	if ev.GitCommit != "abc123def456" {
		t.Errorf("expected git commit %q, got %q", "abc123def456", ev.GitCommit)
	}
	if ev.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
	if _, ok := g.Evidence[ev.ID]; !ok {
		t.Error("expected evidence to be stored in graph")
	}
}

func TestAddEvidenceRequiresGitCommit(t *testing.T) {
	g := New()
	ev := g.AddEvidence("/home/user/project/main.go", "1-3", "")

	if ev != nil {
		t.Error("expected nil for empty git commit")
	}
}

func TestAddEvidenceRequiresAbsolutePath(t *testing.T) {
	g := New()
	ev := g.AddEvidence("relative/path.go", "1-3", "abc123")

	if ev != nil {
		t.Error("expected nil for relative path")
	}
}

func TestAddClaim(t *testing.T) {
	g := New()
	claim := g.AddClaim("The authentication module validates tokens correctly")

	if claim.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if claim.Content != "The authentication module validates tokens correctly" {
		t.Errorf("expected content %q, got %q", "The authentication module validates tokens correctly", claim.Content)
	}
	if claim.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
	if _, ok := g.Claims[claim.ID]; !ok {
		t.Error("expected claim to be stored in graph")
	}
}

func TestLinkEvidenceToClaim(t *testing.T) {
	g := New()
	claim := g.AddClaim("Auth works")
	ev := g.AddEvidence("/home/user/auth.go", "10-25", "abc123")

	err := g.LinkEvidence(claim.ID, ev.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.Edges[0].ClaimID != claim.ID {
		t.Errorf("expected claim ID %q, got %q", claim.ID, g.Edges[0].ClaimID)
	}
	if g.Edges[0].EvidenceID != ev.ID {
		t.Errorf("expected evidence ID %q, got %q", ev.ID, g.Edges[0].EvidenceID)
	}
}

func TestLinkEvidenceInvalidClaim(t *testing.T) {
	g := New()
	ev := g.AddEvidence("/home/user/auth.go", "10-25", "abc123")

	err := g.LinkEvidence("nonexistent", ev.ID)
	if err == nil {
		t.Error("expected error for nonexistent claim")
	}
}

func TestLinkEvidenceInvalidEvidence(t *testing.T) {
	g := New()
	claim := g.AddClaim("Auth works")

	err := g.LinkEvidence(claim.ID, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent evidence")
	}
}

func TestGetEvidenceForClaim(t *testing.T) {
	g := New()
	claim := g.AddClaim("Auth works")
	ev1 := g.AddEvidence("/home/user/auth.go", "10-25", "abc123")
	ev2 := g.AddEvidence("/home/user/auth_test.go", "1-50", "abc123")
	g.AddEvidence("/home/user/unrelated.go", "1-5", "abc123") // not linked

	g.LinkEvidence(claim.ID, ev1.ID)
	g.LinkEvidence(claim.ID, ev2.ID)

	evidence := g.GetEvidenceForClaim(claim.ID)
	if len(evidence) != 2 {
		t.Fatalf("expected 2 evidence nodes, got %d", len(evidence))
	}

	ids := map[string]bool{}
	for _, e := range evidence {
		ids[e.ID] = true
	}
	if !ids[ev1.ID] || !ids[ev2.ID] {
		t.Error("expected both linked evidence nodes")
	}
}

func TestGetEvidenceByID(t *testing.T) {
	g := New()
	ev := g.AddEvidence("/home/user/main.go", "1-10", "abc123")

	found := g.GetEvidence(ev.ID)
	if found == nil {
		t.Fatal("expected to find evidence")
	}
	if found.ID != ev.ID {
		t.Errorf("expected ID %q, got %q", ev.ID, found.ID)
	}
}

func TestGetEvidenceByIDNotFound(t *testing.T) {
	g := New()
	found := g.GetEvidence("nonexistent")
	if found != nil {
		t.Error("expected nil for nonexistent evidence")
	}
}

func TestGetClaimByID(t *testing.T) {
	g := New()
	claim := g.AddClaim("test claim")

	found := g.GetClaim(claim.ID)
	if found == nil {
		t.Fatal("expected to find claim")
	}
	if found.ID != claim.ID {
		t.Errorf("expected ID %q, got %q", claim.ID, found.ID)
	}
}

func TestGetClaimByIDNotFound(t *testing.T) {
	g := New()
	found := g.GetClaim("nonexistent")
	if found != nil {
		t.Error("expected nil for nonexistent claim")
	}
}

// mockGitChecker is a test double for GitChecker
type mockGitChecker struct {
	changed map[string]bool // key: "commit:filepath"
	err     error
}

func (m *mockGitChecker) HasFileChangedSince(commit, filePath string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	key := commit + ":" + filePath
	return m.changed[key], nil
}

func TestCheckEvidenceValidWhenUnchanged(t *testing.T) {
	g := New()
	ev := g.AddEvidence("/home/user/auth.go", "10-25", "abc123")

	checker := &mockGitChecker{changed: map[string]bool{}}

	valid, err := g.CheckEvidence(ev.ID, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected evidence to be valid when file has not changed")
	}
}

func TestCheckEvidenceInvalidWhenChanged(t *testing.T) {
	g := New()
	ev := g.AddEvidence("/home/user/auth.go", "10-25", "abc123")

	checker := &mockGitChecker{
		changed: map[string]bool{
			"abc123:/home/user/auth.go": true,
		},
	}

	valid, err := g.CheckEvidence(ev.ID, checker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected evidence to be invalid when file has changed")
	}
}

func TestCheckEvidenceNotFound(t *testing.T) {
	g := New()
	checker := &mockGitChecker{changed: map[string]bool{}}

	_, err := g.CheckEvidence("nonexistent", checker)
	if err == nil {
		t.Error("expected error for nonexistent evidence")
	}
}

func TestCheckEvidenceGitError(t *testing.T) {
	g := New()
	ev := g.AddEvidence("/home/user/auth.go", "10-25", "abc123")

	checker := &mockGitChecker{
		err: fmt.Errorf("git not found"),
	}

	_, err := g.CheckEvidence(ev.ID, checker)
	if err == nil {
		t.Error("expected error when git check fails")
	}
}
