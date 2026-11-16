# feat-88-04 — Job `spec_review` e aplicação do parecer

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** feat-88-01

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** `internal/bus` e `internal/pipeline` (ARQ-01): o Go decide e grava (SEG-06).

## Decisões do projeto aplicadas
- ADR-0004, ADR-0012, ADR-0013.
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Roteamento de jobs | `bus.RoleForType`, tipos `feature_build`/`pr_rework`/`pr_review` | Mapa tipo→papel | estender | Novo tipo `spec_review` → reviewer. |
| Aplicar veredito | `pipeline.ApplyVerdict`, `queue.Verdict`, `persistVerdict` | Go aplica e guarda cópia | reusar | `ApplySpecVerdict` irmão, mesmo contrato `queue.Verdict`. |
| Limite de tentativas | `reviewer.max_rework_attempts` | Já existe | reusar | Sem chave nova. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `SEG-01`, `SEG-06`, `DUP-01`, `TST-01`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo
1. Tipo de job `spec_review`, criado só em `advanced` e só depois de `forge gate` sem bloqueio.
2. `ApplySpecVerdict`: `aprovado` libera a fila e grava `review-spec.md` (escrito pelo Go); `ajustar` devolve feedback à Discovery até o limite; `escalar` ou limite vira `blocked` + inbox.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/bus ./internal/pipeline -run 'TestSpecReview|TestApplySpecVerdict'
```
- `TestSpecReview_RoteiaParaReviewer`, `TestSpecReview_SoEmAvancado`, `TestApplySpecVerdict_Aprovado`, `TestApplySpecVerdict_AjustarELimite`.
- `go vet ./...` limpo.
