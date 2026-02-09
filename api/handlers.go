package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"trees/graph"
	"trees/store"
)

type Handler struct {
	store   *store.Store
	checker graph.GitChecker
	mux     *http.ServeMux
}

func NewHandler(storePath string, checker graph.GitChecker) (*Handler, error) {
	s, err := store.New(storePath)
	if err != nil {
		return nil, err
	}
	h := &Handler{store: s, checker: checker}
	h.setupRoutes()
	return h, nil
}

func (h *Handler) Mux() *http.ServeMux {
	return h.mux
}

func (h *Handler) setupRoutes() {
	h.mux = http.NewServeMux()
	h.mux.HandleFunc("GET /health", h.health)
	h.mux.HandleFunc("POST /claims", h.createClaim)
	h.mux.HandleFunc("GET /claims", h.listClaims)
	h.mux.HandleFunc("GET /claims/{id}", h.getClaim)
	h.mux.HandleFunc("POST /claims/{id}/evidence", h.linkEvidence)
	h.mux.HandleFunc("DELETE /claims/{id}", h.deleteClaim)
	h.mux.HandleFunc("PUT /claims/{id}", h.updateClaim)
	h.mux.HandleFunc("POST /evidence", h.createEvidence)
	h.mux.HandleFunc("GET /evidence", h.listEvidence)
	h.mux.HandleFunc("GET /evidence/{id}", h.getEvidence)
	h.mux.HandleFunc("DELETE /evidence/{id}", h.deleteEvidence)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) createClaim(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, `{"error": "content is required"}`, http.StatusBadRequest)
		return
	}

	var claim *graph.ClaimNode
	h.store.WithGraph(func(g *graph.Graph) {
		claim = g.AddClaim(req.Content)
	})
	h.store.Save()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(claim)
}

func (h *Handler) listClaims(w http.ResponseWriter, r *http.Request) {
	g := h.store.Graph()

	query := r.URL.Query().Get("q")
	var claims []*graph.ClaimNode
	if query != "" {
		claims = g.SearchClaims(query)
	} else {
		claims = make([]*graph.ClaimNode, 0, len(g.Claims))
		for _, c := range g.Claims {
			claims = append(claims, c)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claims)
}

type evidenceWithValidity struct {
	*graph.EvidenceNode
	Valid bool `json:"valid"`
}

func (h *Handler) getClaim(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g := h.store.Graph()

	claim := g.GetClaim(id)
	if claim == nil {
		http.Error(w, `{"error": "claim not found"}`, http.StatusNotFound)
		return
	}

	rawEvidence := g.GetEvidenceForClaim(id)
	evidence := make([]evidenceWithValidity, 0, len(rawEvidence))
	for _, ev := range rawEvidence {
		valid, _ := g.CheckEvidence(ev.ID, h.checker)
		evidence = append(evidence, evidenceWithValidity{EvidenceNode: ev, Valid: valid})
	}

	resp := struct {
		*graph.ClaimNode
		Evidence []evidenceWithValidity `json:"evidence"`
	}{
		ClaimNode: claim,
		Evidence:  evidence,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) linkEvidence(w http.ResponseWriter, r *http.Request) {
	claimID := r.PathValue("id")

	var req struct {
		EvidenceID string `json:"evidence_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	var linkErr error
	h.store.WithGraph(func(g *graph.Graph) {
		linkErr = g.LinkEvidence(claimID, req.EvidenceID)
	})
	if linkErr != nil {
		http.Error(w, `{"error": "`+linkErr.Error()+`"}`, http.StatusNotFound)
		return
	}
	h.store.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "linked"})
}

func (h *Handler) createEvidence(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath  string `json:"file_path"`
		LineRef   string `json:"line_ref"`
		GitCommit string `json:"git_commit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	var ev *graph.EvidenceNode
	h.store.WithGraph(func(g *graph.Graph) {
		ev = g.AddEvidence(req.FilePath, req.LineRef, req.GitCommit)
	})
	if ev == nil {
		http.Error(w, `{"error": "file_path must be absolute and git_commit is required"}`, http.StatusBadRequest)
		return
	}
	h.store.Save()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ev)
}

func (h *Handler) listEvidence(w http.ResponseWriter, r *http.Request) {
	g := h.store.Graph()
	evidence := make([]*graph.EvidenceNode, 0, len(g.Evidence))
	for _, e := range g.Evidence {
		evidence = append(evidence, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evidence)
}

func (h *Handler) getEvidence(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	g := h.store.Graph()

	ev := g.GetEvidence(id)
	if ev == nil {
		http.Error(w, `{"error": "evidence not found"}`, http.StatusNotFound)
		return
	}

	valid, _ := g.CheckEvidence(id, h.checker)

	resp := struct {
		*graph.EvidenceNode
		Valid bool `json:"valid"`
	}{
		EvidenceNode: ev,
		Valid:        valid,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) deleteClaim(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var deleted bool
	h.store.WithGraph(func(g *graph.Graph) {
		deleted = g.DeleteClaim(id)
	})
	if !deleted {
		http.Error(w, `{"error": "claim not found"}`, http.StatusNotFound)
		return
	}
	h.store.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *Handler) deleteEvidence(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var deleted bool
	h.store.WithGraph(func(g *graph.Graph) {
		deleted = g.DeleteEvidence(id)
	})
	if !deleted {
		http.Error(w, `{"error": "evidence not found"}`, http.StatusNotFound)
		return
	}
	h.store.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *Handler) updateClaim(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, `{"error": "content is required"}`, http.StatusBadRequest)
		return
	}

	var claim *graph.ClaimNode
	h.store.WithGraph(func(g *graph.Graph) {
		claim = g.UpdateClaim(id, req.Content)
	})
	if claim == nil {
		http.Error(w, `{"error": "claim not found"}`, http.StatusNotFound)
		return
	}
	h.store.Save()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claim)
}
