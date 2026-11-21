# feat-106-26b — Skill do Builder por feature, perguntas, templates e goldens

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 2
**Depende de:** feat-106-15, feat-106-24c, feat-106-26a

## Problema
Task `26b` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`.agents/skills/sdd-build/**`, `.claude/commands/implementar.md`, `split-features.md`, templates e goldens.
**Colisões e ordem:** Sequenciar depois da onda 1 da 105 e da onda 2 da 107.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `.agents/skills/sdd-build/**` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`DUP-10`, `TST-02`, `COD-05`, `FLX-01`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/scaffold/...
# negativo: TestImplementarSemGhPrCreateFill existe de fato e passa
go test ./internal/scaffold/... -list 'TestImplementarSemGhPrCreateFill' | grep -F 'TestImplementarSemGhPrCreateFill'
go test ./internal/scaffold/... -run 'TestImplementarSemGhPrCreateFill'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
- Nomes novos (não constam no `criteria.md`; o Builder os cria): `TestImplementarSemGhPrCreateFill`.
