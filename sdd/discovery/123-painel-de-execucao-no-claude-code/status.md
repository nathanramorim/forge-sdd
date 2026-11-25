# Status — 123 Painel de execução no Claude Code

**Gate 1:** aprovado pelo Nathan em 24/11/2026 23h ("vamos rodar o split"). Alvo: 6.0.0-beta.
**Status:** split em 4 features (`features/feat-123-01..04`); implementação na trilha C.
**Próximo passo:** mod forge-painel (feat-123-02), validação na instalação (03) e forge up --terminal painel (04); ADR-0030 aguarda aprovação.
**A confirmar:** limites dos alertas (70% e 90%); limite do aviso de orçamento; spike no desktop do dono.

## Spike de carregamento (feat-123-01, Claude Code 2.1.332)
- `claude --plugin-dir <pasta>` carrega o mod só naquela sessão e funciona em `claude -p` (status, `turn.complete` e `session.end` rodaram; os módulos de ganchos esperam um sinal de liberação no início do turno).
- `CLAUDE_CODE_PLUGIN_DIRS` (env do processo ou bloco `env` de `~/.claude/settings.json`, nunca configuração do projeto) é o caminho onde não há flag: app desktop e SDK.
- `.claude/skills/<slug>` do projeto NÃO carregou (testado, cópia e link). `/plugin install` e marketplace gravam em `~/.claude`: fora (ADR-0025).
- Decisão: `forge up --terminal painel` imprime `claude --plugin-dir .agents/plugins/forge-painel`. ADR-0030 (proposta).
- A confirmar: desktop do dono (`CLAUDE_CODE_PLUGIN_DIRS`), limites 70/90 e orçamento 80.
