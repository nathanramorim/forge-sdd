# Feature 5ae2-02 — `/status` Prescritivo

Elimina a necessidade de o usuário memorizar a ordem dos comandos: `/status` passa a sempre indicar o próximo passo exato.

## Critérios de Aceitação Executáveis

1. O prompt `/status` (replicado em Claude, Gemini e Copilot) deve, além do estado atual do projeto, emitir sempre uma linha final no formato `Próximo comando sugerido: <comando>`.
2. A sugestão deve ser calculada a partir do estado de `sdd/memory/progress.md`:
   - Nenhum discovery e nenhuma feature registrada → sugerir `/discovery`.
   - Discovery presente em `sdd/discovery/` sem features correspondentes em `sdd/features/` → sugerir `/split-features`.
   - Features com status `todo` → sugerir `/proxima-feature`.
   - Todas as features `done` → sugerir `/archive` ou `/discovery` (novo ciclo).
3. Os três arquivos de prompt (`.claude/commands/status.prompt.md.tmpl`, `.gemini/prompts/status.prompt.md.tmpl`, `.github/prompts/status.prompt.md.tmpl`) devem ser atualizados de forma equivalente.
