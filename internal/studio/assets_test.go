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
		"id=\"drawerBackdrop\"",
		"id=\"stationDrawer\"",
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

func TestStudioAssetsDrawerAndLinks(t *testing.T) {
	// 1. Validar Markup no index.html
	htmlData, err := assetsFS.ReadFile("assets/index.html")
	if err != nil {
		t.Fatalf("falha ao ler assets/index.html: %v", err)
	}
	htmlStr := string(htmlData)
	requiredHTML := []string{
		`id="drawerBackdrop"`,
		`class="drawer-backdrop"`,
		`closeStationDrawer()`,
		`id="stationDrawer"`,
		`class="station-drawer"`,
		`aria-label="Painel de Inspeção da Estação"`,
	}
	for _, req := range requiredHTML {
		if !strings.Contains(htmlStr, req) {
			t.Errorf("index.html não contém trecho obrigatório: %s", req)
		}
	}

	// 2. Validar CSS em studio.css (CA-01, CA-02, CA-10)
	cssData, err := assetsFS.ReadFile("assets/studio.css")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.css: %v", err)
	}
	cssStr := string(cssData)
	requiredCSS := []string{
		".station-card",
		"cursor: pointer",
		".station-drawer",
		".drawer-backdrop",
		".drawer-header",
		".drawer-tabs",
		".drawer-tab",
		".drawer-task-card",
		".btn-ide",
		".ide-link-btn",
		".copy-btn",
		".inspect-spec-btn",
		"@media (max-width: 600px)",
	}
	for _, req := range requiredCSS {
		if !strings.Contains(cssStr, req) {
			t.Errorf("studio.css não contém seletor/regra obrigatória: %s", req)
		}
	}

	// 3. Validar JS em studio.js (CA-01, CA-04, CA-10)
	jsData, err := assetsFS.ReadFile("assets/studio.js")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.js: %v", err)
	}
	jsStr := string(jsData)
	requiredJS := []string{
		"openStationDrawer",
		"closeStationDrawer",
		"switchDrawerTab",
		"selectTaskInDrawer",
		"copyCommand",
		"inspectSpec",
		"vscode://",
		"cursor://",
		"Escape",
		"activeDrawerStationId",
	}
	for _, req := range requiredJS {
		if !strings.Contains(jsStr, req) {
			t.Errorf("studio.js não contém função/recurso obrigatório: %s", req)
		}
	}

	// 4. Teste HTTP integrado dos assets servidos pelo NewServer
	srv, err := NewServer(Config{Dir: t.TempDir()})
	if err != nil {
		t.Fatalf("falha ao instanciar servidor NewServer: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/studio.js", nil)
	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado status 200 para /studio.js, obteve %d", rec.Code)
	}
	servedJS := rec.Body.String()
	if !strings.Contains(servedJS, "openStationDrawer") || !strings.Contains(servedJS, "vscode://") {
		t.Errorf("conteúdo servido de /studio.js não contém as rotinas do drawer e deep links")
	}
}
