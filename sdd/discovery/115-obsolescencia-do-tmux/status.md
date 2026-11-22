# Status — 115 pipeline e seleção de modelos

**Gate 1:** aprovado
**Status:** Gate 1 aprovado; fix-115-01 (#216) e 115-02 (#224) mesclados na main. Falta a feat-115-04 (tmux como opção) e as 115-03.
**Decisão do dono (22/11/2026):** o tmux continua opcional. A 115-02 o removeu do código; a 115-04 o devolve como opção, e o padrão segue em janelas do sistema.
**Fluxo:** spec/plano aprovados → Builder com tier por task → Camada A/PR → Reviewer.
**Base:** ADR-0004, ADR-0005, ADR-0014; ADR-0024 proposta (tmux é opção de interface).
**Próximo passo:** Reviewer na spec (#211); Builder segue as 115-03 e a 115-04.
**A confirmar:** ADR substituindo 0006/0007/0008 (tmux), conforme o desfecho da 115-04; menções ao tmux nas skills de status.
