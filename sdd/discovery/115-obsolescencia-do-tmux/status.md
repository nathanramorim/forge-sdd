# Status — 115-obsolescencia-do-tmux

**Status:** conception pronta (discovery, criteria e plan). Aguarda **Gate 1** do dono.
**Decisão do dono (21/11/2026):** o tmux não é mais usado; o chat orquestra e escala os subagentes.
**Entrega:** o modo `sessions`/tmux sai do produto; `deliver.go` e `status_live.go` são adaptados; a configuração `terminal: tmux` recebe mensagem de transição; doc e release acompanham.
**Base:** ADR-0015 (modo chat) passa a aceita; ADR-0008 (sessões no tmux) é substituída.
**Próximo passo:** Gate 1, depois `/split-features` (já há plano com 8 tasks em 3 ondas).
**A confirmar:** transição de uma versão (aviso) ou corte direto; se `deliver.go` serve o modo `worker` headless.