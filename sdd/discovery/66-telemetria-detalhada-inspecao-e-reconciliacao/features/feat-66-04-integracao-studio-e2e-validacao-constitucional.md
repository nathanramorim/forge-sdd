# feat-66-04 — Integração no Studio, Testes E2E e Validação Constitucional Completa

**Branch:** `feat/66-telemetria-detalhada-inspecao-e-reconciliacao`
**Fase:** 66-04
**Tamanho:** `M`
**Depende de:** `feat-66-01`, `feat-66-02`, `feat-66-03`
**Status:** `ready`

## Contexto & Motivação

Com as fundações do novo formato unificado de discovery/fixes, o reconciliador de atividades em Go e o modal interativo no frontend prontos nas Fases 66-01 a 66-03, a fase final consiste em amarrar a telemetria ao ciclo de vida do Studio e consolidar a validação ponta a ponta (E2E).

Esta subfeature fecha o ciclo do sistema:
1. Conectar a reconciliação automática de atividades às transições de estado do `SessionPool` (`internal/acp/pool.go` e `internal/studio/server.go`): quando qualquer estação entrar em `standby`, atividades pendentes são imediatamente reconciliadas para `success`.
2. Sanitização automática no boot do Studio (`server.go` e `state.go`), garantindo que nenhum projeto inicie com atividades fantasmas em `running`.
3. Adicionar flag `--sanitize` ao subcomando `forge-sdd queue activity` (`cmd/forge-sdd/queue.go`).
4. Criar suíte de testes de integração e E2E cobrindo o fluxo completo de emissão, reconciliação, renderização e integridade sob `-race`.
5. Recompilar o binário em cache e validar o Studio ativo na porta 4321.

## Modelo do módulo

- **Modelo:** `enxuto` em `internal/studio`, `internal/acp` e `cmd/forge-sdd`.
- **Justificativa:** Amarração e testes de ponta a ponta sem frameworks externos (`ARQ-03`, `ARQ-04`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Servidor do Studio | `internal/studio/server.go` | `NewServer`, hooks de ciclo de vida | `estender` | Disparar `SanitizeOrphanActivities` no boot e acoplar reconciliação aos listeners de estado do pool. |
| Agregador de Estado | `internal/studio/state.go` | `GetSnapshot()` | `estender` | Assegurar que estações em standby não emitam atividades com status `running`. |
| CLI de Atividades | `cmd/forge-sdd/queue.go` | `queueActivityCmd` | `estender` | Adicionar flag `--sanitize` para limpeza manual via terminal. |
| Testes E2E | `internal/studio/server_test.go` | Testes do Studio | `estender` | Adicionar validação E2E da telemetria detalhada e reconciliação. |

## Dependências novas

- **Nenhuma dependência externa nova.** (`SEG-07`, `ARQ-04`).

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-governanca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-06`, `SEG-01`, `SEG-06`, `DUP-02`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-04`, `TST-05`, `FLX-01`, `FLX-02`, `FLX-03`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Inicialização do Studio higieniza atividades órfãs no boot
  Dado um arquivo activity.json com 3 atividades antigas em "running"
  Quando o servidor Forge Studio é inicializado com NewServer
  Então todas as atividades anteriores devem ser saneadas para "success"
  E o snapshot retornado pela API /api/state não deve conter atividades em "running" para estações em standby

Cenário: Transição de sessão para standby dispara reconciliação automática
  Dado que a sessão sess-04-builder está em processing com uma atividade running registrada
  Quando a sessão for transicionada para SessionStateStandby
  Então a atividade running deve ser automaticamente atualizada para "success"
  E um evento SSE deve ser emitido notificando a mudança aos clientes web

Cenário: Execução do comando CLI com flag --sanitize
  Dado um projeto com atividades pendentes no histórico
  Quando o operador executa "forge-sdd queue activity --sanitize"
  Então as atividades órfãs devem ser reconciliadas com mensagem de sucesso
```

## Plano de Implementação

1. Injetar `SanitizeOrphanActivities` na inicialização de `NewServer` em `internal/studio/server.go`.
2. Registrar hook de transição de estado no `SessionPool` para chamar `ReconcileStationActivities`.
3. Adicionar flag `--sanitize` no subcomando `queueActivityCmd` em `cmd/forge-sdd/queue.go`.
4. Criar testes automatizados de integração cobrindo os cenários descritos com `-race`.
5. Recompilar o binário `forge-sdd` e reiniciar o daemon do Studio.

## Comandos de Verificação Determinísticos

```bash
go test -v -race ./cmd/forge-sdd/... ./internal/queue/... ./internal/studio/... ./internal/acp/... && go vet ./...
```
