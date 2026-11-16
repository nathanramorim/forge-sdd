# feat-88-03 — ADR-0013: Reviewer como revisor da spec no modo avançado

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** nenhuma

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Documento em `sdd/spec/adr/` (FLX-05, GOV-01); sem código.

## Decisões do projeto aplicadas
- ADR-0004, ADR-0012 (Reviewer somente leitura, Go grava o veredito).
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| ADRs sobre revisão e gates | `sdd/spec/adr/adr-0004`, `adr-0012`; Constitution `FLX-01`, `FLX-06` | Gate 1 humano; Camadas A/B/C | estender por nova ADR | FLX-01 só muda para o modo avançado; modo normal intacto. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-01`, `FLX-05`, `FLX-06`, `GOV-01`, `SEG-06`
- Rules: `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo
1. Criar `sdd/spec/adr/adr-0013-reviewer-revisa-spec-no-modo-avancado.md` seguindo `adr-0000-modelo.md`, status `proposto`.
2. Decisão: no modo `advanced` o Reviewer (sessão própria, contexto limpo, somente leitura) substitui a aprovação humana do Gate 1; Camada A obrigatória; `escalar`/limite vira `blocked`; modo normal exige o humano.
3. Consequências: risco de mesmo modelo escrever e revisar; mitigação recomendada (agentes diferentes).

## Critérios de Aceitação Executáveis
```bash
forge gate
```
- ADR existe, segue o modelo e cita FLX-01, FLX-05 e GOV-01.
- `go vet ./...` limpo.
