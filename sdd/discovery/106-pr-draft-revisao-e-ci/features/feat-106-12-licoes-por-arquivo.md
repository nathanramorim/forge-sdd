# feat-106-12 — scripts/licoes.py e lições por arquivo

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 1
**Depende de:** nenhuma

## Problema
Task `12` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`scripts/licoes.py`, `scripts/tests/`, `sdd/memory/licoes/`, `cmd/forge-sdd/lessons.go`.
**Colisões e ordem:** sem colisão com outra feature ativa além da ordem de `Depende de`.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `scripts/licoes.py` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`DUP-01`, `ARQ-03`, `GOV-03`, `TST-03`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
uv run --no-project python -m unittest discover -s scripts/tests -p 'test_*.py'
uv run scripts/licoes.py --rule COD-02 | grep -F 'COD-02'
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/licoes.py --rule ZZZ-99; then exit 1; fi
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
- Nomes novos (não constam no `criteria.md`; o Builder os cria): `ZZZ-99` (regra inexistente recusada).
