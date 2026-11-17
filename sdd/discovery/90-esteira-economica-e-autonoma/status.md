# Status — Iniciativa 90

**Branch:** `feat/90-esteira-economica-e-autonoma` (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** 90-01 a 90-08 implementadas; PR aberto contra a 88, em revisão.
**Próximo passo:** Reviewer revisa o PR; depois comparar `forge report --initiative 90-esteira-economica-e-autonoma --jobs` com a linha de base (RN-1).

## Confirmado no build (90-07)
- `claude` e `agy`: `--model` e `--effort low|medium|high|xhigh|max`; `gemini`: só `--model`.
- `codex` e `copilot` não estão instalados: nenhuma flag emitida; escolher modelo ou esforço neles é erro.
- Lint: `golangci-lint` (latest) sem achados.
