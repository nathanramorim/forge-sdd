# Discovery 44 — Aprimoramento do Fluxo Decisório do Forge-SDD

## Por quê
Hoje o usuário chega ao `/discovery` já com a ideia formada e o agente aceita a demanda sem contestar suposições, o `/split-features` quebra tarefas sem indicar quais podem rodar em paralelo, e o desenvolvimento paralelo de várias features sempre mira `main` — mesmo quando várias PRs deveriam convergir para uma única release. Além disso, `/status` não confere o estado real do repositório (PRs abertos, versão ainda não publicada) antes de sugerir novo trabalho, o que já causou o cenário desta sessão: seguir para uma nova fase sem confirmar se a anterior estava de fato fechada.

## Para quem
Para o próprio mantenedor do forge-sdd (e para qualquer time que adote a metodologia) que conduz discoveries, quebra features e orquestra desenvolvimento paralelo via múltiplos agentes/worktrees, e que precisa de checkpoints de decisão explícitos em vez de suposições silenciosas do agente.

## Como (macro)
1. **Sabatina sempre ativa, com opção de pular:** ao final de `/discovery`, `/split-features` e `/nova-feature`, o agente questiona decisões relevantes (ex.: escopo, paralelismo, branch de destino) apresentando um caminho sugerido; o usuário responde escolhendo a sugestão, respondendo "pular" para aceitar o default, ou dando sua própria decisão.
2. **Paralelismo mecânico via script:** `/split-features` valida dependências entre tasks e marca quais podem rodar em paralelo. Abrir/fechar os git worktrees correspondentes é feito por um script mecânico (`sdd/scripts/worktree.sh`), não pelo agente gerando comandos token a token — o orquestrador apenas invoca o script como ação sugerida (ou o usuário aciona diretamente no chat), e o fechamento é sugerido ao concluir a implementação de cada task, junto com a geração do PR.
3. **PRs direcionados a uma branch de release:** quando várias features/fixes convergem para uma única entrega, a branch de destino dos PRs passa a ser uma `active_release_branch` configurada em `sdd/.sddrc`, em vez de `main`. `/novo-fix` e `/nova-feature` leem essa chave para decidir a base do PR.
4. **`/status` audita o repositório antes de liberar novo trabalho:** verifica PRs abertos (via `gh pr list`) e compara a versão em `sdd/.sdd-version` com a versão publicada no NPM (`npm view forge-sdd version`), sinalizando se há uma release pendente de publicação antes de sugerir iniciar um novo discovery/feature.

## Critérios de sucesso
- Nenhum discovery, split ou nova feature termina sem uma sabatina explícita das decisões de escopo/paralelismo/branch — mesmo que o usuário opte por pular rapidamente.
- Abrir/fechar worktrees nunca depende de o agente compor comandos git manualmente; sempre passa pelo script.
- PRs de trabalho paralelo apontam corretamente para a release branch ativa quando configurada, e para `main` quando não há release em andamento.
- `/status` nunca sugere iniciar algo novo sem antes reportar PRs abertos e o estado de publicação da versão atual.
