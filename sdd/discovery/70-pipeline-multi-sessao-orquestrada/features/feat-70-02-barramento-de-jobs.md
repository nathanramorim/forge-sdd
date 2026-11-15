# feat-70-02 — Barramento de Jobs (`internal/bus`, `forge-sdd bus`)

**Branch:** `feat/70-pipeline-multi-sessao-orquestrada`
**Fase:** 70-02
**Tamanho:** `G`
**Depende de:** feat-70-01
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
As três sessões trocam jobs por arquivos. O barramento define o contrato do job, os estados por diretório, a reivindicação atômica com lease, a prioridade e a recuperação de jobs presos. É o coração da concorrência e precisa de testes de corrida.

> Rule of 7: tamanho `G`. A Governança pode exigir dividir em contrato e comandos no Portão.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** concorrência entre processos (`rename`), leases, recuperação e contrato versionado; fronteira própria (`COD-01`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Escrita atômica | `internal/queue/atomic.go` (`WriteJSONAtomic`) | Temporário mais `os.Rename` | reusar | Todo job é gravado por ela |
| Filas por estação | `internal/queue/stations.go` | Um JSON por estação, idempotência SHA-256, em `sdd/.runtime/` | reusar o padrão | O barramento precisa de um arquivo por job para reivindicar com `rename`; guarda a ideia de idempotência por id |
| Fila de tasks | `internal/queue/queue.go` | Tasks de uma iniciativa | reusar | Continua separada: o job aponta para a feature, não substitui a fila de tasks |
| Caminhos | `internal/layout` (70-01) | `QueueState`, `Inbox` | reusar | Única fonte (`DUP-04`) |
| Barramento entre sessões | — | Inexistente | criar `internal/bus` | Conceito novo; entra em `sdd/spec/concepts.md` |
| Esqueleto de subcomando | `.agents/skills/novo-subcomando-cli-sdd` | Procedimento aprovado | reusar | `DUP-05` |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-10`, `SEG-01`, `SEG-02`, `SEG-05`, `COD-01`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Skill: `novo-subcomando-cli-sdd`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. **Contrato do job** com os campos da especificação (`id`, `type`, `priority`, `feature`, `title`, `spec_path`, `adr_refs`, `rules_refs`, `card`, `pr`, `attempt`, `max_rework_attempts`, `rework_feedback`, `created_at`, `updated_at`) mais `schema_version`, `initiative`, `task_ids`, `lease` e `events[]`. Tipos `feature_build`, `pr_rework`, `pr_review`; prioridades `high` e `normal`. Validação na borda: id `^[a-z0-9][a-z0-9._-]{0,80}$`, tamanho máximo de 256 KB.
2. **Estados por diretório** e nome de arquivo `<prio>-<timestamp>-<id>.json` (`0-` para `high`, `1-` para `normal`); a ordem lexicográfica é prioridade e depois FIFO.
3. **Operações:** `Enqueue` (duplicata por id é idempotente), `Claim` (`rename` de `pending` para `processing` com lease `worker` e `expires_at`), `Complete`, `Fail` (com motivo), `Recover` (lease vencido volta a `pending` com evento) e `List`. JSON corrompido em `pending` vai para `failed` com o motivo, sem travar a fila.
4. `forge-sdd bus enqueue|claim|complete|fail|recover|list|show` com saída `--json`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/bus/... ./cmd/forge-sdd/...
go test ./internal/bus -run 'TestClaim_PrioridadeHighAntesDeNormalEFIFO|TestClaim_Concorrente|TestRecover_LeaseVencido'
```
- `high` é consumido antes de `normal`; dentro da mesma prioridade, FIFO.
- Oito reivindicações concorrentes do mesmo job: exatamente uma vence (`-race`).
- Lease vencido volta a `pending`; lease válido não.
- Violações (`TST-03`): id com `..`, `/` ou espaço, papel ou estado inválido, job acima de 256 KB, tipo desconhecido, duplicata, JSON corrompido.
- Nenhum segredo nem caminho absoluto fora do projeto nos eventos (`SEG-05`).
