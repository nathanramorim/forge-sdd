# Status — 115 pipeline e seleção de modelos

**Status:** Gate 1 aprovado; gates passaram; jobs 115 e fix-115-01 enfileirados. Fix 115-01 na branch `fix/115-01-workers-por-iniciativa`; 115-02 (remoção do tmux) mesclada na main (#224).
**Decisão do dono (21/11/2026):** o tmux não é mais usado; o chat orquestra e escala os subagentes. Seleção de modelo por complexidade (tiers, 115-03).
**Fluxo:** spec/plano aprovados → Builder com tier por task → Camada A/PR → Reviewer; workers escopados à iniciativa.
**Base:** ADR-0004, ADR-0005, ADR-0014, ADR-0015 aceita; ADR-0008 substituída; ADR-0024 proposta.
**Próximo passo:** Reviewer no fix 115-01; Builder segue as features 115-03. Follow-ups da 115-02: ADR substituindo 0006/0007/0008 e menções ao tmux nas skills de status.
**A confirmar:** uma nota anterior dizia "tmux é opcional", o que contradiz a decisão acima; confirmar qual vale. Corte direto ou transição de uma versão para `terminal: tmux`.
