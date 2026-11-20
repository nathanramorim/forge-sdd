# Status — 106-pr-draft-revisao-e-ci

**Status:** Gate 1 (`FLX-01`) aprovado pelo dono em 2026-11-20 com as 5 recomendações. Nada implementado. PR em draft.
**Próximo passo:** `/split-features` da onda 1 (sequenciar as tasks 1.7 e 2.5 com a 105).

## Decisões do Gate 1
1. Prefixo da branch: manter `feat/`.
2. Draft em repositório pessoal privado: provar com PR de teste (task 2.6).
3. Job "pronto?": sim, só no check obrigatório da `main`.
4. Primeira leva: hub-busize, hub-enterprise, forge-sdd, gerador-videos.
5. `lessons.md`: migrar para `sdd/memory/licoes/` (ajusta `GOV-03` via `GOV-01`).

## A confirmar
- PRs de task do leque também nascem em draft?
- Hooks só valem no Claude Code: Gemini e Copilot dependem dos passos dos comandos?

## Colisão com a 105
Independente: lições, CI, `pr_guard`, veredito com SHA. Colide: templates e goldens, `hooks.go`, `doctor`, numeração de ADR (livre após 0020).
