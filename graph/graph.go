package graph

import (
	"crypto/rand"
	"fmt"
	"path/filepath"
	"time"
)

type EvidenceNode struct {
	ID        string    `json:"id"`
	FilePath  string    `json:"file_path"`
	LineRef   string    `json:"line_ref"`
	GitCommit string    `json:"git_commit"`
	CreatedAt time.Time `json:"created_at"`
}

type ClaimNode struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Edge struct {
	ClaimID    string `json:"claim_id"`
	EvidenceID string `json:"evidence_id"`
}

type Graph struct {
	Evidence map[string]*EvidenceNode `json:"evidence"`
	Claims   map[string]*ClaimNode    `json:"claims"`
	Edges    []Edge                   `json:"edges"`
}

func New() *Graph {
	return &Graph{
		Evidence: make(map[string]*EvidenceNode),
		Claims:   make(map[string]*ClaimNode),
		Edges:    []Edge{},
	}
}

func (g *Graph) AddEvidence(filePath, lineRef, gitCommit string) *EvidenceNode {
	if !filepath.IsAbs(filePath) {
		return nil
	}
	if gitCommit == "" {
		return nil
	}
	ev := &EvidenceNode{
		ID:        newID(),
		FilePath:  filePath,
		LineRef:   lineRef,
		GitCommit: gitCommit,
		CreatedAt: time.Now(),
	}
	g.Evidence[ev.ID] = ev
	return ev
}

func (g *Graph) AddClaim(content string) *ClaimNode {
	claim := &ClaimNode{
		ID:        newID(),
		Content:   content,
		CreatedAt: time.Now(),
	}
	g.Claims[claim.ID] = claim
	return claim
}

func (g *Graph) LinkEvidence(claimID, evidenceID string) error {
	if _, ok := g.Claims[claimID]; !ok {
		return fmt.Errorf("claim %q not found", claimID)
	}
	if _, ok := g.Evidence[evidenceID]; !ok {
		return fmt.Errorf("evidence %q not found", evidenceID)
	}
	g.Edges = append(g.Edges, Edge{ClaimID: claimID, EvidenceID: evidenceID})
	return nil
}

func (g *Graph) GetEvidenceForClaim(claimID string) []*EvidenceNode {
	var result []*EvidenceNode
	for _, edge := range g.Edges {
		if edge.ClaimID == claimID {
			if ev, ok := g.Evidence[edge.EvidenceID]; ok {
				result = append(result, ev)
			}
		}
	}
	return result
}

func (g *Graph) GetEvidence(id string) *EvidenceNode {
	return g.Evidence[id]
}

// CheckEvidence returns true if the evidence is still valid (the referenced
// file has not changed since the recorded git commit). Returns an error if
// the evidence ID is not found or the git check fails.
func (g *Graph) CheckEvidence(id string, checker GitChecker) (bool, error) {
	ev, ok := g.Evidence[id]
	if !ok {
		return false, fmt.Errorf("evidence %q not found", id)
	}
	changed, err := checker.HasFileChangedSince(ev.GitCommit, ev.FilePath)
	if err != nil {
		return false, err
	}
	return !changed, nil
}

func (g *Graph) GetClaim(id string) *ClaimNode {
	return g.Claims[id]
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
