# Plano de Execução — feat-63: Studio Navegável, Visualizador de Spec com Aprovação Humana (FLX-01) e Deep-Link de IDE

Plano de fatiamento determinístico e roadmap de implementação da funcionalidade `feat-63-studio-navegavel-aprovacao-spec`, em conformidade com o papel de Planejador (`.agents/roles/planner.md`), a Constituição (`sdd/memory/constitution.md`) e os critérios técnicos aprovados no Portão FLX-01 (`criteria.md` e `discovery.md`).

Todas as subtarefas operam sobre uma **única branch** (`feat/63-studio-navegavel-aprovacao-spec`, conforme Regra 15 e `FLX-02`) e são agrupadas dentro da subpasta `sdd/features/feat-63-studio-navegavel-aprovacao-spec/` (Regra 13).

---

## 1. Quebra de Tasks / Subfeatures (`P / M / G`) — Rule of 7

A quebra foi estruturada em 3 tarefas ordenadas de baixo para cima (backend seguro → UI interativa e deep-links → visualizador, aprovação no portão e tempo real SSE):

```
feat-63-01 (Backend API de Spec + Sanitização Anti-Traversal SEG-01/02/04) [P]
    │
    ▼
feat-63-02 (Drawer de Inspeção das 5 Estações + Deep-Link IDE + UI) [M]
    │
    ▼
feat-63-03 (Visualizador Markdown + Portão FLX-01 Aprovação Humana + SSE) [M]
```

---

### Fase 1 — Backend & Segurança: Endpoint de Leitura de Spec e Sanitização Anti-Traversal

