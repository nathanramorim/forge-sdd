# feat-94-01 — Merge PR #91 (fix-88-03 aprovado)

**Objetivo:** mergear o PR #91 (`fix/88-03-ask-user-e-seletor-obrigatorios → feat/88-modo-pipeline-avancado`), cujo veredito Reviewer já é `conforme`.

**Entradas:**
- PR #91 aberto em `https://github.com/nathanramorim/forge-sdd/pull/91`
- Veredito em `sdd/discovery/88-modo-pipeline-avancado/verdicts/job-20261119-000439-88-modo-pipeline-avancado-fix-88-03-review.json` (`"veredito": "conforme"`)

**Saídas:**
- PR #91 mergeado (`--squash --delete-branch`)
- Branch `fix/88-03-ask-user-e-seletor-obrigatorios` removida

**O que ler:** `sdd/discovery/88-modo-pipeline-avancado/verdicts/job-20261119-000439-88-modo-pipeline-avancado-fix-88-03-review.json`

**Critério de aceitação:**
```bash
gh pr view 91 --json state | python3 -c "import json,sys; assert json.load(sys.stdin)['state']=='MERGED'"
```

## Decisões do projeto aplicadas
ADR-0004 (Pipeline Multi-Sessão). Rules: `sdd-fluxo-e-portao` (FLX-02).

## O que já existe
| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Veredito do Reviewer | `sdd/discovery/88-modo-pipeline-avancado/verdicts/` | `conforme` | mergear diretamente |

## Dependências novas
Nenhuma. `gh` já está em `allow_commands`.

## Rules e IDs da Constituição Aplicáveis
`FLX-02`, `GOV-01`.

## Critérios de Aceitação Executáveis
```bash
gh pr merge 91 --squash --delete-branch
gh pr view 91 --json state | python3 -c "import json,sys; d=json.load(sys.stdin); assert d['state']=='MERGED', d['state']"
```

## Status
- [x] PR #91 mergeado (squash + delete-branch) em 2026-11-19
