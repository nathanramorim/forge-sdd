# Status — 105-experiencia-cli-e-harness

**Status:** Onda 1 em 7 features (1.1 a 1.5 nos PRs #123 a #127, 1.6 no #128, 1.7 no #129, todos draft); onda 3 ganhou a 3.6 (spec 105-08, log de achados). PR #121 em draft.
**Próximo passo:** Builder da 1.7 parte da branch da 1.3. A fila da onda 3 nasce no split da onda 3 (ainda sem fila).
## Decisões do Gate 1
- Base resolvida; ADRs 0016-0020 e `obsoletos.md` da 104 entram pelo PR #119.
- Obsoletos: só corrigir texto e avisar; `lider` só muda a descrição. Nada removido.
- `init`: menu de agente; sem TTY exige `--agent`. ADRs 0009 a 0014 aceitas (2026-11-20).
- Limiares H1 a H9 aprovados (sem dado = a medir); consertar tokens zerados (onda 3).
- Dono: toda spec nova de feature e fix traz "Descrição de negócio (release notes)" (task 2.7, spec 105-09).
- CSV da 1.7 e da 3.6 em `sdd/memory/logs/` versionados com `.gitattributes` `merge=union` (decisão do dono).
