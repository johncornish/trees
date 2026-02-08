package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"trees/graph"
)

// mockGitChecker always returns a fixed result for testing
type mockGitChecker struct {
	changed bool
}

func (m *mockGitChecker) HasFileChangedSince(commit, filePath string) (bool, error) {
	return m.changed, nil
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	h, err := NewHandler(path, &mockGitChecker{changed: false})
	if err != nil {
		t.Fatalf("unexpected error creating handler: %v", err)
	}
	return h
}

func newTestHandlerWithChecker(t *testing.T, checker graph.GitChecker) *Handler {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	h, err := NewHandler(path, checker)
	if err != nil {
		t.Fatalf("unexpected error creating handler: %v", err)
	}
	return h
}

func TestCreateClaim(t *testing.T) {
	h := newTestHandler(t)
	body := `{"content": "The auth module validates JWT tokens"}`
	req := httptest.NewRequest(http.MethodPost, "/claims", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["id"] == nil || resp["id"] == "" {
		t.Error("expected non-empty id in response")
	}
	if resp["content"] != "The auth module validates JWT tokens" {
		t.Errorf("expected content in response, got %v", resp["content"])
	}
}

func TestListClaims(t *testing.T) {
	h := newTestHandler(t)

	// Create two claims
	for _, content := range []string{"claim one", "claim two"} {
		body := `{"content": "` + content + `"}`
		req := httptest.NewRequest(http.MethodPost, "/claims", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Mux().ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/claims", nil)
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 claims, got %d", len(resp))
	}
}

func TestGetClaim(t *testing.T) {
	h := newTestHandler(t)

	// Create a claim
	body := `{"content": "test claim"}`
	req := httptest.NewRequest(http.MethodPost, "/claims", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	var created map[string]interface{}
	json.NewDecoder(w.Body).Decode(&created)
	id := created["id"].(string)

	// Get the claim
	req = httptest.NewRequest(http.MethodGet, "/claims/"+id, nil)
	w = httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["id"] != id {
		t.Errorf("expected id %q, got %v", id, resp["id"])
	}
}

func TestGetClaimNotFound(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/claims/nonexistent", nil)
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCreateEvidence(t *testing.T) {
	h := newTestHandler(t)
	body := `{"file_path": "/home/user/project/main.go", "line_ref": "1-3,7,13-70", "git_commit": "abc123def456"}`
	req := httptest.NewRequest(http.MethodPost, "/evidence", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["id"] == nil || resp["id"] == "" {
		t.Error("expected non-empty id")
	}
	if resp["file_path"] != "/home/user/project/main.go" {
		t.Errorf("expected file_path, got %v", resp["file_path"])
	}
	if resp["line_ref"] != "1-3,7,13-70" {
		t.Errorf("expected line_ref, got %v", resp["line_ref"])
	}
	if resp["git_commit"] != "abc123def456" {
		t.Errorf("expected git_commit, got %v", resp["git_commit"])
	}
}

func TestCreateEvidenceRequiresGitCommit(t *testing.T) {
	h := newTestHandler(t)
	body := `{"file_path": "/home/user/project/main.go", "line_ref": "1-3"}`
	req := httptest.NewRequest(http.MethodPost, "/evidence", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateEvidenceRejectsRelativePath(t *testing.T) {
	h := newTestHandler(t)
	body := `{"file_path": "relative/path.go", "line_ref": "1-3", "git_commit": "abc123"}`
	req := httptest.NewRequest(http.MethodPost, "/evidence", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListEvidence(t *testing.T) {
	h := newTestHandler(t)

	body := `{"file_path": "/home/user/file.go", "line_ref": "1-10", "git_commit": "abc123"}`
	req := httptest.NewRequest(http.MethodPost, "/evidence", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	req = httptest.NewRequest(http.MethodGet, "/evidence", nil)
	w = httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 evidence, got %d", len(resp))
	}
}

func TestGetEvidence(t *testing.T) {
	h := newTestHandler(t)

	body := `{"file_path": "/home/user/file.go", "line_ref": "1-10", "git_commit": "abc123"}`
	req := httptest.NewRequest(http.MethodPost, "/evidence", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	var created map[string]interface{}
	json.NewDecoder(w.Body).Decode(&created)
	id := created["id"].(string)

	req = httptest.NewRequest(http.MethodGet, "/evidence/"+id, nil)
	w = httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["valid"] != true {
		t.Errorf("expected valid=true, got %v", resp["valid"])
	}
}

func TestGetEvidenceInvalidWhenFileChanged(t *testing.T) {
	h := newTestHandlerWithChecker(t, &mockGitChecker{changed: true})

	body := `{"file_path": "/home/user/file.go", "line_ref": "1-10", "git_commit": "abc123"}`
	req := httptest.NewRequest(http.MethodPost, "/evidence", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	var created map[string]interface{}
	json.NewDecoder(w.Body).Decode(&created)
	id := created["id"].(string)

	req = httptest.NewRequest(http.MethodGet, "/evidence/"+id, nil)
	w = httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["valid"] != false {
		t.Errorf("expected valid=false, got %v", resp["valid"])
	}
}

func TestLinkEvidenceToClaim(t *testing.T) {
	h := newTestHandler(t)

	// Create claim
	req := httptest.NewRequest(http.MethodPost, "/claims", strings.NewReader(`{"content": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)
	var claim map[string]interface{}
	json.NewDecoder(w.Body).Decode(&claim)
	claimID := claim["id"].(string)

	// Create evidence
	req = httptest.NewRequest(http.MethodPost, "/evidence", strings.NewReader(`{"file_path": "/home/user/f.go", "line_ref": "1-5", "git_commit": "abc123"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)
	var ev map[string]interface{}
	json.NewDecoder(w.Body).Decode(&ev)
	evID := ev["id"].(string)

	// Link evidence to claim
	linkBody := `{"evidence_id": "` + evID + `"}`
	req = httptest.NewRequest(http.MethodPost, "/claims/"+claimID+"/evidence", strings.NewReader(linkBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Get claim and verify evidence is linked
	req = httptest.NewRequest(http.MethodGet, "/claims/"+claimID, nil)
	w = httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	var claimResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&claimResp)

	evidence, ok := claimResp["evidence"].([]interface{})
	if !ok {
		t.Fatalf("expected evidence array, got %T", claimResp["evidence"])
	}
	if len(evidence) != 1 {
		t.Errorf("expected 1 linked evidence, got %d", len(evidence))
	}
}

func TestHealthEndpoint(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected body %q, got %q", "OK", w.Body.String())
	}
}
