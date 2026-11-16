# fix-84-02 — Painel e fila defasados após a aprovação de um job pelo Reviewer

**Branch:** `fix/84-02-fila-defasada-apos-aprovacao-por-job` (a partir de `feat/84-painel-interativo-chat-e-fila-drilldown`; PR contra ela)
**Origem:** Iniciativa 84 (`internal/pipeline/review.go`, `internal/queue/queue.go`)
**Tamanho:** `M`
**Status:** `done`

## Sintoma
Na iniciativa `55e0-credenciais-convites-midia-google-calendar`, o job de rework foi aprovado pelo Reviewer e notificado como "PR #4 pronto para aprovação final". Mesmo assim `forge status` e `forge queue list` seguem em 0/6, as 5 tasks continuam `queued`, a task 02 continua `changes_requested` (registrada à mão com `forge queue review`) e o `status.md` fica em `em_execucao`.

## Causa-raiz (verificada no código)
1. **A aprovação não toca na fila.** No ramo `conforme` de `ApplyVerdict` (`internal/pipeline/review.go`, ~l.87-107) só ocorrem a notificação, `bus.Complete` do job e `CheckAndAutoArchive`. Nenhuma chamada a `queue.Review`/atualização de `queue.json`.
2. **O vínculo já existe e não é usado:** `bus.Job` tem `Initiative` e `TaskIDs` (`internal/bus/job.go`); `ApplyVerdict` ignora ambos ao aprovar.
3. **Efeito colateral:** `AutoArchiveInitiative` (`internal/pipeline/archive.go`) só arquiva se todos os itens estiverem `approved`, então a fila defasada também impede o auto-archive.
4. **Veredito em `/tmp`:** `pr-review.md` instrui gravar o JSON em "arquivo temporário"; fica volátil e fora da pasta da iniciativa.

## Premissas (a confirmar na revisão)
- Um job aprovado cobre exatamente as tasks de `job.TaskIDs` da iniciativa `job.Initiative`.
- Task em `changes_requested` cujo rework foi aprovado passa direto a `approved`.
- Job sem `TaskIDs`/`Initiative` mantém o comportamento atual (sem erro).

## Critérios de aceitação
- [x] Ao aplicar veredito `conforme`, cada task de `job.TaskIDs` fica `approved` no `queue.json` da iniciativa; `forge status` e `forge queue list` refletem a contagem.
- [x] Reaplicar o mesmo veredito é idempotente (sem duplicar efeitos).
- [x] O Builder marca as tasks entregues como `review_pending` ao concluir o job de build/rework.
- [x] ~~O `status.md` sai de `em_execucao`~~: nenhum código Go escreve esse arquivo (é atualizado pelo Líder, pelo fechamento da skill); fica fora do escopo. A fila correta já viabiliza o auto-archive.
- [x] O veredito da Camada B é preservado em `sdd/discovery/<iniciativa>/verdicts/<job>.json`. **Decisão:** a cópia é feita pelo Go em `ApplyVerdict`, não pelo agente, para o Reviewer continuar somente leitura (`SEG-06`); a skill `pr-review.md` (template e golden) só informa isso.
- [x] Violação e escalação não alteram o status das tasks além do já existente.

## Escopo
1. `internal/pipeline/review.go`: propagar `approved` às tasks do job no ramo `conforme`, antes do auto-archive.
2. Ponto de entrega do Builder (`cmd/forge-sdd/worker.go` / `internal/pipeline/`): atualizar as tasks para `review_pending`.
3. `internal/scaffold/templates/.agents/skills/sdd-governanca/references/pr-review.md.tmpl` + golden: caminho do veredito na pasta da iniciativa.
4. Testes de regressão (aprovação propaga; idempotência; job sem `TaskIDs`; `queue.json` ausente).
5. Validação: `go build ./...`, `go vet ./...`, `go test -race ./...`.

## Fora do escopo
Mudar o contrato do veredito ou a regra de tentativas de rework.

## Escopo Aplicado
1. `internal/queue/taskstatus.go`: `SetTaskStatuses` (por iniciativa, idempotente, sem regredir `approved`, fila ausente ou iniciativa inválida não é erro).
2. `internal/pipeline/review.go`: `conforme` propaga `approved` antes do `bus.Complete`; `persistVerdict` guarda a cópia do veredito.
3. `cmd/forge-sdd/worker.go`: `pipeline pr-opened` marca as tasks do job como `review_pending`.
4. `pr-review.md.tmpl` e golden: mencionam a cópia do veredito.
5. Testes: `taskstatus_test.go`, `review_queue_test.go`, `worker_test.go`. A remoção da correção faz o teste de aprovação falhar (verificado).
6. Versão `4.7.7-beta` (`cmd/forge-sdd/main.go`, `npm/package.json`, `sdd/releases/history.md`).
7. `go build ./...`, `go vet ./...` e `go test -race ./...` passam.
