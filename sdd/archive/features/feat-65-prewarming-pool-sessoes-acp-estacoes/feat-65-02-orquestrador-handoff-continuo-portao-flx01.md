# feat-65-02 — Orquestrador de Handoff Contínuo (`internal/acp/orchestrator.go`), Despacho Automático de Filas e Integração com Portão FLX-01

**Branch:** `feat/65-prewarming-pool-sessoes-acp-estacoes`
**Fase:** 65-02
**Tamanho:** `M`
**Depende de:** `feat-65-01`
**Status:** `ready`

## Contexto & Motivação

No modelo anterior do `forge-sdd`, a transição entre estações exigia intervenção manual do desenvolvedor para trocar de terminal ou rodar comandos isolados de handoff. Com a introdução do `SessionPool` pré-aquecido na Fase 65-01, as sessões permanecem em prontidão (`standby`), permitindo uma esteira contínua e autônoma. No entanto, essa autonomia deve respeitar com rigor absoluto os portões de governança humana da Constituição (`FLX-01`).

Esta subfeature implementa o `HandoffOrchestrator` em [`internal/acp/orchestrator.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/orchestrator.go):
1. Escutar eventos e itens nas filas atômicas de [`internal/queue/stations.go`](file:///Users/nathanramorim/git/forge-sdd/internal/queue/stations.go) e despachar tarefas para a sessão correspondente em `standby`.
2. Orquestrar a transição contínua entre Concepção (Estação 02) e Planejador (Estação 03):
   - Quando `sess-02-pm` conclui Discovery e Spec, seu estado transiciona obrigatoriamente para `approval_pending`.
   - O orquestrador emite a notificação ACP `gate/approval_required` e congela qualquer despacho subsequente até aprovação humana explícita (`FLX-01`).
   - Ao receber o comando ou RPC `gate/approve`, valida deterministicamente os critérios `P-02` e `P-03` via [`internal/gate/gate.go`](file:///Users/nathanramorim/git/forge-sdd/internal/gate/gate.go).
   - Com o portão aprovado, a sessão `sess-02-pm` retorna para `standby` e a sessão `sess-03-planner` é ativada automaticamente para `processing`.
3. Emitir notificações estruturadas `session/notification` do tipo `station_state` informando o Studio e IDEs em tempo real.
4. Prevenir tentativas de bypass de portão com erro canônico `-32001` (`CodeServerError`).
5. Criar suíte de testes de integração com detecção de concorrência (`go test -race`) em [`internal/acp/orchestrator_test.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/orchestrator_test.go).

## Modelo do módulo

