# Plano de Discovery 63 — Studio Navegável, Visualizador de Spec com Aprovação Humana (FLX-01) e Deep-Link de IDE

Roadmap técnico e quebra em subtarefas para o fatiamento e execução dentro da subpasta única `sdd/features/feat-63-studio-navegavel-aprovacao-spec/` (Regra 13), operando sobre uma **única branch** (`feat/63-studio-navegavel-aprovacao-spec`, conforme Regra 15 e `FLX-02`).

Respeita a **Rule of 7** (3 subtarefas determinísticas ordenadas de baixo para cima), incorpora as decisões aprovadas no Portão `FLX-01` e reutiliza os pacotes canônicos `internal/studio` e `internal/queue`.

---

## 1. Quebra de Tasks / Subfeatures (`P / M / G`)

### Fase 1 — Backend & Segurança: Endpoint de Leitura de Spec e Sanitização Anti-Traversal

- [ ] **`feat-63-01` — Endpoint de Leitura de Spec (`GET /api/spec`) e Sanitização Estrita contra Path Traversal** · **Tamanho: `P`**
  - **Depende de:** —
  - **Escopo:**
    1. Implementar o endpoint `GET /api/spec?path=<path>` em `internal/studio/server.go`.
    2. Aplicar sanitização na borda com `filepath.Clean` e verificação de confinamento ao diretório do projeto (`s.Dir`) e diretório SDD (`sdd/`).
    3. Rejeitar com `400 Bad Request` ou `403 Forbidden` qualquer tentativa de path traversal (`..`, caminhos absolutos arbitrários, `/etc/`, etc.) em conformidade com `SEG-01`, `SEG-02` e `SEG-04`.
    4. Retornar payload JSON determinístico contendo `path`, `absolute_path`, `content`, `status`, `station` e `feature_id`.
    5. Testes unitários com concorrência segura cobrindo cenários conformes e negativos (`internal/studio/server_test.go`).
  - **Critério Executável:** `go test -v -race -run TestAPISpec ./internal/studio/... && go vet ./...`

---

### Fase 2 — UI Interativa: Gaveta Lateral de Inspeção das Estações e Deep-Link de IDE

- [ ] **`feat-63-02` — Componente Station Drawer (`.station-drawer`), Cartões Clicáveis e Deep-Link para IDE Local** · **Tamanho: `M`**
  - **Depende de:** `feat-63-01`
  - **Escopo:**
    1. Implementar a gaveta lateral deslizante (`.station-drawer`) nos assets embutidos (`internal/studio/assets/studio.css`, `studio.js`, `index.html`).
    2. Tornar os cartões das 5 estações no dashboard clicáveis, abrindo o Drawer correspondente.
    3. Exibir no Drawer contratos de I/O, fila detalhada de tarefas (`sdd/.queue/stations/<station>.json`), regras constitucionais aplicáveis (`ARQ-*`, `SEG-*`, `DUP-*`, `FLX-*`) e histórico recente de ferramentas.
    4. Implementar o botão Deep-Link para IDE (`vscode://file/<abs_path>` e `cursor://file/<abs_path>`).
    5. Testes unitários em `internal/studio/assets_test.go` verificando elementos, acessibilidade e integridade do markup.
  - **Critério Executável:** `go test -v -run TestStudioAssetsDrawerAndLinks ./internal/studio/... && go vet ./...`

---

### Fase 3 — Visualizador de Spec, Portão FLX-01 com Aprovação Humana e Sincronização SSE

- [ ] **`feat-63-03` — Visualizador de Spec Embutido, Aprovação Humana no Portão FLX-01 e Sincronização em Tempo Real via SSE** · **Tamanho: `M`**
  - **Depende de:** `feat-63-01`, `feat-63-02`
  - **Escopo:**
    1. Implementar modal/leitor embutido de Spec (`.spec-viewer-modal`) com renderização semântica de Markdown (títulos, tabelas, código e checklist).
    2. Destaque visual da fase de aprovação humana da Spec na Estação 02 (badge pulsante, mascote em `approval`) e no Trilho de Handoff (Portão FLX-01 pendente).
    3. Endpoint seguro `POST /api/spec/approve` que valida critérios do portão, atualiza atomicamente a fila da estação para `approved` com idempotência e aciona o `SSEHub`.
    4. Sincronização instantânea via SSE (`event: state`) sem reload do navegador (`F5`).
    5. Testes integrados com concorrência segura (`go test -race`).
  - **Critério Executável:** `go test -v -race ./internal/studio/... ./internal/queue/... && go vet ./...`

---

## 2. Tabela de Mapeamento de Rules e IDs da Constituição

| Task | Tamanho | Arquivos `.agents/rules/` Aplicáveis | IDs da Constituição Cobertos | Arquivos de Escopo | Critério Executável |
|---|---|---|---|---|---|
| **`feat-63-01`** | `P` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md` | `ARQ-01`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `COD-01`, `COD-02`, `TST-01`, `TST-03..05` | `internal/studio/server.go`, `internal/studio/server_test.go` | `go test -v -race -run TestAPISpec ./internal/studio/... && go vet ./...` |
| **`feat-63-02`** | `M` | `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md` | `ARQ-03`, `ARQ-04`, `DUP-02`, `COD-01`, `COD-05`, `TST-01`, `TST-04`, `TST-05` | `internal/studio/assets/index.html`, `internal/studio/assets/studio.css`, `internal/studio/assets/studio.js`, `internal/studio/assets_test.go` | `go test -v -run TestStudioAssetsDrawerAndLinks ./internal/studio/... && go vet ./...` |
| **`feat-63-03`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md` | `ARQ-01`, `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `DUP-01`, `DUP-02`, `COD-01`, `COD-02`, `TST-01..05`, `FLX-01`, `FLX-04`, `FLX-06` | `internal/studio/server.go`, `internal/studio/server_test.go`, `internal/studio/state.go`, `internal/studio/assets/*` | `go test -v -race ./internal/studio/... ./internal/queue/... && go vet ./...` |
