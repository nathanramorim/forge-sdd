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
		"id=\"govPillarsGrid\"",
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
		".drawer-section-title",
		".queue-items-box",
		".btn-ide",
		".ide-link-btn",
		".copy-btn",
		".inspect-spec-btn",
		".btn-handoff",
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
		"sendHandoff",
		"handleHandoffClick",
		"/api/handoff",
		"vscode://",
		"cursor://",
		"Escape",
		"activeDrawerStationId",
		"getToolColor",
		"queue-ver-",
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

func TestStudioAssetsSpecViewer(t *testing.T) {
	// 1. Validar Markup no index.html (feat-63-03)
	htmlData, err := assetsFS.ReadFile("assets/index.html")
	if err != nil {
		t.Fatalf("falha ao ler assets/index.html: %v", err)
	}
	htmlStr := string(htmlData)
	requiredHTML := []string{
		`id="specModalBackdrop"`,
		`class="spec-modal-backdrop"`,
		`closeSpecViewer()`,
		`id="specViewerModal"`,
		`class="spec-viewer-modal"`,
		`aria-label="Visualizador de Especificação"`,
		`copySpecPath()`,
		`handleApproveSpecClick()`,
	}
	for _, req := range requiredHTML {
		if !strings.Contains(htmlStr, req) {
			t.Errorf("index.html não contém trecho obrigatório: %s", req)
		}
	}

	// 2. Validar CSS em studio.css (feat-63-03)
	cssData, err := assetsFS.ReadFile("assets/studio.css")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.css: %v", err)
	}
	cssStr := string(cssData)
	requiredCSS := []string{
		".spec-modal-backdrop",
		".spec-viewer-modal",
		"max-width: 92vw",
		"max-height: 88vh",
		".approve-spec-btn",
		"#3ecf8e",
		".markdown-body",
		".checklist-item",
	}
	for _, req := range requiredCSS {
		if !strings.Contains(cssStr, req) {
			t.Errorf("studio.css não contém seletor/regra obrigatória: %s", req)
		}
	}

	// 3. Validar JS em studio.js (feat-63-03)
	jsData, err := assetsFS.ReadFile("assets/studio.js")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.js: %v", err)
	}
	jsStr := string(jsData)
	requiredJS := []string{
		"openSpecViewer",
		"closeSpecViewer",
		"approveSpec",
		"copySpecPath",
		"handleApproveSpecClick",
		"renderMarkdown",
		"/api/spec/approve",
		"Aprovar Spec (FLX-01)",
		"vscode://",
		"window.copySpecPath = copySpecPath;",
		"window.handleApproveSpecClick = handleApproveSpecClick;",
	}
	for _, req := range requiredJS {
		if !strings.Contains(jsStr, req) {
			t.Errorf("studio.js não contém função/recurso obrigatório: %s", req)
		}
	}
}

func TestStudioGovernanceAndPillars(t *testing.T) {
	// 1. Validar Markup no index.html
	htmlData, err := assetsFS.ReadFile("assets/index.html")
	if err != nil {
		t.Fatalf("falha ao ler assets/index.html: %v", err)
	}
	htmlStr := string(htmlData)
	requiredHTML := []string{
		`id="govSection"`,
		`id="govFindingCard"`,
		`id="govPillarsGrid"`,
		`Parecer da Auditoria (Camada B)`,
		`Pilares Constitucionais Verificados`,
	}
	for _, req := range requiredHTML {
		if !strings.Contains(htmlStr, req) {
			t.Errorf("index.html não contém trecho obrigatório de governança: %s", req)
		}
	}

	// 2. Validar CSS em studio.css
	cssData, err := assetsFS.ReadFile("assets/studio.css")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.css: %v", err)
	}
	cssStr := string(cssData)
	requiredCSS := []string{
		".gov-pillars-grid",
		".gov-pillar-card",
		".gov-pillar-card.falha",
		".gov-pillar-card.aprovado",
		".gov-pillar-chip",
		".finding-item.aprovado",
		".finding-item.falha",
	}
	for _, req := range requiredCSS {
		if !strings.Contains(cssStr, req) {
			t.Errorf("studio.css não contém estilo obrigatório de governança: %s", req)
		}
	}

	// 3. Validar JS em studio.js
	jsData, err := assetsFS.ReadFile("assets/studio.js")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.js: %v", err)
	}
	jsStr := string(jsData)
	requiredJS := []string{
		"CONSTITUTIONAL_PILLARS",
		"HUMAN_RULE_BADGES",
		"formatRuleBadge",
		"formatTaskStatus",
		"govPillarsGrid",
		"govFindingCard",
		"FALHA DETECTADA NA AUDITORIA",
		"AUDITORIA APROVADA (CONFORME)",
		"APROVADO PELA GOVERNANÇA",
		"FALHA APONTADA",
		"✓ CONCLUÍDO & APROVADO",
		"⏳ NA FILA (AGUARDANDO INÍCIO)",
		"⚡ EM ANDAMENTO",
		"🔍 AGUARDANDO AUDITORIA",
		"⚠️ RETRABALHO SOLICITADO",
	}
	for _, req := range requiredJS {
		if !strings.Contains(jsStr, req) {
			t.Errorf("studio.js não contém recurso/termo obrigatório: %s", req)
		}
	}
}