- **Modelo:** `enxuto` dentro de [`internal/acp`](file:///Users/nathanramorim/git/forge-sdd/internal/acp).
- **Justificativa:** Conecta o despachante de eventos e o `SessionPool` às rotinas já existentes de `internal/queue` e `internal/gate`. Não duplica regras de validação nem cria persistência paralela (`ARQ-01`, `ARQ-02`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Motor de Filas e Handoffs | [`internal/queue/stations.go`](file:///Users/nathanramorim/git/forge-sdd/internal/queue/stations.go) | `Handoff`, `EnqueueStationItem`, `LoadStationQueue` | `reusar` | O orquestrador utiliza as funções canônicas de `internal/queue` para movimentação atômica e idempotente (`DUP-01`). |
| Validador de Portão Camada A | [`internal/gate/gate.go`](file:///Users/nathanramorim/git/forge-sdd/internal/gate/gate.go) | `ValidateFLX01`, `checkP02SpecExists`, `checkP03ReuseTable` | `reusar` | `HandoffOrchestrator` delega a validação de portão diretamente a `internal/gate` para garantir conformidade determinística com `P-02` e `P-03`. |
| SessionPool Pré-aquecido | [`internal/acp/pool.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/pool.go) | `SessionPool`, `TransitionState`, `GetSessionByStation` | `estender` | O orquestrador é acoplado ao `SessionPool` para coordenar transições de estado entre sessões durante o handoff. |
| Orquestrador de Handoff | `internal/acp` | Inexistente | `criar` | Criar `internal/acp/orchestrator.go` e `internal/acp/orchestrator_test.go`. |

## Contrato de Entrada e Saída

### Contrato de Entrada
- **Evento de Conclusão da Estação 02:**
  Sessão `sess-02-pm` sinaliza finalização de artefato (`specPath`).
- **Comando de Aprovação de Portão (`gate/approve`):**
  ```json
  {
    "jsonrpc": "2.0",
    "id": 10,
    "method": "gate/approve",
    "params": {
      "station": "pm-architect",
      "itemId": "feat-65-prewarming-pool-sessoes-acp-estacoes",
      "specPath": "sdd/features/feat-65-prewarming-pool-sessoes-acp-estacoes/plan-65-prewarming-pool-sessoes-acp-estacoes.md"
    }
  }
  ```

### Contrato de Saída
- **Notificação `gate/approval_required`:**
  Emitida quando uma estação requer chancela humana, contendo `station`, `itemId`, `specPath` e regras aplicadas.
- **Notificação `session/notification`:**
  Tipo `station_state` informando `station`, `state` anterior e novo `state`.
- **Resultado do RPC `gate/approve`:**
  ```json
  {
    "jsonrpc": "2.0",
    "id": 10,
    "result": {
      "approved": true,
      "gate": "FLX-01",
      "checks": [
        {"id": "P-02", "passed": true},
        {"id": "P-03", "passed": true}
      ],
      "nextStation": "planner"
    }
  }
  ```

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-fluxo-e-portao.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-02`, `SEG-01`, `DUP-01`, `DUP-02`, `FLX-01`, `FLX-04`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-02`, `TST-03`, `TST-04`, `TST-05`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Bloqueio estrito no Portão FLX-01 após conclusão de Spec pelo PM
  Dado que a sessão "sess-02-pm" concluiu o Discovery e Spec
  Quando o HandoffOrchestrator processar o evento de conclusão
  Então o estado de "sess-02-pm" deve transicionar para "approval_pending"
  E uma notificação "gate/approval_required" deve ser emitida
  E a sessão "sess-03-planner" NÃO deve ser ativada enquanto o portão não for aprovado

Cenário: Desbloqueio e handoff contínuo após aprovação de Portão FLX-01
  Dado que a sessão "sess-02-pm" está em "approval_pending" aguardando aprovação
  Quando a chamada RPC "gate/approve" for recebida com os artefatos em conformidade
  Então a validação P-02 e P-03 de internal/gate deve passar com sucesso
  E o status do item na fila deve ser atualizado para "approved"
  E a sessão "sess-02-pm" deve retornar para "standby"
  E a sessão "sess-03-planner" deve ser ativada automaticamente transicionando para "processing"

Cenário: Rejeição de avanço sem aprovação válida (Tentativa de bypass)
  Dado que a sessão "sess-02-pm" está em "approval_pending"
  Quando uma chamada de despacho tentar avançar para a Estação 03 sem aprovação de portão
  Então o HandoffOrchestrator deve rejeitar a operação com erro canônico CodeServerError (-32001)
  E o estado das sessões deve permanecer inalterado
```

## Plano de Implementação Passo a Passo

1. **Definição da Interface e Estrutura em `internal/acp/orchestrator.go`:**
   - Criar struct `HandoffOrchestrator` contendo referências para `*SessionPool`, `dir string` e `*RPCDispatcher`.
   - Implementar método `NewHandoffOrchestrator(pool *SessionPool, dir string, dispatcher *RPCDispatcher) *HandoffOrchestrator`.
2. **Implementação de Gatilhos de Handoff:**
   - Implementar método `OnSpecCreated(station string, itemID string, specPath string) error`. Transiciona para `StateApprovalPending` e emite `gate/approval_required`.
   - Implementar método `OnGateApproved(station string, itemID string, specPath string) (*GateApproveResult, error)`. Invoca `gate.ValidateFLX01`, atualiza item na fila para `approved`, devolve estação de origem para `standby` e aciona próxima estação.
3. **Mapeamento de Próxima Estação Canônica:**
   - Implementar `NextStation(currentStation string) (string, bool)` respeitando o fluxo: `lider` -> `pm-architect` -> `planner` -> `builder` -> `governance`.
4. **Integração com `RPCDispatcher`:**
   - Registrar handler do método `gate/approve` apontando para o `HandoffOrchestrator`.
5. **Testes Unitários e de Concorrência:**
   - Criar `internal/acp/orchestrator_test.go` cobrindo o fluxo completo de handoff, validação de bloqueio FLX-01 e rejeição de bypass concorrente com `go test -race`.

## Comandos de Verificação Determinísticos

```bash
go test -v -race -run "TestHandoffOrchestrator" ./internal/acp/... ./internal/queue/... && go vet ./...
```
