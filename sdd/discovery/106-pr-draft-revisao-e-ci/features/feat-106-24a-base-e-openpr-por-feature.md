# feat-106-24a — Base sempre base_branch e OpenPR pelo adaptador (uma branch e um PR draft por feature)

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 2
**Depende de:** feat-106-07a, feat-106-21, feat-106-22a

## Problema
Task `24a` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`internal/queue/base.go` (`ResolveBase`), `internal/pipeline/pr.go`, `goal.go`.
**Colisões e ordem:** `pr.go` depois da 21.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `internal/queue/base.go` (`ResolveBase`) | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`DUP-03`, `ARQ-02`, `TST-04`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/... -run 'TestUmaBranchEUmPRDraftPorFeature'
# negativo: TestFeatureSemPRDraftNaoAvanca existe de fato e passa
go test ./internal/pipeline/... -list 'TestFeatureSemPRDraftNaoAvanca' | grep -F 'TestFeatureSemPRDraftNaoAvanca'
go test ./internal/pipeline/... -run 'TestFeatureSemPRDraftNaoAvanca'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
