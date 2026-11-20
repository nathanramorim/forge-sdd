# Status — 105-experiencia-cli-e-harness

**Status:** Split da onda 1 pronto (6 features 105-01 a 105-06), fila criada (6 tasks `queued`), aguardando Builder; ondas 2 e 3 no plan.md. PR em draft.
**Próximo passo:** entregar ao Builder (ordem 1.1, 1.4, 1.5, 1.2, 1.3, 1.6); 1.2, 1.4 e 1.6 em série com a 106/107 (templates, goldens, `doctor`).
**A confirmar:** nada.

## Decisões do Gate 1
- Base resolvida; ADRs 0016-0020 e `obsoletos.md` da 104 entram pelo PR #119 (dependência de documentação).
- Obsoletos: só corrigir texto (O1, O8, O9, O10, O11, O14) e avisar (O3, O4, O5, O7, O13). Mantém `lider` (só descrição), skills `sdd-*` e a lista "Mantém". Nada removido.
- `init`: menu de agente, detectado recomendado; sem TTY exige `--agent`.
- ADRs 0009 a 0014 aceitas (2026-11-20).
- Limiares H1 a H9 aprovados (sem dado = a medir); consertar tokens zerados (onda 3); arquivar 79 a 88 via `/archive` após corrigir o filtro do `status`.
