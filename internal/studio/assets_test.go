package studio

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAssetsEmbedding(t *testing.T) {
	requiredFiles := []string{
		"assets/index.html",
		"assets/studio.css",
		"assets/studio.js",
		"assets/three.min.js",
	}

	for _, file := range requiredFiles {
		data, err := assetsFS.ReadFile(file)
		if err != nil {
			t.Fatalf("falha ao ler asset embutido %s: %v", file, err)
		}
		if len(data) == 0 {
			t.Errorf("arquivo embutido %s está vazio", file)
		}
	}
}

func TestIndexHTMLContract(t *testing.T) {
	data, err := assetsFS.ReadFile("assets/index.html")
	if err != nil {
		t.Fatalf("falha ao ler index.html: %v", err)
	}

	content := string(data)
	requiredElements := []string{
		"id=\"projectName\"",
		"id=\"projectVersion\"",
		"id=\"statusChip\"",
		"id=\"btnLive\"",
		"id=\"btnDemo\"",
		"id=\"hoFrom\"",
		"id=\"hoTo\"",
		"id=\"hoPayload\"",
		"id=\"hoGate\"",
		"id=\"stationsGrid\"",
		"id=\"govFindingCard\"",
		"id=\"lastUpdate\"",
		"studio.css",
		"studio.js",
		"three.min.js",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(content, elem) {
			t.Errorf("index.html está faltando elemento obrigatório: %s", elem)
		}
	}
}

func TestStaticAssetsServing(t *testing.T) {
	srv, err := NewServer(Config{Dir: t.TempDir()})
	if err != nil {
		t.Fatalf("falha ao criar servidor: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado status 200 ao requisitar /, obteve %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "forge-sdd studio") {
		t.Errorf("resposta de / não contém título esperado: %s", body)
	}

	// Testa CSS
	reqCSS := httptest.NewRequest(http.MethodGet, "/studio.css", nil)
	recCSS := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recCSS, reqCSS)
	if recCSS.Code != http.StatusOK {
		t.Fatalf("esperado status 200 ao requisitar /studio.css, obteve %d", recCSS.Code)
	}

	// Testa JS
	reqJS := httptest.NewRequest(http.MethodGet, "/studio.js", nil)
	recJS := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(recJS, reqJS)
	if recJS.Code != http.StatusOK {
		t.Fatalf("esperado status 200 ao requisitar /studio.js, obteve %d", recJS.Code)
	}

	jsBody, _ := io.ReadAll(recJS.Body)
	if !strings.Contains(string(jsBody), "renderMascotFallbackSvg") {
		t.Errorf("studio.js servido não contém função essencial renderMascotFallbackSvg")
	}
}
