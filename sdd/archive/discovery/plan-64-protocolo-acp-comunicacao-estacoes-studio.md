# Plano de Discovery 64 — Integração do Protocolo ACP no Forge-SDD e Forge Studio

Roadmap preliminar, arquitetura de fatiamento e mapeamento de governança para execução da iniciativa `feat-64-protocolo-acp-comunicacao-estacoes-studio`.

A execução será agrupada na subpasta única `sdd/features/feat-64-protocolo-acp-comunicacao-estacoes-studio/` (Regra 13), utilizando uma **única branch de trabalho** (`feat/64-protocolo-acp-comunicacao-estacoes-studio`, conforme Regra 15 e `FLX-02`). O fatiamento respeita a **Rule of 7** (4 subtarefas de baixo para cima com escopos claros e não sobrepostos).

---

## 1. Quebra de Tarefas / Subfeatures (`P / M / G`)

### Fase 1 — Núcleo do Protocolo ACP em Go stdlib (Transporte Stdio & Mensagens JSON-RPC 2.0)

- [ ] **`feat-64-01` — Núcleo do Protocolo ACP (`internal/acp`), Tipos JSON-RPC 2.0 e Subcomando CLI `forge-sdd acp`** · **Tamanho: `M`**
  - **Depende de:** —
  - **Escopo:**
    1. Criar o pacote `internal/acp` com definição dos tipos padrão do protocolo: `Request`, `Response`, `Error`, `Notification` e códigos de erro padrão.
    2. Implementar o `RPCDispatcher` para roteamento de métodos (`initialize`, `session/new`, `session/load`, `session/cancel`).
    3. Implementar o `StdioTransport` baseado em `bufio.Scanner` e `io.Writer` concorrentemente seguro (`sync.Mutex`).
    4. Criar o subcomando CLI `cmd/acp.go` (`forge-sdd acp`) operando sobre o transporte `stdio` (`ARQ-01`, `ARQ-06`).
    5. Testes unitários com concorrência segura (`go test -race`) cobrindo parsing de mensagens, handshake e tratamento de erros malformados.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestACP" ./internal/acp/... && go vet ./...
    ```

---

### Fase 2 — Integração com as 5 Estações, Filas e Portão FLX-01

- [ ] **`feat-64-02` — Integração de Sessões ACP com Motor de Filas (`internal/queue`) e Aprovação de Portão (`gate/approve`)** · **Tamanho: `M`**
  - **Depende de:** `feat-64-01`
  - **Escopo:**
    1. Implementar o `SessionManager` em `internal/acp/session.go` integrando o ciclo de vida da sessão com `sdd/.queue/stations/` e `sdd/memory/progress.md`.
    2. Implementar o método `session/prompt` para enfileiramento determinístico de demandas na estação ativa via `internal/queue`.
    3. Implementar o método `gate/approve` para resolução remota do Portão FLX-01 via ACP, invocando validações estruturais de `internal/gate` (`P-02` e `P-03`) e transicionando a fila com escrita atômica.
    4. Implementar emissor de notificações JSON-RPC unidirecionais (`session/notification`) propagando mudanças de estado e chamadas de ferramentas.
    5. Testes de integração entre ACP, filas e portões cobrindo cenários válidos e violações.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestACPSessionAndGate" ./internal/acp/... ./internal/queue/... && go vet ./...
    ```

---

### Fase 3 — Transporte HTTP/SSE para Forge Studio e Streaming em Tempo Real

