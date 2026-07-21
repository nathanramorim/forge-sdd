package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
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
			"beta":   "1.7.1-beta.5",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Override URL
	oldURL := NpmRegistryURL
	NpmRegistryURL = server.URL
	defer func() { NpmRegistryURL = oldURL }()

	latest, beta, err := FetchNpmVersions()
	if err != nil {
		t.Fatalf("FetchNpmVersions falhou: %v", err)
	}

	if latest != "1.6.0" {
		t.Errorf("esperava latest '1.6.0', obteve '%s'", latest)
	}
	if beta != "1.7.1-beta.5" {
		t.Errorf("esperava beta '1.7.1-beta.5', obteve '%s'", beta)
	}
}

func TestFetchNpmVersions_NetworkFailureReturnsExplicitError(t *testing.T) {
	oldURL := NpmRegistryURL
	oldTimeout := npmFetchTimeout
	NpmRegistryURL = "http://127.0.0.1:1" // porta inválida: falha de conexão imediata
	npmFetchTimeout = 500 * time.Millisecond
	defer func() {
		NpmRegistryURL = oldURL
		npmFetchTimeout = oldTimeout
	}()

	latest, beta, err := FetchNpmVersions()
	if err == nil {
		t.Fatal("esperava erro explícito quando o NPM Registry está inacessível, obteve nil")
	}
	if latest != "" || beta != "" {
		t.Errorf("esperava latest/beta vazios em caso de erro, obteve latest=%q beta=%q", latest, beta)
	}
}

func TestResolveUpgradeTarget(t *testing.T) {
	cases := []struct {
		name, latest, beta, want string
	}{
		{"beta mais recente que latest", "1.9.0", "1.9.1-beta", "1.9.1-beta"},
		{"latest mais recente que beta", "1.9.2", "1.9.1-beta", "1.9.2"},
		{"apenas latest disponível", "1.9.0", "", "1.9.0"},
		{"apenas beta disponível", "", "1.9.1-beta", "1.9.1-beta"},
		{"nenhuma versão disponível", "", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ResolveUpgradeTarget(c.latest, c.beta)
			if got != c.want {
				t.Errorf("ResolveUpgradeTarget(%q, %q) = %q, esperava %q", c.latest, c.beta, got, c.want)
			}
		})
	}
}
