# feat-106-07a — Estado por feature, claim pela branch e mutadores da fila

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 0
**Depende de:** feat-106-11

## Problema
Task `07a` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`internal/queue/queue.go`, `taskstatus.go` (`SetTaskStatuses`, `BlockTasks`, `Review`, `Load`), `internal/pipeline` (claim).
**Colisões e ordem:** Ordem: depois da 11 (`queue.go`, formato do parecer); `queue.go` segue a 07a antes da 16 e da 24.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `internal/queue/queue.go` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`DUP-04`, `DUP-01`, `ARQ-02`, `TST-03`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/... -run 'TestClaimPelaExistenciaDaBranch|TestEstadoDeFeaturesParalelasSemConflito'
# negativo: TestDuasFeaturesNaoEscrevemNoMesmoArquivo existe de fato e passa
go test ./internal/pipeline/... -list 'TestDuasFeaturesNaoEscrevemNoMesmoArquivo' | grep -F 'TestDuasFeaturesNaoEscrevemNoMesmoArquivo'
go test ./internal/pipeline/... -run 'TestDuasFeaturesNaoEscrevemNoMesmoArquivo'
# negativo: TestSegundoClaimDaMesmaFeatureFalha existe de fato e passa
go test ./internal/pipeline/... -list 'TestSegundoClaimDaMesmaFeatureFalha' | grep -F 'TestSegundoClaimDaMesmaFeatureFalha'
go test ./internal/pipeline/... -run 'TestSegundoClaimDaMesmaFeatureFalha'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