- [x] **`feat-63-01` — Endpoint de Leitura de Spec (`GET /api/spec`) e Sanitização Estrita contra Path Traversal** · **Tamanho: `P`**
  - **Branch:** `feat/63-studio-navegavel-aprovacao-spec`
  - **Depende de:** —
  - **Objetivo & Escopo:**
    1. Implementar o endpoint `GET /api/spec?path=<path>` em `internal/studio/server.go`.
    2. Aplicar validação estrita na borda (`SEG-01`, `SEG-02`, `SEG-04`):
       - Sanitizar o caminho com `filepath.Clean` e resolver o caminho absoluto baseado no diretório raiz do projeto (`s.Dir`).
       - Rejeitar com `400 Bad Request` ou `403 Forbidden` qualquer tentativa de evasão de diretório (`..`, prefixos externos como `/etc/`, `/var/`, caminhos fora de `s.Dir`).
       - Restringir a leitura estritamente a arquivos de documentação e specs dentro do escopo SDD (`sdd/features/`, `sdd/discovery/`, `sdd/spec/`).
       - Retornar `404 Not Found` caso o arquivo não exista no disco.
    3. Retornar payload JSON determinístico contendo:
       - `path`: caminho relativo normalizado (ex: `sdd/features/feat-63-.../discovery.md`).
       - `absolute_path`: caminho absoluto no filesystem local para deep-link.
       - `content`: conteúdo integral do Markdown.
       - `status`: status atual da spec/tarefa (`pending_review`, `approved`, etc.).
       - `station`: estação responsável (ex: `pm-architect`).
       - `feature_id`: identificador canônico da feature.
    4. Implementar testes unitários determinísticos em `internal/studio/server_test.go` cobrindo sucesso, path traversal negativo e inexistência de arquivo (`TST-01`, `TST-03`, `TST-04`).
  - **Rules Aplicáveis:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`
  - **IDs da Constituição:** `ARQ-01`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `COD-01`, `COD-02`, `TST-01`, `TST-03`, `TST-04`, `TST-05`
  - **Arquivos de Escopo:**
    - `internal/studio/server.go`
    - `internal/studio/server_test.go`
  - **Critério Executável:**
    ```bash
    go test -v -race -run TestAPISpec ./internal/studio/... && go vet ./...
    ```

---

### Fase 2 — UI Interativa: Gaveta Lateral de Inspeção das Estações e Deep-Link de IDE

- [ ] **`feat-63-02` — Componente Station Drawer (`.station-drawer`), Cartões Clicáveis e Deep-Link para IDE Local** · **Tamanho: `M`**
  - **Branch:** `feat/63-studio-navegavel-aprovacao-spec`
  - **Depende de:** `feat-63-01`
  - **Objetivo & Escopo:**
    1. Implementar o componente de gaveta lateral deslizante (`.station-drawer`) em `internal/studio/assets/studio.css`, `internal/studio/assets/studio.js` e `internal/studio/assets/index.html`.
    2. Tornar os 5 cartões de estação clicáveis no grid do Studio com indicação de foco, cursor `pointer` e acessibilidade (`role="button"`, `aria-expanded`).
    3. A gaveta deve exibir seções detalhadas:
       - **Cabeçalho:** Nome, número, cor temática, marcação da estação e resumo do papel canônico (`.agents/roles/`).
       - **Contratos:** `Recebe (Entrada)` e `Entrega (Saída)` claros.
       - **Lista Detalhada de Tarefas:** Itens da fila (`sdd/.queue/stations/<station>.json`), id, status, tentativa, idempotency key e regras constitucionais aplicáveis (`ARQ-*`, `SEG-*`, etc.).
       - **Uso Recente de Ferramentas:** Histórico filtrado de comandos executados pela estação com timestamps.
    4. Ao clicar em uma tarefa ou artefato, disponibilizar o botão de Deep-Link para a IDE:
       - Link nativo `vscode://file/<abs_path>` para abertura direta no VS Code.
       - Link secundário `cursor://file/<abs_path>` e botão de cópia do comando `code <path>`.
       - Botão "Ver Especificação" disparando a navegação para o leitor de spec.
    5. Acessibilidade e ergonomia: fechamento via botão `x`, clique no backdrop escurecido (`.drawer-backdrop`) ou tecla `ESC`. Layout responsivo sem quebra de pills (`white-space: nowrap; flex-shrink: 0` conforme `CA-10`).
    6. Testes em `internal/studio/assets_test.go` verificando elementos, classes e integridade do markup embutido via `embed.FS` (`ARQ-03`).
  - **Rules Aplicáveis:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`
  - **IDs da Constituição:** `ARQ-03`, `ARQ-04`, `DUP-02`, `COD-01`, `COD-05`, `TST-01`, `TST-04`, `TST-05`
  - **Arquivos de Escopo:**
    - `internal/studio/assets/index.html`
    - `internal/studio/assets/studio.css`
    - `internal/studio/assets/studio.js`
    - `internal/studio/assets_test.go`
  - **Critério Executável:**
    ```bash
    go test -v -run TestStudioAssetsDrawerAndLinks ./internal/studio/... && go vet ./...
    ```

---

### Fase 3 — Visualizador de Spec, Portão FLX-01 com Aprovação Humana e Sincronização SSE

- [ ] **`feat-63-03` — Visualizador de Spec Embutido, Aprovação Humana no Portão FLX-01 e Sincronização em Tempo Real via SSE** · **Tamanho: `M`**
  - **Branch:** `feat/63-studio-navegavel-aprovacao-spec`
  - **Depende de:** `feat-63-01`, `feat-63-02`
  - **Objetivo & Escopo:**
    1. Criar o componente modal/painel `.spec-viewer-modal` em `internal/studio/assets/` com renderizador semântico de Markdown (cabeçalhos, tabelas de reuso, blocos de código com destaque, badges de status e checklist executável).
    2. Implementar a detecção e destaque do **Portão FLX-01 (Revisão Humana de Spec)**:
       - Quando a fila `sdd/.queue/stations/pm-architect.json` contiver item em `pending_review` ou em elaboração:
         - A Estação 02 (PM & Arquiteto) exibe badge pulsante **"Aguardando Revisão Humana (FLX-01)"** e mascote em estado visual `approval`.
         - O Trilho de Handoff superior posiciona o pacote entre `02` e `03` com badge visual "Portão FLX-01 Pendente".
    3. Ação Segura de Aprovação Humana:
       - Endpoint `POST /api/spec/approve` em `internal/studio/server.go`:
         - Recebe payload `{ "spec_path": "...", "station": "pm-architect", "task_id": "..." }`.
         - Valida existência da spec e conformidade com critérios de gate (`internal/gate`).
         - Atualiza atomicamente a fila da estação para `approved` com idempotência (`internal/queue/stations.go`).
         - Dispara broadcast no `SSEHub`, emitindo evento `event: state` para os clientes conectados.
       - Na interface web, botão destacado **"Aprovar Spec (Portão FLX-01)"** que invoca a rota via `fetch` assíncrono.
    4. Sincronização reativa em tempo real:
       - Ao receber o evento SSE, a interface web atualiza instantaneamente os cards, o drawer, o Handoff e o status do portão em menos de 500ms, sem reload (`F5`).
    5. Testes completos integrados com verificação de concorrência e race detector (`go test -race`).
  - **Rules Aplicáveis:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`
  - **IDs da Constituição:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `DUP-01`, `DUP-02`, `COD-01`, `COD-02`, `TST-01`, `TST-03`, `TST-04`, `TST-05`, `FLX-01`, `FLX-04`, `FLX-06`
  - **Arquivos de Escopo:**
    - `internal/studio/server.go`
    - `internal/studio/server_test.go`
    - `internal/studio/state.go`
    - `internal/studio/assets/index.html`
    - `internal/studio/assets/studio.css`
    - `internal/studio/assets/studio.js`
    - `internal/studio/assets_test.go`
  - **Critério Executável:**
    ```bash
    go test -v -race ./internal/studio/... ./internal/queue/... && go vet ./...
    ```

