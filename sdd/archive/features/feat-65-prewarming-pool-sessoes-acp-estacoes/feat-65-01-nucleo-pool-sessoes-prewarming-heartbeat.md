# feat-65-01 — Núcleo do SessionPool (`internal/acp/pool.go`), Pre-warming no Boot e Monitor de Heartbeat (`internal/acp/heartbeat.go`)

**Branch:** `feat/65-prewarming-pool-sessoes-acp-estacoes`
**Fase:** 65-01
**Tamanho:** `M`
**Depende de:** `feat-64-01`, `feat-64-02`
**Status:** `ready`

## Contexto & Motivação

Atualmente no `forge-sdd`, a criação de sessões ACP ocorre sob demanda com `session/new`, exigindo que ferramentas externas ou o usuário configurem e inicializem sessões individualmente para cada uma das 5 estações canônicas (`lider`, `pm-architect`, `planner`, `builder`, `governance`). Isso acarreta "cold starts" recorrentes e impede uma orquestração contínua de ponta a ponta sem atrito.

Esta subfeature estabelece o núcleo do `SessionPool` pré-aquecido e o subsistema de monitoramento contínuo:
1. Implementar o `SessionPool` em [`internal/acp/pool.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/pool.go), instanciando e mantendo em memória as 5 sessões persistentes no boot da aplicação:
   - `sess-01-lider` (estação `lider`)
   - `sess-02-pm` (estação `pm-architect`)
   - `sess-03-planner` (estação `planner`)
   - `sess-04-builder` (estação `builder`)
   - `sess-05-governance` (estação `governance`)
2. Implementar a máquina de estados operacionais das sessões: `standby`, `processing`, `approval_pending`, `error` e `offline`.
3. Implementar o `HeartbeatMonitor` em [`internal/acp/heartbeat.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/heartbeat.go), executando ciclos periódicos não-bloqueantes (`tick`) para atualizar timestamps `lastSeenAt`, detectar stalls e garantir a saúde do pool (`ARQ-03`).
4. Expor os métodos RPC `pool/warmup` e `pool/status` no [`RPCDispatcher`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/dispatcher.go).
5. Criar testes unitários com concorrência segura (`go test -race`) em [`internal/acp/pool_test.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/pool_test.go).

## Modelo do módulo

- **Modelo:** `enxuto` dentro de [`internal/acp`](file:///Users/nathanramorim/git/forge-sdd/internal/acp).
- **Justificativa:** O módulo é um orquestrador em memória gerenciando as 5 sessões ACP com estruturas puras Go stdlib (`sync.RWMutex`, canais `chan`, `time.Ticker`). Não cria daemons paralelos no S.O., não usa Docker e reutiliza as estruturas de isolamento de workspace (`ARQ-03`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Protocolo ACP e Tipos RPC | [`internal/acp/types.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/types.go) | `Request`, `Response`, `Session`, `Notification` | `estender` | Adicionar tipos `SessionPool`, `PooledSession`, `SessionState` e payloads de `pool/warmup` e `pool/status`. |
| Despachante de RPCs | [`internal/acp/dispatcher.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/dispatcher.go) | `RPCDispatcher` thread-safe | `estender` | Registrar handlers `pool/warmup` e `pool/status`. |
| Estações Canônicas | [`internal/queue/stations.go`](file:///Users/nathanramorim/git/forge-sdd/internal/queue/stations.go) | `CanonicalStations()`, `IsValidStation()` | `reusar` | O pool pré-aquece estritamente as 5 estações canônicas retornadas por `queue.CanonicalStations()` (`DUP-01`). |
| Validação de Workspace e Isolamento | [`internal/acp/session.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/session.go) | `ValidateWorkspace` | `reusar` | O método `Warmup` valida caminhos e confina ao repositório via `ValidateWorkspace` (`SEG-01`, `SEG-02`). |
| Componentes do Pool e Heartbeat | `internal/acp` | Inexistente | `criar` | Criar `internal/acp/pool.go`, `internal/acp/heartbeat.go` e `internal/acp/pool_test.go`. |

