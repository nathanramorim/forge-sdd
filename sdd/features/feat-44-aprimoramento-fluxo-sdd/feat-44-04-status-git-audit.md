# Feature 44-04 — `/status` Audita PRs Abertos e Versão Publicada

**Branch:** `feat/44-04-status-git-audit`
**Depende de:** —
**Paralelizável (worktree):** Sim — só toca o prompt `/status`, sem overlap com as demais.

## Descrição
Hoje `/status` reporta apenas o estado de `progress.md`, sem checar o estado real do repositório. Esta feature faz `/status` auditar PRs abertos e se a versão local já foi publicada, antes de sugerir iniciar um novo discovery/feature.

## Critérios de Aceitação Executáveis

1. `/status` (Claude, Gemini, Copilot) executa `gh pr list --state open` (quando `gh` disponível) e lista os PRs em aberto no relatório.
2. `/status` compara `sdd/.sdd-version` com `npm view forge-sdd version` (quando há rede) e sinaliza explicitamente se a versão local ainda não foi publicada.
3. Se `gh` não estiver instalado/autenticado ou não houver rede, o item correspondente é reportado como "não verificável", sem falhar o comando inteiro.
4. O relatório final de `/status` recomenda resolver PRs abertos e/ou publicar a release pendente antes de sugerir iniciar um novo `/discovery` ou `/nova-feature`.
