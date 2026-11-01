# feat-66-02 — Extensão do Modelo de Telemetria e Reconciliador de Atividades Órfãs em Go (`internal/queue/activity.go`)

**Branch:** `feat/66-telemetria-detalhada-inspecao-e-reconciliacao`
**Fase:** 66-02
**Tamanho:** `M`
**Depende de:** `feat-66-01`
**Status:** `ready`

## Contexto & Motivação

Atualmente, `ActivityItem` possui apenas campos básicos (`timestamp`, `station`, `tool`, `action`, `details`, `status`). Além disso, o método `RecordActivity` opera exclusivamente em append-only sem fechar itens anteriores, resultando em status `running` órfãos que persistem visualmente no Studio mesmo com os agentes ociosos em `standby`.

Esta subfeature implementa o motor de reconciliação e enriquece a estrutura de dados:
1. Enriquecer `ActivityItem` com campos estruturados: `ID`, `Command`, `Args`, `FilePath`, `Output`, `Error`, `DurationMs`, `CompletedAt`, `MascotState`.
2. Truncamento seguro de saídas para prevenir estouro de memória e DoS (`SEG-01`, `SEG-05`), limitando `Output` e `Error` a 8 KB.
3. Função `ReconcileStationActivities(dir, station string) error`: fecha automaticamente atividades anteriores em `running` de uma estação quando uma nova atividade é registrada ou a estação transiciona de estado.
4. Função `SanitizeOrphanActivities(dir string) error`: varre `activity.json` e encerra todas as atividades antigas presas em `running`.
5. Testes unitários com `-race` comprovando concorrência thread-safe com mutex.

## Modelo do módulo

- **Modelo:** `enxuto` em `internal/queue`.
- **Justificativa:** Adiciona lógica pura em Go stdlib com proteção de concorrência (`sync.Mutex`), sem novas dependências externas (`ARQ-03`, `ARQ-04`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Modelo de Atividades | `internal/queue/activity.go` | `ActivityItem`, `RecordActivity`, `LoadActivities` | `estender` | Adicionar novos campos com tags `omitempty` para 100% de retrocompatibilidade. |
| Escrita Atômica | `internal/queue/atomic.go` | `WriteJSONAtomic` | `reusar` | Garantir que operações de reconciliação usem gravação atômica em disco. |
| Reconciliador de Atividades | `internal/queue` | Inexistente | `criar` | Implementar `ReconcileStationActivities` e `SanitizeOrphanActivities`. |

## Dependências novas

- **Nenhuma dependência externa nova.** (`SEG-07`, `ARQ-04`).

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-governanca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-03`, `ARQ-04`, `SEG-01`, `SEG-02`, `SEG-05`, `DUP-02`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-04`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Registro de nova atividade reconcilia tarefas pendentes em running
  Dado que a estação "builder" possui uma atividade com status "running"
  Quando uma nova atividade da estação "builder" for registrada via RecordActivity
  Então a atividade anterior deve ter seu status atualizado para "success"
  E a nova atividade deve ser inserida no topo com status "running"

Cenário: Sanitização de atividades órfãs no filesystem
  Dado um arquivo activity.json contendo 5 atividades antigas em "running"
  Quando a função SanitizeOrphanActivities for executada
  Então todas as 5 atividades devem ter seu status atualizado para "success"
  E nenhuma atividade deve permanecer com status "running"

Cenário: Truncamento seguro de payloads excessivos de output
  Dado um comando cujo output gerou 50 KB de logs
  Quando a atividade for gravada
  Então o campo Output deve ser truncado para no máximo 8192 bytes
  E um sufixo indicando o truncamento deve ser adicionado
```

## Plano de Implementação

1. Modificar `internal/queue/activity.go` para adicionar os campos enriquecidos a `ActivityItem`.
2. Implementar `ReconcileStationActivities(dir, station string) error`.
3. Implementar `SanitizeOrphanActivities(dir string) error`.
4. Criar `internal/queue/activity_test.go` cobrindo concorrência (`-race`), reconciliação determinística e truncamento seguro.

## Comandos de Verificação Determinísticos

```bash
go test -v -race ./internal/queue/... && go vet ./...
```