func TestStudioBadgeReactivityAndIdeSecurity(t *testing.T) {
	// 1. Validar JS em studio.js
	jsData, err := assetsFS.ReadFile("assets/studio.js")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.js: %v", err)
	}
	jsStr := string(jsData)

	requiredJS := []string{
		"station-badges-",
		"studio_project_dir",
		"activeSpecTask",
		"activePlannerSpecTask",
		"queue-task-name",
	}
	for _, req := range requiredJS {
		if !strings.Contains(jsStr, req) {
			t.Errorf("studio.js não contém trecho obrigatório: %s", req)
		}
	}

	// Garantir que nenhum card contenha link de texto redundante 'Ver histórico ->'
	if strings.Contains(jsStr, "Ver histórico →") || strings.Contains(jsStr, "Ver histórico ->") {
		t.Errorf("studio.js não deve conter link redundante 'Ver histórico ->'")
	}

	// 2. Validar CSS em studio.css
	cssData, err := assetsFS.ReadFile("assets/studio.css")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.css: %v", err)
	}
	cssStr := string(cssData)

	requiredCSS := []string{
		".queue-task-row",
		".queue-task-name",
		"text-overflow: ellipsis",
	}
	for _, req := range requiredCSS {
		if !strings.Contains(cssStr, req) {
			t.Errorf("studio.css não contém regra obrigatória: %s", req)
		}
	}
}

func TestStudioAssetsAutoRunContract(t *testing.T) {
	// 1. Validar Markup no index.html
	htmlData, err := assetsFS.ReadFile("assets/index.html")
	if err != nil {
		t.Fatalf("falha ao ler assets/index.html: %v", err)
	}
	htmlStr := string(htmlData)
	if !strings.Contains(htmlStr, `id="autoRunChip"`) {
		t.Errorf("index.html deve conter id=\"autoRunChip\"")
	}
	if !strings.Contains(htmlStr, "MODO AUTÔNOMO") {
		t.Errorf("index.html deve conter 'MODO AUTÔNOMO'")
	}

	// 2. Validar CSS em studio.css
	cssData, err := assetsFS.ReadFile("assets/studio.css")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.css: %v", err)
	}
	cssStr := string(cssData)
	requiredCSS := []string{
		".chip.autorun-chip",
		".autorun-running-chip",
		".autorun-card-badge",
		"@keyframes autorunGlow",
		"@keyframes boltPulse",
	}
	for _, req := range requiredCSS {
		if !strings.Contains(cssStr, req) {
			t.Errorf("studio.css não contém regra de auto-run obrigatória: %s", req)
		}
	}

	// 3. Validar JS em studio.js
	jsData, err := assetsFS.ReadFile("assets/studio.js")
	if err != nil {
		t.Fatalf("falha ao ler assets/studio.js: %v", err)
	}
	jsStr := string(jsData)
	if !strings.Contains(jsStr, "⚡ Auto-Run: Executando automaticamente") {
		t.Errorf("studio.js deve conter a indicação '⚡ Auto-Run: Executando automaticamente'")
	}
}
