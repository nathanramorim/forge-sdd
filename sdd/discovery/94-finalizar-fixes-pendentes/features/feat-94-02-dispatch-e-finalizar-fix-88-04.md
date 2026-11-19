# feat-94-02 — Dispatch e finalizar fix-88-04

**Objetivo:** despachar fix-88-04 para o Builder, aguardar ciclo completo (Builder→Reviewer) e mergear o PR resultante.

**Entradas:**
- Spec: `sdd/discovery/88-modo-pipeline-avancado/fixes/fix-88-04-ask-user-merge-apos-aprovacao.md`
- Binário Go: `~/go/bin/forge-sdd` (tem flag `--fix`)

**Saídas:**
- Job fix-88-04 enfileirado no Builder
- PR criado pelo Builder e mergeado após aprovação do Reviewer

**O que ler:** `sdd/discovery/88-modo-pipeline-avancado/fixes/fix-88-04-ask-user-merge-apos-aprovacao.md`

**Critério de aceitação:**
```bash
# Dispatch (usar binário dev que tem --fix)
~/go/bin/forge-sdd dispatch 88-modo-pipeline-avancado --fix fix-88-04
# Aguardar PR aberto pelo Builder e veredito conforme do Reviewer
# Merge do PR quando conforme
```

**Depende de:** feat-94-01

## Decisões do projeto aplicadas
ADR-0004 (Pipeline Multi-Sessão). Rules: `sdd-fluxo-e-portao` (FLX-01, FLX-02).

## O que já existe
| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Flag --fix no forge instalado | `/usr/local/bin/forge-sdd` | Ausente (npm) | usar `~/go/bin/forge-sdd` que tem a flag |
| Spec fix-88-04 | `sdd/discovery/88-modo-pipeline-avancado/fixes/` | Pronta | despachar diretamente |

## Dependências novas
Nenhuma. `gh` e `forge-sdd` já em `allow_commands`; binário Go compilado disponível em `~/go/bin/forge-sdd`.

## Rules e IDs da Constituição Aplicáveis
`FLX-01`, `FLX-02`, `GOV-01`.

## Critérios de Aceitação Executáveis
```bash
# 1. Dispatch
~/go/bin/forge-sdd dispatch 88-modo-pipeline-avancado --fix fix-88-04
# 2. Verificar enfileirado
ls .forge/runtime/queues/builder/pending/ | grep fix-88-04
# 3. Após Builder e Reviewer processarem:
# gh pr merge <numero> --squash --delete-branch
# 4. Verificar veredito conforme antes do merge:
# cat sdd/discovery/88-modo-pipeline-avancado/verdicts/*fix-88-04* | python3 -c "import json,sys; d=json.load(sys.stdin); assert d['veredito']=='conforme'"
```

## Status
- [x] Dispatch executado em 2026-11-19: job `job-20261119-011901-88-modo-pipeline-avancado-fix-88-04` enfileirado em `builder/pending`
- [ ] PR de fix-88-04 aberto pelo Builder (pendente — próximo ciclo do Builder)
- [ ] Veredito `conforme` do Reviewer (pendente)
- [ ] Merge do PR de fix-88-04 (pendente — Líder executa após veredito conforme)
