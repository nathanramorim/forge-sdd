# Status — 105-experiencia-cli-e-harness

**Status:** discovery, criteria, plan, obsoletos e harness escritos; aguardando Gate 1 (FLX-01). Nada implementado nem enfileirado. Branch `feat/105-experiencia-cli-e-harness` (da `main`), commit local.
**Próximo passo:** o dono responde `A confirmar`; depois `/split-features` da onda 1.

## A confirmar (decisões do dono)
1. Base: a onda 3 precisa da 104 na `main` (`doctor --harness`, `eval`, ADR-0016 a 0020 estão só em `bbd26d6`). Esperar o merge ou rebasear a 105?
2. O que remover de fato: `lider`, `queue stations`, `attach`, `queue handoff`, aliases `/implement` e `/proxima-feature` (proposta: só avisar; ver `obsoletos.md`).
3. Padrão de `init --agent` (hoje `copilot`).
4. Aprovar as ADRs 0009 a 0014 (em uso, `proposto`).
5. Limiares H1 a H9 de `harness.md` e consertar a telemetria com tokens zerados.
6. Arquivar as iniciativas 79 a 88 concluídas.
