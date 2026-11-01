# feat-65-04 — Integração do Pool com Forge Studio (`internal/studio/server.go`), Atualização de Mascotes e Validação E2E

**Branch:** `feat/65-prewarming-pool-sessoes-acp-estacoes`
**Fase:** 65-04
**Tamanho:** `P`
**Depende de:** `feat-65-01`, `feat-65-02`, `feat-65-03`
**Status:** `ready`

## Contexto & Motivação

Com os blocos construtivos do `SessionPool`, `HandoffOrchestrator` e `AttachBridge` prontos nas Fases 65-01 a 65-03, a etapa final consiste em amarrar a infraestrutura do pool ao servidor do Forge Studio e consolidar a suíte de testes de ponta a ponta (E2E).

Esta subfeature realiza a integração visual e a validação completa do sistema:
1. Integrar o `SessionPool` ao bootstrap do Forge Studio em [`internal/studio/server.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/server.go), disparando o aquecimento (`Warmup`) automático no momento em que `forge-sdd studio` é inicializado.
2. Conectar os eventos de transição de estado das sessões ao `SSEHub` do Studio, fazendo com que as transições operacionais (🟢 `standby`, ⚡ `processing`, ⏸️ `approval_pending`, 🔴 `error`) reflitam instantaneamente nos cards das 5 estações e nos mascotes 3D no navegador web via Server-Sent Events.
3. Expor os métodos de attach e status via transporte HTTP/SSE (`/api/acp`) em [`internal/acp/transport_http.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/transport_http.go).
4. Criar teste de integração de ponta a ponta em [`internal/acp/e2e_pool_test.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/e2e_pool_test.go) cobrindo todo o ciclo de vida: arranque do servidor -> pre-warming das 5 sessões -> despacho de tarefa para a Estação 02 -> bloqueio em FLX-01 -> aprovação de portão -> handoff para Estação 03 -> attach via terminal.
5. Atualizar documentação e telemetria de progresso em [`sdd/memory/progress.md`](file:///Users/nathanramorim/git/forge-sdd/sdd/memory/progress.md).

## Modelo do módulo

- **Modelo:** `enxuto` dentro de [`internal/studio`](file:///Users/nathanramorim/git/forge-sdd/internal/studio) e [`internal/acp`](file:///Users/nathanramorim/git/forge-sdd/internal/acp).
- **Justificativa:** Atua exclusivamente na fiação de eventos entre o `SessionPool` e o servidor HTTP/SSE já existente do Studio, sem criar novas abstrações ou frameworks pesados (`ARQ-03`, `ARQ-04`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Servidor Web do Forge Studio | [`internal/studio/server.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/server.go) | `Server`, `NewServer`, rotas HTTP | `estender` | Injetar o `SessionPool` no ciclo de vida do servidor do Studio e disparar o `Warmup` durante o arranque. |
| Hub de Server-Sent Events | [`internal/studio/sse.go`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/sse.go) | `SSEHub`, broadcast de eventos | `reusar` | Conectar os eventos de transição de estado emitidos pelo pool diretamente ao stream do `SSEHub`. |
| Transporte HTTP/SSE do ACP | [`internal/acp/transport_http.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/transport_http.go) | `HTTPTransport`, `/api/acp` | `estender` | Roteamento das chamadas `pool/*` e `session/attach` via HTTP/SSE. |
| Testes E2E do Pool | `internal/acp` | Inexistente | `criar` | Criar `internal/acp/e2e_pool_test.go` para validação determinística do ciclo completo. |

## Contrato de Entrada e Saída

### Contrato de Entrada
- Inicialização do servidor Forge Studio (`forge-sdd studio`).
- Conexão SSE de cliente web em `/api/events`.
- Chamadas JSON-RPC 2.0 via POST em `/api/acp`.

### Contrato de Saída
- Eventos SSE emitidos para a interface web com payload:
  ```json
  {
    "type": "station_state",
    "station": "pm-architect",
    "state": "approval_pending",
    "lastSeenAt": "2026-11-01T02:10:00Z"
  }
  ```
- Respostas síncronas HTTP JSON-RPC 2.0 conformes à especificação da Feature 64.

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-governanca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-03`, `ARQ-04`, `FLX-02`, `FLX-03`, `GOV-04`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-04`, `TST-05`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Pré-aquecimento do pool durante o boot do Forge Studio
  Dado que o servidor Forge Studio é inicializado no workspace
  Quando o método NewServer concluir o arranque
  Então o SessionPool deve ter pré-aquecido as 5 sessões em standby
  E a rota HTTP /api/acp deve estar disponível para requisições JSON-RPC

Cenário: Atualização em tempo real de estados via Server-Sent Events (SSE)
  Dado que um cliente web está conectado ao stream SSE do Forge Studio
  Quando uma sessão transicionar de "standby" para "processing"
  Então uma mensagem SSE com evento "station_state" deve ser recebida pelo cliente web
  E o card da respectiva estação e seu mascote 3D devem refletir o novo estado

Cenário: Execução ponta a ponta (E2E) com aprovação de portão e handoff
  Dado um fluxo completo simulado no ambiente de teste
  Quando a Estação 02 concluir a Spec e entrar em "approval_pending"
  E o Portão FLX-01 for aprovado via chamada RPC
  Então o handoff para a Estação 03 deve ocorrer de forma autônoma
  E o cliente anexado via terminal deve receber os eventos de streaming com sucesso
```

## Plano de Implementação Passo a Passo

1. **Integração no Servidor do Studio (`internal/studio/server.go`):**
   - No construtor `NewServer`, instanciar e executar `pool.Warmup(ctx, dir)`.
   - Registrar listener de transição de estado conectando o canal de notificações do pool ao `hub.Broadcast`.
2. **Exposição de Handlers em `internal/acp/transport_http.go`:**
   - Garantir que o endpoint `/api/acp` despache chamadas para os métodos `pool/warmup`, `pool/status`, `gate/approve` e `session/attach`.
3. **Criação da Suíte de Testes E2E em `internal/acp/e2e_pool_test.go`:**
   - Criar teste `TestE2ESessionPoolFlow` cobrindo o ciclo ponta a ponta com um workspace temporário, simulando boot do Studio, handoffs entre estações, validação de portão e attach.
4. **Atualização da Telemetria e Progresso:**
   - Atualizar `sdd/memory/progress.md` garantindo tamanho `<= 1024` bytes (`Regra 14`).

## Comandos de Verificação Determinísticos

```bash
go test -v -race ./internal/acp/... ./internal/studio/... ./cmd/forge-sdd/... && go vet ./...
```
