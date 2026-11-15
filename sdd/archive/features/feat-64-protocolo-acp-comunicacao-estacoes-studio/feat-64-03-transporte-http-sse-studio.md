# feat-64-03 — Transporte ACP HTTP/SSE no Forge Studio (`internal/studio`) e Bridge com SSEHub

**Branch:** `feat/64-protocolo-acp-comunicacao-estacoes-studio`
**Fase:** 64-03
**Tamanho:** `M`
**Depende de:** `feat-64-01`, `feat-64-02`
**Status:** `ready`

## Objetivo

Estender a camada de rede do Forge Studio (`internal/studio`) para suportar comunicação ACP via transporte HTTP/SSE, viabilizando interações em tempo real tanto para interfaces web quanto para agentes remotos:
1. Estender [`internal/studio/server.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/server.go) adicionando os endpoints:
   - `POST /api/acp/rpc`: recepção e execução de comandos JSON-RPC 2.0 delegados ao `RPCDispatcher`.
   - `GET /api/acp/events`: stream Server-Sent Events (SSE) dedicado para eventos e notificações ACP.
2. Implementar [`internal/acp/transport_http.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/transport_http.go) para adaptação segura dos envelopes HTTP/JSON com limitação de tamanho de payload (máximo 2 MB) e sanitização estrita de cabeçalhos (`SEG-01`, `SEG-05`).
3. Conectar as notificações do `RPCDispatcher` ao [`SSEHub`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/sse.go) existente em `internal/studio/sse.go`, unificando a distribuição de eventos das 5 estações sem abrir portas extras (`ARQ-03`).
4. Preservar rigorosamente o funcionamento das rotas legadas do Forge Studio (`/api/state`, `/api/spec`, etc.), garantindo retrocompatibilidade total (`DUP-02`).
5. Criar testes de estresse e concorrência em [`internal/studio/acp_transport_test.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/acp_transport_test.go) com múltiplos clientes HTTP simultâneos e canais SSE abertos (`TST-01`, `TST-03`).

## Modelo do módulo
- **Modelo:** `enxuto` dentro de [`internal/studio`](file:///Users/nathanramorim/git/forge-sdd/internal/studio).
- **Justificativa:** Reutiliza a instância existente do servidor HTTP e o hub SSE (`internal/studio/sse.go`), evitando a alocação de novas portas de rede ou processos concorrentes (`ARQ-03`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Servidor HTTP Studio | [`internal/studio/server.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/server.go) | `Server` com `ServeMux`, `handleState`, `handleEvents` | `estender` | Mapear rotas `/api/acp/rpc` e `/api/acp/events` mantendo os middlewares existentes. |
| Distribuição SSE | [`internal/studio/sse.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/sse.go) | `SSEHub` com broadcast concorrente e registro de clientes | `reusar` | Aproveitar a infraestrutura de SSE já homologada e testada contra concorrência e desconexões abruptas. |
| Agregador de Estado | [`internal/studio/state.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/state.go) | `BuildStudioSnapshot` e estruturas de DTO | `reusar` | Transmitir eventos formatados em consonância com o modelo visual do Studio. |

## Dependências novas (`SEG-07`)

- **Nenhuma dependência externa.**
- Implementação estrita com 100% Go stdlib (`net/http`, `sync`, `time`, `encoding/json`, `fmt`, `context`).

## Rules e IDs da Constituição Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-11`, `SEG-01`, `SEG-04`, `SEG-05`, `DUP-02`, `COD-01`, `COD-02`, `TST-01`, `TST-03`, `TST-04`, `TST-05`.
- **Skills Reutilizadas:** —

## Arquivos Tocados

- `internal/studio/server.go` (estender)
- `internal/acp/transport_http.go` (criar)
- `internal/studio/acp_transport_test.go` (criar)

## Critérios de Aceitação Executáveis

1. **CA-01 (Despacho JSON-RPC via HTTP POST `/api/acp/rpc`):**
   ```bash
   go test -v -race -run "TestStudioACPRPC" ./internal/studio/... && go vet ./...
   ```
2. **CA-02 (Conexão e Broadcast via SSE `/api/acp/events`):**
   ```bash
   go test -v -race -run "TestStudioACPEventsSSE" ./internal/studio/... && go vet ./...
   ```
3. **CA-03 (Concorrência Multicliente e Ausência de Regressão em Rotas Legadas):**
   ```bash
   go test -v -race -run "TestStudioACPTransport" ./internal/studio/... && go vet ./...
   ```
