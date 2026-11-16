# fix-84-03 — Fila continua `queued` após aprovação: IDs do job não casam com os da fila

**Branch:** `fix/84-03-ids-da-fila-nao-casam-com-o-job` (a partir de `feat/88-modo-pipeline-avancado`; PR contra ela)
**Origem:** Iniciativa 84, regressão do `fix-84-02` (visto na iniciativa 88, release 4.9.0-beta)
**Tamanho:** `S`
**Status:** `ready`

## Sintoma
O Reviewer aprova o job (`aprovado: PR sem inconformidades`, notificação enviada), mas `forge queue list` e `forge status` seguem com todas as tasks `queued`. Ocorreu na iniciativa 88 (8 tasks) e na 84 (6 tasks, mesmo depois da 4.7.7). O auto-archive nunca dispara.

## Feature Correlacionada
- **Arquivo de origem:** `fix-84-02-fila-defasada-apos-aprovacao-por-job.md` (`internal/queue/taskstatus.go`, `internal/pipeline/review.go`).
- **Causa-raiz:** `Dispatch` grava `job.TaskIDs` como `88-01` (`listTasks` tira o `feat-` do nome, `internal/pipeline/dispatch.go:295`), enquanto o `task_id` da fila é `feat-88-01` ou `feat-84-01-<nome>`. `matchesTask` aceita só ID exato ou prefixo `<id>-`, então nada casa. `SetTaskStatuses` devolve `0` sem erro (tasks desconhecidas não são erro) e a aprovação não chega à fila. `BlockTasks` (feat-88-06) usa a mesma regra e herda o defeito.
- **Evidência:** job `…-88-modo-pipeline-avancado-review` com `task_ids: ["88-01"…"88-08"]`, `initiative` preenchida, estado `done`; `queue.json` com 8 `queued`. Os testes atuais usam IDs `01`/`02`, que casam por prefixo e escondem o defeito.

## Modelo do módulo
**Enxuto** (ARQ-11): ajuste de uma função em `internal/queue`; sem módulo novo.

## Decisões do projeto aplicadas
ADR-0012 (a aprovação do job atualiza a fila; D-3 trata tasks desconhecidas como "nada a fazer", o que mascarou o defeito). Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Comparação de ID de task | `internal/queue/taskstatus.go` (`matchesTask`) | Exato ou prefixo `<id>-` | **estender**: ignorar `feat-` no ID da fila | Um único ponto, usado por `SetTaskStatuses` e `BlockTasks` |
| Normalização de ID | `internal/queue/queue.go` (`normalizeID`, usado na resolução do DAG) | Já normaliza IDs da fila | **reusar** se cobrir o prefixo `feat-`; senão estender | DUP-01: um conceito, uma regra de ID |
| Aviso de efeito nulo | `ApplyVerdict` (`internal/pipeline/review.go`) | Ignora o `0` de `SetTaskStatuses` | **estender**: avisar quando há `TaskIDs` e nada mudou nem estava `approved` | Falha deixa de ser silenciosa |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `COD-01`, `COD-03`, `DUP-01`, `TST-01`.

## Escopo
1. `matchesTask` casa `88-01` com `feat-88-01` e com `feat-88-01-<nome>`; nunca casa `88-1` com `88-10`.
2. `ApplyVerdict` e `BlockTasks` registram aviso (inbox ou log) quando o job traz `TaskIDs` e nenhuma task da fila casou.
3. Testes de regressão com os IDs reais (job `88-01` × fila `feat-88-01-…`), idempotência e não regressão de `approved`.
4. Reparar o `queue.json` das iniciativas 84 e 88 com `SetTaskStatuses` (sem edição manual do JSON).

## Fora do escopo
Mudar o formato de `task_id` já gravado ou o contrato do job.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/queue -run 'TestMatchesTask|TestSetTaskStatuses|TestBlock'
go test -v -race ./internal/pipeline -run TestApplyVerdict
go vet ./... && go test ./...
```
- Job com `task_ids: ["88-01"]` aprova a task `feat-88-01-nome` na fila.
- `88-1` não casa com `feat-88-10-…`.
- Job com `TaskIDs` sem nenhuma correspondência gera aviso, não erro.
