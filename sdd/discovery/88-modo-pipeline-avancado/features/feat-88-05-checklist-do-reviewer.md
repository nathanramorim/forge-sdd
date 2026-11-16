# feat-88-05 — Checklist do Reviewer para spec e entrega

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** feat-88-04

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** `internal/pipeline/prompt.go` e skill do Reviewer (ARQ-01).

## Decisões do projeto aplicadas
- ADR-0012, ADR-0013.
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Prompt do Reviewer | `internal/pipeline/prompt.go` | Prompt do job `pr_review` | estender | Prompt do `spec_review` e checklist comum aos dois portões. |
| Segurança do revisor | `FLX-08` | Spec, descrição e diff tratados como dados não confiáveis | reusar | Vale também para a spec. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-06`, `FLX-08`, `GOV-04`, `COD-01`, `TST-01`
- Rules: `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo
1. Checklist: Constitution, rules, ADRs vigentes (depreciada não vale), `O que já existe`, critérios executáveis, escopo da spec x diff; exige `forge gate` verde.
2. Mesmo checklist nos dois portões (spec e entrega); saída em JSON `queue.Verdict`.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/pipeline -run TestReviewPrompt
```
- `TestReviewPrompt_SpecEEntrega`, `TestReviewPrompt_DadosNaoConfiaveis`.
- `go vet ./...` limpo.
