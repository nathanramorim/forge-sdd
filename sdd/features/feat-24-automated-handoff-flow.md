# Feature: Fluxo de Handoff Automático (Commit, Push e PR)

**Status:** `done`

## Contexto
Atualmente, a finalização de uma feature exige passos manuais de commit, push e abertura de PR. Queremos que o Orquestrador automatize essa sequência para garantir consistência e agilidade no desenvolvimento.

## Objetivos
- Automatizar a geração de mensagens de commit baseadas no trabalho realizado.
- Automatizar o push da branch de feature.
- Solicitar explicitamente ao usuário se deseja abrir um Pull Request após o push bem-sucedido.

## Critérios de Aceite
- [x] O Orquestrador deve gerar uma mensagem de commit semântica ao finalizar uma feature.
- [x] O Orquestrador deve realizar o `git add`, `git commit` e `git push` automaticamente.
- [x] Após o push, o Orquestrador deve perguntar ao usuário: "Deseja abrir o Pull Request agora?".
- [x] As instruções em `.gemini/skills/orquestrador.chatmode.md` e `.github/chatmodes/orquestrador.chatmode.md` devem ser atualizadas.
- [x] Os prompts de finalização (`revisar.prompt.md`) devem refletir este novo fluxo.

## Plano de Implementação
1. **Research:** Revisar os comandos git disponíveis no ambiente.
2. **Strategy:** Atualizar a seção "Finalizar" da skill do Orquestrador.
3. **Execution:**
   - Atualizar `internal/scaffold/templates/agents/gemini/.gemini/skills/orquestrador.chatmode.md.tmpl`.
   - Atualizar `internal/scaffold/templates/.github/chatmodes/orquestrador.chatmode.md.tmpl`.
   - Atualizar `internal/scaffold/templates/agents/gemini/.gemini/prompts/revisar.prompt.md.tmpl`.
   - Aplicar as mesmas mudanças nos arquivos ativos do próprio projeto `forge-sdd`.
4. **Validation:**
   - Finalizar uma feature de teste e verificar se o fluxo automático é disparado.
