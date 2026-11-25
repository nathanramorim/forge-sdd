# Status — 123 Painel de execução no Claude Code

**Gate 1:** aprovado pelo Nathan em 24/11/2026 23h ("vamos rodar o split"). Alvo: 6.0.0-beta.
**Status:** implementada na 6.0.0-beta (feat-123-01..04): mod forge-painel (25 testes, `claude plugin validate` ok), contrato de mod e validação na instalação, `forge up --terminal painel`. O mod grava `janela.json` para o teto de paralelo da 122.
**Próximo passo:** aprovação da ADR-0030 e spike no desktop do dono (`CLAUDE_CODE_PLUGIN_DIRS`).
**Decisões (24/11, dono):** alertas em 70% e 90%; aviso de orçamento em 85%, sem bloquear.

## Spike de carregamento (feat-123-01, Claude Code 2.1.332)
- `claude --plugin-dir <pasta>` carrega o mod só naquela sessão; funciona em `claude -p`.
- `CLAUDE_CODE_PLUGIN_DIRS` é o caminho do app desktop e do SDK.
- `.claude/skills/<slug>` do projeto não carrega; `/plugin install` grava em `~/.claude`: fora (ADR-0025).
- `forge up --terminal painel` imprime `claude --plugin-dir .agents/plugins/forge-painel`. ADR-0030 (proposta).
