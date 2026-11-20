# Status — 105-experiencia-cli-e-harness

**Status:** Gate 1 aprovado pelo dono em 2026-11-20. Nada implementado nem enfileirado. PR em draft.
**Próximo passo:** `/split-features` da onda 1, em série com a 106 (ela também mexe em templates, goldens, `split-features.md`, `hooks.go`, `doctor` e menus; tasks 106 1.7, 2.5 e 2.6 dependem da ordem).
**A confirmar:** nada.

## Decisões do Gate 1
- Base resolvida; ADRs 0016-0020 e `obsoletos.md` da 104 entram pelo PR #119 (dependência de documentação).
- Obsoletos: só corrigir texto (O1, O8, O9, O10, O11, O14) e avisar (O3, O4, O5, O7, O13). Mantém `lider` (só descrição), skills `sdd-*` e a lista "Mantém". Nada removido.
- `init`: menu de agente, detectado recomendado; sem TTY exige `--agent`.
- ADRs 0009 a 0014 aceitas (2026-11-20).
- Limiares H1 a H9 aprovados (sem dado = a medir); consertar tokens zerados (onda 3); arquivar 79 a 88 via `/archive` após corrigir o filtro do `status`.
