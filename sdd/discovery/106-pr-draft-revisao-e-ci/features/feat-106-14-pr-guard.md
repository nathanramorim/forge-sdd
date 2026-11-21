# feat-106-14 — scripts/pr_guard.py (barreira de gh/az no Claude Code)

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 1
**Depende de:** feat-106-02, feat-106-11

## Problema
Task `14` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`scripts/pr_guard.py` e testes.
**Colisões e ordem:** sem colisão com outra feature ativa além da ordem de `Depende de`.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `scripts/pr_guard.py` e testes | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`SEG-01`, `SEG-03`, `SEG-07`, `FLX-08`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
uv run scripts/pr_guard.py --check 'gh pr create --draft --fill'
uv run scripts/pr_guard.py --check 'az repos pr create --draft true --title x'
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/pr_guard.py --check 'gh pr create --fill'; then exit 1; fi
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/pr_guard.py --check 'gh pr merge 7'; then exit 1; fi
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/pr_guard.py --check 'az repos pr create --title x'; then exit 1; fi
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/pr_guard.py --check 'az repos pr update --id 7 --draft false'; then exit 1; fi
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
