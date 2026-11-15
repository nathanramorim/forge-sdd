# Discovery 84 — Painel Interativo Amigável: Stream Narrativo Estilo Chat nos Workers e Fila com Drill-Down

## 1. Problema
No estado atual das interfaces de acompanhamento do pipeline Forge-SDD:
1. **Painel ao vivo sobrecarregado e estático (`forge status --live`):** A lista de tarefas é truncada em "... e mais 19 tarefas", misturando tabelas de processos brutos (CPU/RAM/PIDs) com o fluxo de entrega de software, dificultando entender de relance onde a esteira está e o que foi concluído.
2. **Caixa-preta nos workers do tmux (`forge-builder` e `forge-reviewer`):** Os painéis de execução emitem contadores secos de tempo (`... trabalhando há 4m40s • 3 arquivos alterados`) ou saídas cruas de bash/git. Isso cria uma experiência fria e confusa em comparação direta com o `forge-discovery`, onde o usuário usufrui de uma interação natural, fluida e amigável com o agente.
3. **Ausência de inspeção rápida de histórico:** Não existe um modo interativo no terminal para inspecionar os detalhes de cada tarefa executada (arquivos tocados, laudo do reviewer, tempo de build, erros) sem vasculhar manualmente pastas internas do sistema.

## 2. Para Quem
Desenvolvedores e líderes técnicos que utilizam o Forge SDD para orquestrar o desenvolvimento autônomo multi-agente e desejam visibilidade clara, agradável e humanizada do progresso da esteira em tempo real no terminal.

## 3. Jornada do Usuário
1. **Abertura do Pipeline:** O usuário executa `forge` ou `forge up` e as 3 sessões são inicializadas (em 3 painéis no tmux ou no Windows Terminal).
2. **Acompanhamento Natural no Builder (`forge-builder`):**
   - Ao receber uma tarefa da fila, o painel do Builder não cospe dumps ininteligíveis.
   - O usuário acompanha um stream estruturado em formato de balões de etapas com status humano e emojis claros:
     - 💭 *Analisando dependências e requisitos de `fix-abcc-01`...*
     - 📋 *Lendo spec e critérios de aceitação executáveis*
     - 🔨 *Editando `BottomNav.tsx` e `AppShell.tsx` (2 arquivos modificados)*
     - 🧪 *Rodando validação automatizada (`npm test`)*
     - ✅ *Task concluída com sucesso! Handoff despachado para o Reviewer.*
3. **Acompanhamento no Reviewer (`forge-reviewer`):**
   - O Reviewer exibe em cards limpos os portões avaliados progressivamente (Portão de Governança, Type-Check, Testes, Lint, Segurança) e apresenta o veredito final com clareza.
4. **Inspeção com Drill-Down na Fila (`forge status -i` ou `forge queue --panel`):**
   - O usuário visualiza todas as tarefas da esteira (100% visíveis, sem truncamento).
   - Navega com `↑` e `↓` entre as tarefas.
   - Pressiona `Enter` para inspecionar a tarefa selecionada: card expandido com tempo total, arquivos modificados, resumo do commit, laudo de governança e logs relevantes.
   - Pressiona `Esc` ou `q` para retornar à lista ou sair.

## 4. Regras de Negócio
- **RN-01 (Transparência Conversacional):** O stream nos workers deve traduzir eventos brutos de processos (leitura de arquivos, escrita, execução de testes, git) em etapas descritivas legíveis por humanos, sem suprimir eventuais mensagens de erro reais.
- **RN-02 (Sem Truncamento Cego):** A lista de tarefas no modo interativo deve suportar paginação/scroll e permitir visualizar todas as tarefas da iniciativa atual.
- **RN-03 (Drill-Down Não Bloqueante):** A inspeção detalhada de uma tarefa não deve interromper nem concorrer com a leitura contínua da fila pelo Builder e Reviewer.
- **RN-04 (Portabilidade e Robustez):** A renderização visual deve utilizar sequências ANSI padrão e componentes Charm (`huh`), operando de forma idêntica no macOS, Linux, WSL e Windows Terminal.
- **RN-05 (Compatibilidade Arquitetural):** Não alterar a máquina de estados nem o esquema de `queue.json` (`ADR-0004`); estender apenas a camada de apresentação e telemetria visual.

## 5. Escopo
### Dentro do Escopo
1. **Formatador de Stream Narrativo para Workers (`internal/monitor/chat_feed.go`):**
   - Captura e formatação de eventos de ciclo de vida do worker em balões de chat e etapas amigáveis.
   - Apresentação de pensamentos, ações ativas (edição, compilação, teste) e sumário de conclusão.
2. **Navegador Interativo de Fila com Drill-Down (`internal/survey/queue_drilldown.go`):**
   - Componente interativo baseado em `huh` para listagem rolável de tarefas com badges visuais de status (`[✓] concluído`, `[▶] em andamento`, `[⌛] na fila`, `[✗] falhou`).
   - Modal/visualização expandida ao pressionar `Enter` com detalhes completos da task.
3. **Integração de Comandos e Menu CLI:**
   - Adicionar flag `-i/--interactive` em `forge status` e subcomando amigável no menu `forge`.
   - Suporte a abertura do painel interativo em janela dedicada do tmux quando invocado com `--tmux`.
4. **Suíte de Testes Unitários:**
   - Testes de renderização de cards, parser de eventos do stream narrativo e navegação da fila.

### Fora do Escopo
- Criar servidores HTTP ou dashboards gráficos baseados em navegador web.
- Alterar o contrato de comunicação física dos agentes no protocolo ACP.
- Modificar o fluxo de branches git ou regras do portão de governança.

## 6. Riscos e Mitigações
| Risco | Impacto | Mitigação |
|---|---|---|
| Excesso de saída gerando "flicker" no terminal | Médio | Uso de buffer de tela alternativo (`Alternate Screen Buffer`) e debounce na atualização de frames. |
| Terminais antigos com caracteres Unicode corrompidos | Baixo | Sanitização e fallback para glifos ASCII seguros quando terminal não suportar UTF-8. |
| Sobrecarga de I/O em filas grandes | Baixo | Carga preguiçosa (*lazy loading*) dos detalhes da task apenas sob demanda (ao teclar `Enter`). |