- [ ] **`feat-64-03` — Transporte ACP HTTP/SSE no Forge Studio (`internal/studio`) e Bridge com SSEHub** · **Tamanho: `M`**
  - **Depende de:** `feat-64-01`, `feat-64-02`
  - **Escopo:**
    1. Estender o servidor do Forge Studio (`internal/studio/server.go`) adicionando os endpoints `/api/acp/rpc` (POST para chamadas JSON-RPC) e `/api/acp/events` (GET para stream SSE).
    2. Conectar as notificações do `RPCDispatcher` ao `SSEHub` (`internal/studio/sse.go`), permitindo que a interface web do Studio e clientes remotos recebam eventos ACP sem polling.
    3. Validação rigorosa na borda (`SEG-01`, `SEG-02`) para evitar exposição indevida ou path traversal em requisições HTTP.
    4. Preservação integral do comportamento das rotas legadas do Studio (`/api/state`, `/api/spec`, etc.).
    5. Testes de concorrência com múltiplos clientes conectados via SSE.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestStudioACPTransport" ./internal/studio/... && go vet ./...
    ```

---

### Fase 4 — Configurações para IDEs (Zed, Cursor, VS Code) e Validação E2E

- [ ] **`feat-64-04` — Exemplos de Integração com IDEs (Zed / Cursor / VS Code), Documentação e Testes E2E** · **Tamanho: `P`**
  - **Depende de:** `feat-64-01`, `feat-64-02`, `feat-64-03`
  - **Escopo:**
    1. Criar documentação técnica de integração em `.agents/rules/sdd-acp.md` e referências de configuração para a IDE Zed (`zed: acp registry` ou `agent_servers` em `settings.json`).
    2. Disponibilizar snippet de configuração para extensões ACP no VS Code e Cursor.
    3. Criar teste E2E simulando um cliente IDE real via stdio interagindo com o Forge-SDD (handshake → new session → prompt → gate approval → finish).
    4. Verificação de métricas de telemetria e conformidade com `sdd/memory/constitution.md`.
  - **Critério Executável:**
    ```bash
    go test -v -race ./internal/acp/... ./internal/studio/... && go vet ./...
    ```

---

## 2. Tabela de Mapeamento de Rules e IDs da Constituição

| Task | Tamanho | Arquivos `.agents/rules/` Aplicáveis | IDs da Constituição Cobertos | Arquivos de Escopo Principal | Critério Executável |
|---|---|---|---|---|---|
| **`feat-64-01`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md` | `ARQ-01`, `ARQ-04`, `ARQ-06`, `SEG-01`, `SEG-04`, `SEG-07`, `COD-01..03`, `TST-01`, `TST-03..05` | `internal/acp/types.go`, `internal/acp/dispatcher.go`, `internal/acp/transport_stdio.go`, `cmd/acp.go`, `internal/acp/acp_test.go` | `go test -v -race -run "TestACP" ./internal/acp/... && go vet ./...` |
| **`feat-64-02`** | `M` | `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md` | `ARQ-01`, `ARQ-02`, `SEG-01`, `SEG-02`, `DUP-01`, `DUP-02`, `FLX-01`, `FLX-04`, `COD-01..03`, `TST-01..05` | `internal/acp/session.go`, `internal/acp/gate.go`, `internal/queue/stations.go`, `internal/acp/session_test.go` | `go test -v -race -run "TestACPSessionAndGate" ./internal/acp/... ./internal/queue/... && go vet ./...` |
| **`feat-64-03`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md` | `ARQ-01`, `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-05`, `DUP-02`, `COD-01`, `COD-02`, `TST-01..05` | `internal/studio/server.go`, `internal/studio/sse.go`, `internal/acp/transport_http.go`, `internal/studio/server_test.go` | `go test -v -race -run "TestStudioACPTransport" ./internal/studio/... && go vet ./...` |
| **`feat-64-04`** | `P` | `sdd-arquitetura.md`, `sdd-governanca.md`, `sdd-codigo-e-testes.md` | `ARQ-06`, `SEG-04`, `COD-05`, `FLX-02`, `FLX-03`, `GOV-04`, `TST-01`, `TST-05` | `.agents/rules/sdd-acp.md`, `internal/acp/e2e_test.go`, `README.md` | `go test -v -race ./internal/acp/... ./internal/studio/... && go vet ./...` |

---

## 3. Matriz de Rastreabilidade Constitucional

- **`ARQ-01` & `ARQ-02` (Fronteiras e Ciclos):** O subcomando `cmd/acp.go` atua apenas como CLI adapter fino; toda a lógica do protocolo reside em `internal/acp`, consumindo `internal/queue` e `internal/gate` sem dependências circulares e sem importar `cmd/`.
- **`ARQ-03` & `ARQ-04` (Embed & Zero Deps de Runtime):** Todo o protocolo é implementado utilizando bibliotecas da stdlib Go (`encoding/json`, `bufio`, `net/http`), mantendo o binário estático e livre de dependências externas.
- **`SEG-01` & `SEG-02` (Validação de Borda e Anti-Traversal):** Todos os métodos RPC que recebem caminhos de workspace ou arquivos de spec passam por sanitização rigorosa (`filepath.Clean`) e validação de limites de diretório.
- **`SEG-07` (Dependências Novas):** Seção explícita atestando zero dependências externas adicionadas.
- **`DUP-01` & `DUP-02` (Anti-Duplicação e Reúso):** Reúso estrito do motor de filas existente (`internal/queue`), agregador de estado do Studio (`internal/studio`) e checagens determinísticas de portão (`internal/gate`).
- **`FLX-01` (Portão de Spec):** A resolução do Portão 1 é suportada nativamente pelo método `gate/approve`, garantindo que nenhuma esteira avance sem revisão e aprovação humana validada.
- **`FLX-02` (Isolamento de Branches e Pastas):** Fatiamento configurado para executar dentro de uma única subpasta (`sdd/features/feat-64-protocolo-acp-comunicacao-estacoes-studio/`) e uma única branch compartilhada (`feat/64-protocolo-acp-comunicacao-estacoes-studio`).
- **`TST-01` a `TST-05` (Testes Determinísticos):** Todas as fases contam com critérios de aceitação executáveis via `go test -race` e `go vet`.