---

## 2. Janelas de Paralelismo e Dependências

```
Onda 1:
feat-63-01 (Backend API de Spec + Sanitização Anti-Traversal SEG-01/02/04) [P]
   │
   ▼
Onda 2:
feat-63-02 (Station Drawer + Cartões Clicáveis + Deep-Link IDE) [M]
   │
   ▼
Onda 3:
feat-63-03 (Visualizador Markdown + Portão FLX-01 Aprovação Humana + SSE) [M]
```

- **Sequenciamento determinístico:** `feat-63-01` fornece o contrato de API e os dados de caminho absoluto/conteúdo necessários para o deep-link e renderizador; `feat-63-02` entrega o container visual (Drawer) e os gatilhos nos cards das 5 estações; `feat-63-03` fecha o fluxo de ponta a ponta com o leitor modal, o Portão FLX-01, endpoint seguro de aprovação e broadcast SSE.

---

## 3. Tabela de Mapeamento de Rules e IDs da Constituição por Task

| Task | Tamanho | Arquivos `.agents/rules/` Aplicáveis | IDs da Constituição Cobertos | Arquivos de Escopo | Critério Executável |
|---|---|---|---|---|---|
| **`feat-63-01`** | `P` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md` | `ARQ-01`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `COD-01`, `COD-02`, `TST-01`, `TST-03..05` | `internal/studio/server.go`, `internal/studio/server_test.go` | `go test -v -race -run TestAPISpec ./internal/studio/... && go vet ./...` |
| **`feat-63-02`** | `M` | `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md` | `ARQ-03`, `ARQ-04`, `DUP-02`, `COD-01`, `COD-05`, `TST-01`, `TST-04`, `TST-05` | `internal/studio/assets/index.html`, `internal/studio/assets/studio.css`, `internal/studio/assets/studio.js`, `internal/studio/assets_test.go` | `go test -v -run TestStudioAssetsDrawerAndLinks ./internal/studio/... && go vet ./...` |
| **`feat-63-03`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md` | `ARQ-01`, `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `DUP-01`, `DUP-02`, `COD-01`, `COD-02`, `TST-01..05`, `FLX-01`, `FLX-04`, `FLX-06` | `internal/studio/server.go`, `internal/studio/server_test.go`, `internal/studio/state.go`, `internal/studio/assets/*` | `go test -v -race ./internal/studio/... ./internal/queue/... && go vet ./...` |

---

## 4. Handoff para o Líder

O plano de fatiamento determinístico da `feat-63-studio-navegavel-aprovacao-spec` está completo e alinhado com a Constituição (`sdd/memory/constitution.md`).
As tarefas individuais foram especificadas e estão prontas para aprovação do Líder e enfileiramento na Fila do Builder (`sdd/.queue/queue.json`).
