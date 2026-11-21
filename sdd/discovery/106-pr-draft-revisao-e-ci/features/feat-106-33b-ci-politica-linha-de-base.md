# feat-106-33b — ci-politica.py: alerta, CSV, linha de base e métricas normalizadas

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 3
**Depende de:** feat-106-33a

## Problema
Task `33b` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`scripts/ci-politica.py`, testes, fixtures.
**Colisões e ordem:** sem colisão com outra feature ativa além da ordem de `Depende de`.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `scripts/ci-politica.py` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`SEG-04`, `SEG-05`, `TST-03`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
uv run scripts/ci-politica.py --offline --csv scripts/tests/fixtures/runs.json
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if CI_POLITICA_TOKEN=segredo-de-teste-123 uv run scripts/ci-politica.py --offline --csv scripts/tests/fixtures/runs.json | grep -F 'segredo-de-teste-123'; then exit 1; fi
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/ci-politica.py --offline --csv scripts/tests/fixtures/runs_com_formula.json | grep -E '^[=+@-]'; then exit 1; fi
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/ci-politica.py --congelar-linha-de-base --sem-data --offline; then exit 1; fi
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
