# Plano de Discovery 65 — Pre-warming do Pool de Sessões ACP e Orquestração Contínua de Estações

Roadmap preliminar, arquitetura de fatiamento e mapeamento de governança para execução da iniciativa `feat-65-prewarming-pool-sessoes-acp-estacoes`.

A execução será agrupada na subpasta única `sdd/features/feat-65-prewarming-pool-sessoes-acp-estacoes/` (Regra 13), utilizando uma **única branch de trabalho** (`feat/65-prewarming-pool-sessoes-acp-estacoes`, conforme Regra 15 e `FLX-02`). O fatiamento respeita a **Rule of 7** (4 subtarefas de baixo para cima com escopos claros e não sobrepostos).

---

## 1. Quebra de Tarefas / Subfeatures (`P / M / G`)

### Fase 1 — Núcleo do SessionPool em Go stdlib, Pre-warming das 5 Sessões e Heartbeat

- [ ] **`feat-65-01` — Núcleo do `SessionPool` (`internal/acp/pool.go`), Pre-warming no Boot e Monitor de Heartbeat (`internal/acp/heartbeat.go`)** · **Tamanho: `M`**
  - **Depende de:** — (reutiliza `internal/acp` da Feat 64)
  - **Escopo:**
    1. Criar o tipo `SessionPool` em `internal/acp/pool.go` gerenciando o ciclo de vida das 5 sessões canônicas (`sess-01-lider`, `sess-02-pm`, `sess-03-planner`, `sess-04-builder`, `sess-05-governance`).
    2. Implementar o método `Warmup(ctx, workspaceDir)` com validação de borda e confinamento de segurança (`SEG-01`, `SEG-02`).
    3. Implementar a máquina de estados operacionais (`StateStandby`, `StateProcessing`, `StateApprovalPending`, `StateError`, `StateOffline`).
    4. Implementar o `HeartbeatMonitor` em `internal/acp/heartbeat.go` executando verificações periódicas não-bloqueantes (`tick`) com suporte a cancelamento via context (`ARQ-03`).
    5. Registrar os métodos RPC `pool/warmup` e `pool/status` no `RPCDispatcher`.
    6. Testes unitários com concorrência segura (`go test -race`) cobrindo aquecimento, consulta de status e transição de estados.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestSessionPool" ./internal/acp/... && go vet ./...
    ```

---

### Fase 2 — Orquestrador de Handoff Contínuo via ACP e Sincronização com Portão FLX-01

- [ ] **`feat-65-02` — Orquestrador de Handoff Contínuo (`internal/acp/orchestrator.go`), Despacho Automático de Filas e Integração com Portão FLX-01** · **Tamanho: `M`**
  - **Depende de:** `feat-65-01`
  - **Escopo:**
    1. Criar o `HandoffOrchestrator` em `internal/acp/orchestrator.go` ouvindo eventos de novas tasks e handoffs de `internal/queue`.
    2. Implementar a transição contínua e assíncrona entre estações:
       - Estação 02 (PM & Arquiteto) conclui Spec → entra em `StateApprovalPending` com bloqueio obrigatório.
       - Resolução do Portão FLX-01 via `gate/approve` → valida regras estruturais `P-02` e `P-03` (`internal/gate`).
       - Desbloqueio e acionamento automático da Estação 03 (Planejador) para `StateProcessing`.
    3. Emitir notificações estruturadas `session/notification` propagando mudanças de estado e progresso para o Studio e IDEs.
    4. Testes de integração cobrindo o fluxo completo de handoff e validação de bloqueio estrito quando o portão não foi aprovado.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestHandoffOrchestrator" ./internal/acp/... ./internal/queue/... && go vet ./...
    ```

---

### Fase 3 — Subcomando CLI `forge-sdd attach [estacao]` e Bridge de Terminal Interativo

