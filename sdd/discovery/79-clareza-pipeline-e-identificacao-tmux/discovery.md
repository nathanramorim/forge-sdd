# Discovery 79 — Clareza na Esteira do Pipeline, Identificação Visual no Tmux, Paralelismo de Filas, Monitor de Processos e Ergonomia do Menu

## Problema
1. **Ambiguidade entre Esteira vs Subagente:** Na abertura da sessão de Discovery, uma regra legada ("pergunte se vai para um subagente") induzia a IA a perguntar se o usuário queria delegar a demanda para um subagente. Isso gerou confusão cognitiva real: o usuário acreditou que delegar significava despachar para as sessões da esteira (Builder e Reviewer). O fluxo natural do Forge SDD é **seguir a esteira** (*pipeline first*). Se há itens em fila (`queued` para Builder ou `review_pending` para Reviewer), a sugestão deve ser seguir a esteira e despachar para essas sessões, em vez de fazer perguntas confusas sobre subagentes.
2. **Confusão entre Múltiplos Projetos no tmux:** Ao abrir múltiplos projetos em paralelo com `forge` / `forge-sdd up`, o tmux abre sessões independentes (`forge-<projeto>-<hash>`), mas visualmente a tela é idêntica (painéis 0, 1 e 2 com Discovery, Builder e Reviewer). A barra de status inferior do tmux e as bordas não evidenciam qual projeto está aberto, dificultando a navegação rápida quando o desenvolvedor transita entre janelas ou sessões.
3. **Menu Fragmentado e Hierarquia Rígida:** O menu interativo `cmd/forge-sdd/menu.go` divide as opções entre "Pipeline multi-sessão" e "Estrutura padrão (sessão única)". Essa divisão é um resquício histórico: o pipeline multi-sessão é o fluxo principal do Forge SDD v4+. Ações imediatas como ver status, fila, rodar doctor ou iniciar sessões ficam dispersas em submenus redundantes.
4. **Painéis Bloqueados com "Aguardando Job na Fila" e Falta de Fluidez:** Ter painéis inteiros dedicados exibindo loops de "Aguardando jobs na fila..." consome espaço visual e dá a sensação de processos pesados travando a máquina. O usuário quer o **terminal livre para trabalhar**, acompanhado de um **mini-painel de monitoramento leve**. Quando o Builder ou Reviewer executam ou demandam aprovação, a interação deve ser limpa, permitindo navegar entre processos/subagentes paralelos sem travar a tela principal.
5. **Gargalo no Paralelismo de Fila:** Embora as specs declarem `**Depende de:**`, a fila na CLI (`internal/queue`) consome itens estritamente em FIFO sequencial. A fila precisa de inteligência topológica (DAG) para liberar tarefas independentes em ondas paralelas automaticamente até o limite de `max_parallel_tasks`.
6. **Falta de Visibilidade de Recursos e Processos do Projeto:** O desenvolvedor não tem uma forma simples de inspecionar quais processos, threads, workers e CLIs estão rodando em background associados ao projeto, quanto de CPU/Memória estão gastando e se há processos zumbis ou orfãos consumindo máquina.

## Para quem
Desenvolvedores e líderes técnicos operando múltiplos projetos com o Forge SDD e gerenciando execução paralela de workers, recursos de máquina e esteira de entrega.

## Jornada Alvo

### 1. Início da Sessão Discovery (Pipeline First)
1. O agente executa `forge-sdd brief` e `forge-sdd status`.
2. Se houver tarefas ativas na esteira (`queued` pertencentes ao Builder ou `review_pending` ao Reviewer), o agente reporta o status e prioriza a esteira:
   - *"Existem 3 tasks em fila para o Builder e 1 aguardando o Reviewer. Deseja despachar/acompanhar a esteira ou conceber uma nova demanda?"*
3. Se não houver pendências externas ou o usuário quiser conceber, o agente assume a Concepção diretamente.

### 2. Tela no tmux com Identificação do Projeto, Terminal Livre e Mini-Monitor
1. Ao rodar `forge` ou `forge-sdd up`, a sessão tmux exibe:
   - **Barra de status inferior (`status-left`):** Nome do projeto e branch com destaque (ex.: `[ 📦 barbearia-app | feat/12-auth ]`).
   - **Bordas dos painéis (`pane-border-format`):** Cada painel exibe o nome do projeto (ex.: `[ barbearia-app ❯ forge-discovery ]`).
   - **Nome da janela:** `forge:<projeto>`.
2. Em repouso, o terminal direito fica livre para uso do desenvolvedor e o **mini-painel de monitoramento** sintetiza a esteira sem poluição visual.
3. Se um worker precisar de aprovação (Camada C), um badge de alerta visual destaca o painel sem travar o terminal ativo.

### 3. Paralelismo Inteligente de Fila por DAG
1. O Planejamento enfileira tarefas com dependências declaradas (`**Depende de:**`).
2. A fila (`internal/queue`) calcula as tarefas elegíveis em tempo real: todas as tasks cujas dependências foram aprovadas ficam prontas simultaneamente.
3. O Builder pode consumir e executar lotes paralelos (`--batch`) com subagentes/worktrees isoladas até o limite de `max_parallel_tasks`.

### 4. Monitor de Processos e Consumo (`forge-sdd ps`)
1. O desenvolvedor executa `forge-sdd ps` (ou seleciona no menu/monitor).
2. O sistema lista exatamente todos os processos associados ao projeto:
   - Workers e despachantes do Forge.
   - Painéis e CLIs ativas do tmux.
   - Processos de teste e compiladores acionados por subagentes.
   - PIDs, % de CPU, Memória RAM (MB), tempo ativo e estado.
   - Oferece opção segura de encerrar processos presos (`--kill-zombies`).

### 5. Menu Unificado e Direto (`forge`)
Menu direto em 1 clique para as funções mais usadas, eliminando a dicotomia "pipeline vs padrão".

## Regras de Negócio
- **RN-01 (Pipeline First):** Priorizar a esteira física do projeto; "subagente" nunca substitui Builder/Reviewer.
- **RN-02 (Identificação no tmux):** Estilização aplicada estritamente na sessão (`-t exact`), preservando configurações globais do usuário.
- **RN-03 (Unificação do Menu):** Menu orientado a ações com cabeçalho de estado e branch.
- **RN-04 (Terminal Livre e Monitoramento Leve):** Workers em repouso não poluem o terminal; alertas de aprovação são destacados no título/borda.
- **RN-05 (Ordenação Topológica da Fila):** Tarefas só são liberadas para construção se todas as suas dependências estiverem aprovadas (`approved`). Ciclos geram erro explícito.
- **RN-06 (Monitoramento Focado no Projeto):** `forge-sdd ps` só inspeciona processos pertencentes ao diretório e sessão do projeto, sem invadir outros processos do sistema operacional.

## Escopo
- **Dentro:**
  - Atualização dos prompts para lógica Pipeline First.
  - Identificação visual de projeto e ergonomia de painéis no tmux (`internal/pipeline/tmux.go`).
  - Menu unificado e direto (`cmd/forge-sdd/menu.go`).
  - Alertas visuais e fluxo de aprovação não bloqueante.
  - DAG e paralelismo por ondas na fila (`internal/queue/queue.go`).
  - Mapeamento e monitor de processos do projeto (`forge-sdd ps` em `internal/monitor/ps.go`).
  - Documentação completa e validação de suíte.
- **Fora:**
  - Substituição do tmux por outro multiplexador.
  - Gerenciador genérico de processos de todo o sistema operacional (foco exclusivo no projeto).