## Contrato de Entrada e Saída

### Contrato de Entrada (JSON-RPC 2.0)
- **`pool/warmup`:**
  ```json
  {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "pool/warmup",
    "params": {
      "workspaceDir": "/caminho/para/workspace"
    }
  }
  ```
- **`pool/status`:**
  ```json
  {
    "jsonrpc": "2.0",
    "id": 2,
    "method": "pool/status"
  }
  ```

### Contrato de Saída (JSON-RPC 2.0)
- **`pool/warmup`:** Retorna `status: "warmed"`, `sessionsCount: 5`, e o array das 5 sessões canônicas em estado `standby`.
- **`pool/status`:** Retorna indicador booleano `healthy: true` e a lista de sessões com `sessionId`, `station`, `state`, `lastSeenAt`, `activeItemId` e `attachedClients`.

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-07`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-03`, `TST-04`, `TST-05`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Inicialização e pré-aquecimento automático das 5 sessões
  Dado que o diretório de workspace fornecido é válido e confinado
  Quando o método Warmup for invocado no SessionPool
  Então exatamente 5 sessões persistentes devem ser instanciadas no pool
  E as sessões devem possuir os IDs: sess-01-lider, sess-02-pm, sess-03-planner, sess-04-builder, sess-05-governance
  E todas as 5 sessões devem estar inicialmente com o estado "standby"
  E deve ser emitida uma notificação "pool/warmed"

Cenário: Monitoramento de heartbeat e atualização de telemetria
  Dado que o SessionPool está ativo com as 5 sessões pré-aquecidas
  Quando o HeartbeatMonitor executar seu ciclo periódico (tick)
  Então o campo "lastSeenAt" de cada sessão deve ser atualizado
  E se uma sessão permanecer em "processing" sem telemetria além do limite de timeout
  Então uma notificação "session/stalled" deve ser emitida com os metadados da sessão

Cenário: Rejeição de workspace inválido na inicialização do pool
  Dado que a requisição "pool/warmup" fornece workspaceDir com tentativa de path traversal ("../etc")
  Quando o SessionPool tentar validar o diretório
  Então a operação deve falhar com erro CodeInvalidParams (-32602)
  E nenhuma sessão fora do limite permitido deve ser alocada
```

## Plano de Implementação Passo a Passo

1. **Definição de Tipos em `internal/acp/types.go`:**
   - Adicionar o enum `SessionState` (`standby`, `processing`, `approval_pending`, `error`, `offline`).
   - Adicionar as structs `PooledSession`, `PoolWarmupParams`, `PoolWarmupResult`, `PoolStatusResult` e `SessionSummary`.
2. **Criação de `internal/acp/pool.go`:**
   - Implementar `SessionPool` com proteção thread-safe via `sync.RWMutex`.
   - Implementar método `NewSessionPool(sm *SessionManager, dispatcher *RPCDispatcher) *SessionPool`.
   - Implementar `Warmup(ctx context.Context, wsDir string) (*PoolWarmupResult, error)` validando com `sm.ValidateWorkspace`.
   - Implementar `GetSessionByStation(station string) (*PooledSession, bool)`.
   - Implementar `ListSessions() []*PooledSession` retornando cópias defensivas (`defensive copy`).
   - Implementar `TransitionState(station string, newState SessionState, reason string) error`.
3. **Criação de `internal/acp/heartbeat.go`:**
   - Implementar `HeartbeatMonitor` com `time.Ticker`, `context.Context` e método `tick()`.
   - Detectar sessões com tempo de execução excedido (`stalled`) e emitir notificação `session/stalled`.
4. **Registro no `RPCDispatcher`:**
   - Conectar os métodos RPC `pool/warmup` e `pool/status` no `dispatcher.go`.
5. **Testes Unitários:**
   - Implementar `internal/acp/pool_test.go` cobrindo concorrência, transições válidas/inválidas e tick do heartbeat com `go test -race`.

## Comandos de Verificação Determinísticos

```bash
go test -v -race -run "TestSessionPool" ./internal/acp/... && go vet ./...
```