- [ ] **`feat-65-03` — Subcomando CLI `forge-sdd attach` (`cmd/forge-sdd/attach.go`) e Multiplexador de Terminal (`internal/acp/attach.go`)** · **Tamanho: `M`**
  - **Depende de:** `feat-65-01`, `feat-65-02`
  - **Escopo:**
    1. Criar `internal/acp/attach.go` implementando a abstração `AttachBridge` com suporte a múltiplos leitores/escritores concorrentes (`sync.RWMutex`).
    2. Implementar método RPC `session/attach` retornando o estado, histórico de eventos da estação e canal de streaming bidirecional.
    3. Criar o subcomando CLI Cobra `cmd/forge-sdd/attach.go` permitindo conectar o terminal diretamente à sessão de qualquer uma das 5 estações:
       ```bash
       forge-sdd attach builder
       forge-sdd attach pm-architect
       ```
    4. Implementar tratamento gracioso de encerramento (`Ctrl+C` ou comando `exit`) garantindo que a desconexão do cliente não encerre a sessão subjacente no pool.
    5. Testes unitários e de CLI simulando conexões, envio de prompts interativos e desconexão controlada.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestAttach" ./cmd/forge-sdd/... ./internal/acp/... && go vet ./...
    ```

---

### Fase 4 — Integração com Forge Studio (HTTP/SSE), Visualizador de Estados e Testes E2E

- [ ] **`feat-65-04` — Integração do Pool com Forge Studio (`internal/studio/server.go`), Atualização de Mascotes e Validação E2E** · **Tamanho: `P`**
  - **Depende de:** `feat-65-01`, `feat-65-02`, `feat-65-03`
  - **Escopo:**
    1. Integrar o `SessionPool` ao bootstrap do Forge Studio (`internal/studio/server.go`), disparando o pre-warming automático no arranque do servidor.
    2. Conectar as transições de estado do pool ao `SSEHub`, permitindo que os mascotes 3D e cards das estações atualizem seus indicadores visuais instantaneamente (🟢 Standby, ⚡ Processing, ⏸️ Approval, 🔴 Error).
    3. Criar teste E2E ponta a ponta simulando o ciclo completo: Boot do Studio → Pre-warming das 5 estações em Standby → Despacho para Estação 02 → Portão FLX-01 → Handoff para Estação 03 → Attach via CLI.
    4. Atualizar documentação e telemetria (`sdd/memory/progress.md`).
  - **Critério Executável:**
    ```bash
    go test -v -race ./internal/acp/... ./internal/studio/... ./cmd/forge-sdd/... && go vet ./...
    ```

---

## 2. Tabela de Mapeamento de Rules e IDs da Constituição

| Task | Tamanho | Arquivos `.agents/rules/` Aplicáveis | IDs da Constituição Cobertos | Arquivos de Escopo Principal | Critério Executável |
|---|---|---|---|---|---|
| **`feat-65-01`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md` | `ARQ-01`, `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-07`, `COD-01..03`, `TST-01`, `TST-03..05` | `internal/acp/pool.go`, `internal/acp/heartbeat.go`, `internal/acp/dispatcher.go`, `internal/acp/pool_test.go` | `go test -v -race -run "TestSessionPool" ./internal/acp/... && go vet ./...` |
| **`feat-65-02`** | `M` | `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md` | `ARQ-01`, `ARQ-02`, `SEG-01`, `DUP-01`, `DUP-02`, `FLX-01`, `FLX-04`, `COD-01..03`, `TST-01..05` | `internal/acp/orchestrator.go`, `internal/queue/stations.go`, `internal/gate/gate.go`, `internal/acp/orchestrator_test.go` | `go test -v -race -run "TestHandoffOrchestrator" ./internal/acp/... ./internal/queue/... && go vet ./...` |
| **`feat-65-03`** | `M` | `sdd-arquitetura.md`, `sdd-ergonomia.md`, `sdd-codigo-e-testes.md` | `ARQ-01`, `ARQ-06`, `SEG-01`, `SEG-04`, `COD-01..03`, `TST-01..05` | `cmd/forge-sdd/attach.go`, `internal/acp/attach.go`, `cmd/forge-sdd/main.go`, `cmd/forge-sdd/commands_test.go` | `go test -v -race -run "TestAttach" ./cmd/forge-sdd/... ./internal/acp/... && go vet ./...` |
| **`feat-65-04`** | `P` | `sdd-arquitetura.md`, `sdd-governanca.md`, `sdd-codigo-e-testes.md` | `ARQ-03`, `ARQ-04`, `FLX-02`, `FLX-03`, `GOV-04`, `TST-01`, `TST-05` | `internal/studio/server.go`, `internal/acp/e2e_pool_test.go`, `sdd/memory/progress.md` | `go test -v -race ./internal/acp/... ./internal/studio/... ./cmd/forge-sdd/... && go vet ./...` |

---

## 3. Matriz de Rastreabilidade Constitucional

- **`ARQ-01` & `ARQ-02` (Fronteiras e Separação de Camadas):** O comando `cmd/forge-sdd/attach.go` funciona exclusivamente como camada de apresentação CLI, delegando toda a lógica de sessão, roteamento e multiplexação ao pacote `internal/acp`.
- **`ARQ-03` & `ARQ-04` (Binário Estático & Zero Dependências Externas):** Implementação baseada unicamente na biblioteca padrão do Go, sem dependência de runtimes externos ou pacotes de terceiros.
- **`SEG-01` & `SEG-02` (Segurança de Workspace e Path Traversal):** Todos os caminhos de workspace passados para inicialização do pool ou attach passam por sanitização rígida via `filepath.Clean` e verificação de fronteiras seguras.
- **`SEG-07` (Zero Dependências Novas):** Não é adicionada nenhuma nova biblioteca externa ao `go.mod`.
- **`DUP-01` & `DUP-02` (Anti-Duplicação e Reúso):** Reúso estrito dos modelos de fila em `internal/queue`, do mecanismo de portões em `internal/gate` e do despachante em `internal/acp`.
- **`FLX-01` (Portão Humano Obrigatório):** O handoff contínuo bloqueia obrigatoriamente após a conclusão da Estação 02, exigindo validação explícita humana antes de permitir que a Estação 03 inicie a decomposição.
- **`FLX-02` (Subpasta e Branch Únicas):** Todo o trabalho de execução será mantido sob a branch `feat/65-prewarming-pool-sessoes-acp-estacoes` e pasta correspondente.
- **`TST-01` a `TST-05` (Testes Determinísticos e Race Detector):** Todas as fases contam com suítes de testes automatizados executáveis com a flag `-race`, garantindo ausência de concorrência espúria no gerenciamento das 5 sessões.
