# Discovery 9b2f — Autonomia e Construção Contínua (Hardness Engineer)

Esta descoberta analisa a arquitetura e metodologia para permitir que o ciclo de desenvolvimento do Forge-SDD rode de forma contínua e autônoma por horas de construção, eliminando a dependência de intervenção humana a cada turno.

## 1. O Problema e o Valor
*   **Problema:** O ciclo SDD padrão (`READ-MIN -> PLAN -> ACT -> WRITE -> CLOSE`) foi desenhado para ser interativo. O Orquestrador planeja e interrompe para validação humana antes de delegar para o Builder. Embora isso garanta alinhamento, impede o desenvolvimento paralelo e de "fundo" (overnight coding), onde o desenvolvedor deseja acordar com a feature completa ou múltiplos cenários validados.
*   **Valor:** Permitir que o time de desenvolvimento ou o desenvolvedor solo configurem "loops autônomos seguros" onde a IA orquestra, implementa e revisa múltiplas features ou tarefas em fila de forma automatizada, notificando apenas no final do ciclo.

## 2. Abordagem por Ferramenta

### A. Claude (Anthropic API, Claude Code, Aider)
*   **Como funciona:** Na interface normal do Claude Desktop, a autonomia contínua é bloqueada por conta da arquitetura reativa do chat. Para rodar por horas, utiliza-se a API em conjunto com ferramentas CLI especializadas (como `claude-code` ou `aider`) integradas a um loop de script local.
*   **Mecanismo de Loop:** O Orquestrador é instruído a pular a validação do PLAN caso um arquivo de flag `.sdd-auto-pilot` esteja presente. Ele avança diretamente, rodando a branch, delegando a si mesmo o papel de Builder e de Revisor, executando os testes no terminal local. Ao fechar a sessão, em vez de parar, ele lê a próxima feature `todo` no `progress.md` e reexecuta o script do ciclo.

### B. Antigravity (Google Gemini AGY)
*   **Como funciona:** O Antigravity oferece o slash command nativo **`/goal`**.
*   **Mecanismo de Loop:** Quando o usuário executa `/goal [objetivo macro]`, o motor do Antigravity entra em modo persistente de background. Ele monta a checklist e executa as tarefas sequencialmente. O motor do Antigravity tem permissão nativa para rodar ferramentas, ler arquivos e compilar código em turnos sucessivos sem pedir confirmações, parando apenas quando todos os critérios de aceitação do objetivo principal forem satisfeitos ou caso ocorra um erro insolúvel de compilação/teste.

### C. GitHub Copilot
*   **Como funciona:** O chat tradicional na barra lateral do VS Code é reativo e sem acesso a loops de terminal autônomos puros. A execução contínua com Copilot requer wrappers externos (ex: Copilot Workspace ou scripts que interagem via CLI com a API do Copilot).

### D. Cenário CLI
*   **Como funciona:** Sim, a automação de loops de longa duração é infinitamente mais robusta e factível no ambiente de CLI. Através de um script de controle (daemon local), é possível monitorar o progresso em `sdd/memory/progress.md` e invocar sequencialmente os agentes (via APIs ou CLIs de chat) até que a fila de features com status `todo` seja zerada.
