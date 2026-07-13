package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sdd-project-test-*")
	if err != nil {
		t.Fatalf("falha ao criar pasta temporaria: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Caso 1: sem projeto
	if DetectProject(tempDir) {
		t.Errorf("DetectProject devia retornar false para pasta limpa")
	}

	// Caso 2: com projeto
	sddDir := filepath.Join(tempDir, "sdd")
	if err := os.Mkdir(sddDir, 0755); err != nil {
		t.Fatalf("falha ao criar pasta sdd: %v", err)
	}
	sddrcPath := filepath.Join(sddDir, ".sddrc")
	if err := os.WriteFile(sddrcPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("falha ao criar .sddrc: %v", err)
	}

	if !DetectProject(tempDir) {
		t.Errorf("DetectProject devia retornar true para pasta com sdd/.sddrc")
	}
}

func TestFetchNpmVersions(t *testing.T) {
	// Mock server
	mockResponse := npmRegistryResponse{
		DistTags: map[string]string{
			"latest": "1.6.0",
			"beta":   "1.9.1-beta",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Override URL
	oldURL := npmRegistryURL
	npmRegistryURL = server.URL
	defer func() { npmRegistryURL = oldURL }()

	latest, beta, err := FetchNpmVersions()
	if err != nil {
		t.Fatalf("FetchNpmVersions falhou: %v", err)
	}

	if latest != "1.6.0" {
		t.Errorf("esperava latest '1.6.0', obteve '%s'", latest)
	}
	if beta != "1.9.1-beta" {
		t.Errorf("esperava beta '1.9.1-beta', obteve '%s'", beta)
	}
}
