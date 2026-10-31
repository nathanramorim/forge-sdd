# feat-63-03 — Visualizador de Spec Embutido, Aprovação Humana no Portão FLX-01 e Sincronização em Tempo Real via SSE

**Branch:** `feat/63-studio-navegavel-aprovacao-spec`
**Fase:** 63-03
**Tamanho:** `M`
**Depende de:** `feat-63-01-api-spec-e-sanitizacao`, `feat-63-02-drawer-inspecao-estacoes-e-deep-link`
**Status:** `todo`

## Objetivo

Concluir o cockpit interativo do Forge Studio integrando o visualizador de especificações embutido, a formalização visual e operacional do **Portão FLX-01 (Revisão Humana de Spec)** e a sincronização instantânea em tempo real via Server-Sent Events (SSE):
1. **Componente Visualizador de Spec (`.spec-viewer-modal`):**
   - Modal embutido no Studio com backdrop escurecido e fechamento via botão `x`, backdrop ou `ESC`.
   - Requisita o conteúdo via endpoint seguro `GET /api/spec?path=...` implementado em `feat-63-01`.
   - Renderizador semântico de Markdown embutido no client (`studio.js`):
     - Cabeçalhos (`#`, `##`, `###`) formatados.
     - Tabelas de reuso (`DUP-02`) com alinhamento visual dark theme.
     - Blocos de código monoespaçados com visual de terminal.
     - Checkboxes e checklists executáveis interativos.
   - Cabeçalho com metadados (caminho da feature, versão, estação de origem) e botões de ação:
     - "Abrir no VS Code / IDE" (`vscode://file/<abs_path>`).
     - "Aprovar Spec (Portão FLX-01)" com destaque visual (quando pendente de revisão).
2. **Destaque Visual do Portão FLX-01 (Estação 02 e Trilho de Handoff):**
   - No agregador de estado (`internal/studio/state.go`), detectar tarefas em `pm-architect` com status `pending_review` ou `approval`.
   - Na Estação 02 (PM & Arquiteto):
     - Exibir badge pulsante de destaque: **"Aguardando Revisão Humana (FLX-01)"**.
     - Colocar o mascote 3D da estação no estado visual `approval`.
   - No Trilho de Handoff superior:
     - Destacar a transição entre `02. PM & Arquiteto` e `03. O Planejador` com pacote em pausa e badge indicador "Portão FLX-01 Bloqueado".
3. **Endpoint Seguro de Aprovação Humana (`POST /api/spec/approve`):**
   - Endpoint HTTP em `internal/studio/server.go`:
     - Recebe payload JSON `{ "spec_path": "...", "station": "pm-architect", "task_id": "..." }`.
     - Validação de entrada na borda (`SEG-01`) e verificação de integridade da spec.
     - Invoca o motor de filas (`internal/queue/stations.go`) para atualizar atomicamente o status da tarefa para `approved` com registro de idempotência.
     - Notifica o `SSEHub`, disparando broadcast imediato de `event: state` para todos os navegadores conectados.
4. **Sincronização em Tempo Real via SSE (Sem Reload):**
   - Ao confirmar a aprovação, a UI recebe o evento SSE em menos de 500ms (`CA-06`).
   - O badge da Estação 02 transiciona para `Aprovado (FLX-01 Superado)`, o Trilho de Handoff libera o fluxo para a Estação 03 e a gaveta/modal reflete o novo estado instantaneamente sem reload (`F5`).
5. **Testes Integrados com Concorrência Segura (`-race`):**
   - Teste automatizado do endpoint `POST /api/spec/approve` garantindo transição de estado na fila, idempotência e ausência de concorrência (`internal/studio/server_test.go`).
   - Teste de notificação SSE assegurando emissão correta de snapshots.
   - Execução com race detector (`go test -race ./...`).

## Modelo do módulo
- **Modelo:** `enxuto` nos pacotes [`internal/studio`](file:///Users/nathanramorim/git/forge-sdd/internal/studio) e [`internal/queue`](file:///Users/nathanramorim/git/forge-sdd/internal/queue).
- **Justificativa:** Conecta os contratos de fila existentes ao visualizador web e ao canal de eventos SSE, mantendo zero dependências externas (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Motor de Filas e Handoff | `internal/queue/stations.go` | `LoadStationQueue`, `SaveStationQueue`, escrita atômica | `reusar` | Invocar transições canônicas de fila preservando idempotência (`DUP-01`) |
| Servidor HTTP Studio | `internal/studio/server.go` | Mux e middleware HTTP | `estender` | Adicionar rota controlada `POST /api/spec/approve` |
| Agregador de Estado | `internal/studio/state.go` | `StudioSnapshot` e `StationDTO` | `estender` | Adicionar sinalizador de spec pendente de aprovação humana no snapshot |
| Stream SSE | `internal/studio/sse.go` | `SSEHub` e `Broadcast` | `reusar` | Disparar `hub.Broadcast()` imediatamente ao aprovar a spec |
| Front-end Studio | `internal/studio/assets/` | Componentes web embutidos | `estender` | Criar `.spec-viewer-modal`, renderizador Markdown e handlers de aprovação |

## Dependências novas (`SEG-07`)
Nenhuma dependência externa. 100% Go stdlib (`net/http`, `sync`, `time`, `encoding/json`, `embed`) e JS vanilla no navegador.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`
- **IDs:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-04`, `DUP-01`, `DUP-02`, `COD-01`, `COD-02`, `TST-01`, `TST-03`, `TST-04`, `TST-05`, `FLX-01`, `FLX-04`, `FLX-06`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `internal/studio/server.go`
- `internal/studio/server_test.go`
- `internal/studio/state.go`
- `internal/studio/assets/index.html`
- `internal/studio/assets/studio.css`
- `internal/studio/assets/studio.js`
- `internal/studio/assets_test.go`

## Critérios de Aceitação Executáveis
1. **CA-05 & CA-06 (Portão FLX-01, Aprovação de Spec e Tempo Real SSE):**
   ```bash
   go test -v -race -run TestAPISpecApprove ./internal/studio/... && go vet ./...
   ```
2. **Suíte Completa de Testes Concorrentes (`-race`):**
   ```bash
   go test -race ./internal/studio/... ./internal/queue/... && go vet ./...
   ```
