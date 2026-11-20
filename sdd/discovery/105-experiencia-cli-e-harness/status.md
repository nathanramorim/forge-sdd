# Status — 105-experiencia-cli-e-harness

**Status:** Split da onda 1 com 7 features; 1.1 a 1.5 entregues em PRs #123 a #127 (draft); 1.6 em andamento; 1.7 especificada. PR em draft.
**Próximo passo:** Builder da 1.7 parte da branch da 1.3 (ordem 1.1, 1.4, 1.5, 1.2, 1.3, 1.7, 1.6).
**A confirmar:** CSV da 1.7 em `sdd/memory/logs/` versionado ou ignorado no git (ruído e conflito entre PRs; ver risco na spec).

## Decisões do Gate 1
- Base resolvida; ADRs 0016-0020 e `obsoletos.md` da 104 entram pelo PR #119 (dependência de documentação).
- Obsoletos: só corrigir texto (O1, O8, O9, O10, O11, O14) e avisar (O3, O4, O5, O7, O13). Mantém `lider` (só descrição), skills `sdd-*` e a lista "Mantém". Nada removido.
- `init`: menu de agente, detectado recomendado; sem TTY exige `--agent`.
- ADRs 0009 a 0014 aceitas (2026-11-20).
- Limiares H1 a H9 aprovados (sem dado = a medir); consertar tokens zerados (onda 3); arquivar 79 a 88 via `/archive` após corrigir o filtro do `status`.
