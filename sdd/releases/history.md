# Release Notes — forge-sdd

Este arquivo registra o histórico de entregas de produto deste projeto.

## Entregas

### Próxima versão (não publicada)

*   **Correção na Troca de Convenção de Nomenclatura:** Agora é possível trocar a convenção de nomenclatura (`sequencial`, `hash` ou `workitem`) de um projeto já inicializado usando `forge-sdd update --naming-convention <valor>` — antes essa opção era aceita mas silenciosamente ignorada em projetos existentes.

### Versão 1.9.1 (Estável)

Esta release estável consolida todo o ciclo de betas desde a v1.7.0: nomenclatura configurável, redução da curva de aprendizado para quem está começando, correções de robustez na telemetria e no diagnóstico do projeto, e correções no comando `update`.

*   **Convenção de Nomenclatura Configurável:** Escolha entre nomenclatura `sequencial`, `hash` ou `workitem` no `init`, com auto-healing e detecção de deriva de convenção no `doctor` para projetos existentes.
*   **Onboarding Mais Simples:** `init` agora imprime um cheat-sheet dos comandos disponíveis, `/status` sempre sugere o próximo passo, e um novo comando `/tutorial` guia um ciclo SDD completo fictício para quem está começando.
*   **Modo Iniciante:** `/constitution` pode gerar explicações em linguagem simplificada, com exemplos no lugar de jargão técnico.
*   **Comando `forge-sdd autopilot`:** Ativa o modo autopilot somente após um número mínimo de ciclos completos registrados em telemetria, com bypass consciente via flag.
*   **Comando `/novo-fix`:** Fluxo dedicado para criar branches e especificações de correção (`fix-`).
*   **Telemetria Mais Confiável:** Estimativa real de tokens de entrada/saída e ativação/desativação dinâmica baseada no `.sddrc`.
*   **Diagnósticos Adicionais no `doctor`:** Detecção de nome padrão de projeto não alterado (`"meu-projeto"`), de métricas granulares corretamente referenciadas em subpastas, e verificação do arquivo de métricas no local correto do projeto.
*   **Atualização Confiável para Versões Beta:** Agora é possível atualizar o CLI diretamente para a última versão beta publicada usando `update --upgrade`, sem precisar informar o número da versão manualmente. Falhas de conexão ao verificar novas versões passam a ser avisadas claramente, em vez de serem ignoradas silenciosamente.

### Versão 1.9.1-beta (Beta)

Esta release beta adiciona convenção de nomenclatura configurável, auto-healing no doctor e melhorias na telemetria.

*   **Convenção de Nomenclatura Configurável:** Configuração de nomenclatura (`sequencial`, `hash` ou `workitem`) no `init` e preservação/auto-healing nos comandos do `doctor`.
*   **Melhorias na Telemetria:** Estimativa real do input/output de tokens e ativação/desativação dinâmica de telemetria baseada no `.sddrc`.
*   **Comando `/novo-fix`:** Adicionado o comando `/novo-fix` e melhor suporte a branches e arquivos `fix-` no fluxo SDD.

### Versão 1.9.0-beta (Beta)

Esta release beta reduz a curva de aprendizado do Forge-SDD para quem está começando, na era dos agentes de IA autônomos (Discovery 5ae2).

*   **Cheat-Sheet de Comandos:** O `forge-sdd init` agora imprime a lista completa dos comandos SDD disponíveis para os agentes escolhidos ao final da inicialização.
*   **`/status` Prescritivo:** O comando `/status` (Copilot, Claude e Gemini) agora sempre sugere o próximo comando a ser executado, com base no estado real do progresso do projeto.
*   **Diagnóstico de Deriva de Nomenclatura:** O comando `forge-sdd doctor` passa a detectar quando um projeto mistura a nomenclatura sequencial (`feat-NN`) com a nomenclatura por hash (`feat-xxxx`), alertando o usuário sobre a inconsistência.
*   **Onboarding Guiado:** Novo comando `/tutorial` (Copilot, Claude e Gemini) guia o usuário por um ciclo SDD completo e fictício, isolado dos dados reais do projeto.
*   **Comando `forge-sdd autopilot`:** Novo comando CLI que ativa o modo autopilot somente após um número mínimo de ciclos completos registrados em telemetria, com bypass consciente disponível via flag (independente do loop de autopilot em si, que segue em teste na branch `feat/cli-autopilot-autonomy`).
*   **Modo Iniciante:** O comando `/constitution` (Copilot, Claude e Gemini) passa a perguntar, opcionalmente, se o usuário prefere explicações em linguagem simplificada, com exemplos concretos no lugar de jargão técnico.

### Versão 1.7.1-beta.5 (Beta)

Esta release beta traz correções e alinhamento no fluxo de descobertas (discovery).

*   **Plano de Discovery Padronizado:** O prompt de `/discovery` de todos os agentes (Copilot, Claude e Gemini) foi atualizado para gerar obrigatoriamente o arquivo `plan-XX-*.md` na pasta `sdd/discovery/` contendo o roadmap e sugestão de quebra de tarefas/features.

### Versão 1.7.1-beta.4 (Beta)

Esta release beta traz atualizações na governança de releases de teste.

*   **Agrupamento Opcional de Releases Beta:** Flexibilizada a Regra 11 da Constituição para permitir, de forma opcional, agrupar e acumular múltiplos fixes ou features na mesma branch ou versão beta antes de gerar um novo bump ou tag, otimizando o fluxo.

### Versão 1.7.1-beta.3 (Beta)

Esta release beta traz melhorias de diagnósticos da integridade do projeto.

*   **Diagnóstico de Nome Padrão do Projeto:** O comando `doctor` e os prompts `/doctor` dos agentes passam a verificar se o projeto ainda utiliza o nome genérico `"meu-projeto"`, listando as correções e os cabeçalhos de agentes recomendados a serem renomeados.

### Versão 1.7.1-beta.2 (Beta)

Esta release beta traz a correção e robustez no mapeamento de métricas de telemetria das sessões.

*   **Métricas Granulares em Subpastas:** O Orquestrador agora referencia o caminho relativo completo da feature ou subtask (ex: `sdd/features/feat-1234-auth/task-01.md`) nas métricas registradas.
*   **Registro Incondicional de Sessões:** Métricas são gravadas incondicionalmente no encerramento de sessão, inclusive cobrindo timeouts, inatividade ou encerramentos sem progresso, registrando o esforço de telemetria.

### Versão 1.7.0 (Estável)

Esta release oficial e estável consolida o ciclo de desenvolvimento de diagnósticos, refinamento e agrupamento físico de especificações.

*   **Comandos `doctor` e `destroy`:** Novo utilitário CLI para diagnóstico de saúde da estrutura SDD do projeto (`doctor`) e purga segura e interativa dos scaffolds (`destroy`).
*   **Refinamento do Comando `init`:** Criação automática e isolamento de pastas de projetos com base nos parâmetros ou formulários interativos, além de sumário de fechamento visual das stacks configuradas.
*   **Agrupamento de Features/Discoveries:** Suporte completo para agrupar fisicamente tarefas e subfeatures complexas em pastas nomeadas para organizar o escopo de entregas e planos de discovery (Regra 13 da Constituição).
*   **Prompts de IA Atualizados:** Templates de comandos (`/split-features`, `/nova-feature`, `/discovery`) e instruções de sistema atualizados para todos os 4 agentes de IA suportados (Gemini, Claude, Copilot, OpenAI).

### Versão 1.6.1-beta.3 (Beta)

- **Fase 41 — Agrupamento de Features e Discovery**:
  - **Agrupamento Físico de Features/Tasks:** As subfeatures e tarefas geradas por quebra de plano ou features complexas agora são agrupadas em subpastas com o nome da feature (ex: `sdd/features/feat-XX-nome-feature/task-YY.md`).
  - **Agrupamento pelo Discovery:** Ao quebrar um plano de discovery, a subpasta de feature criada sob `sdd/features/` reflete exatamente o nome do discovery original (ex: `sdd/features/feat-XX-nome-discovery/`).
  - **Varredura Recursiva no CLI:** O comando CLI `doctor` foi refatorado para varrer recursivamente todas as subpastas em busca de arquivos de feature (`feat-`) e task (`task-`).
  - **Instruções e Prompts Atualizados:** Ajustados os prompts de sistema e chat (`/split-features`, `/nova-feature`, `/discovery`) nos 4 agentes suportados (Gemini, Claude, Copilot e OpenAI) para impor a nova regra física.

### Versão 1.6.1-beta.2 (Beta)

- **Fase 40 — Refinamento do Comando `init` e Relatórios**:
  - **Inteligência de Pastas no CLI:**
    - `forge-sdd init meu-projeto`: Cria a subpasta `./meu-projeto/` e inicializa o scaffolding dentro dela.
    - `forge-sdd init` (sem argumentos): Abre o formulário interativo e, ao final, cria uma subpasta com o nome digitado para o projeto, inicializando nela.
    - `forge-sdd init .`: Inicializa na pasta atual, detectando automaticamente o nome do projeto com base no nome do diretório físico.
  - **Relatórios de Fechamento:** Ao concluir `init` ou `update`, exibe um resumo visual consolidado das escolhas (Nome do projeto, Stack, BD, Idioma, Agentes e Telemetria).

### Versão 1.6.1-beta.1 (Beta)

- **Fase 39 — Comandos doctor e destroy no CLI**: 
  - Adicionado o comando `doctor` para diagnóstico completo de saúde dos componentes da metodologia, detecção de arquivos faltantes e progresso de features locais.
  - Adicionado o comando `destroy` para remover com segurança e de forma interativa a estrutura inteira e configurações de agentes do repositório (com suporte a `--dry-run` e `--yes`).
- **Ajustes de UI de Update**: O prompt de atualização de agentes agora exibe os agentes existentes desmarcados com o sufixo `(Já instalado)`, evitando sobrescritas acidentais e permitindo a seleção seletiva.
- **Purga de Arquivos Obsoletos**: O CLI apaga automaticamente arquivos de comandos descontinuados (como o `/install-skill`) para evitar incompatibilidades.

### Versão 1.6.1-beta.0 (Beta)

- **Fase 38 — Detecção Inteligente e Upgrade no CLI**: O CLI agora detecta automaticamente se o diretório alvo já possui a metodologia estruturada de SDD e redireciona para a interface de upgrade.
- **Consulta ao NPM Registry**: O CLI consulta de forma assíncrona as versões oficiais (`latest`) e de teste (`beta`) direto no repositório NPM para que o usuário selecione para qual deseja atualizar.

### Versão 1.6.0 (Estável)

- **Fase 36 — Documentação Estruturada**: Estruturação de diagramas conceituais Mermaid detalhados para visualização clara de fluxos e arquitetura.
- **Fase 35 — Remoção de Skill Incompleta**: Limpeza completa da funcionalidade de instalação de skill (`/install-skill`) em todos os prompts, agentes e documentação.
- **Fase 34 — Reestruturação de Prompts por Agente**: Migração dos prompts do Copilot e unificação dos prompts de sistema e habilidades customizadas específicas para Gemini, Claude e Copilot.
- **Fase 33 — Sincronização Automática da Wiki**: A documentação oficial (diretório docs/ e README) agora é enviada e atualizada automaticamente na wiki do repositório lp-forge-sdd a cada nova tag publicada.
- **Fase 32 — Geração de Release Notes**: Agora o agente documenta as entregas de cada feature de forma automática e em linguagem clara de produto ao encerrar as tarefas.
- **Fase 31 — Preservação de Domínio no Upgrade**: As atualizações do Forge-SDD ou do CLI protegem e preservam todas as suas especificações e arquivos de domínio do usuário.
- **Fase 30 — Correções de Fluxo de Branches**: Ajustado o fluxo de desenvolvimento automático do agente para garantir criação prévia de branches e integração robusta de PRs via `gh` CLI.

### Versões Anteriores (v1.5.3 para trás)
- **Fase 29 — Consolidação de Distribuição NPM**: Remoção do suporte ao Homebrew para focar a distribuição e execução do CLI inteiramente via `npx` simplificando pipelines.
- **Fase 28 — Guia de Fluxo de Desenvolvimento**: Lançamento do Guia de Fluxo do SDD (`sdd/FLOW.md`) detalhando boas práticas de splits de feature e discovery.
- **Fase 27 — Atualização de Node no CI**: Upgrade dos workflows do GitHub Actions para utilizar Node 24 para garantir o pipeline livre de depreciation warnings.
- **Fase 26 — Posicionamento Open Source**: Nova documentação e README estruturado que estabelece formalmente o forge-sdd como código aberto para controle de fluxos IA.
- **Fase 24 — Handoff Automático**: Encerramento automático de features com geração de commit semântico e criação direta de Pull Request via linha de comando.
- **Fase 23 — MCP e Habilidades por Agente**: Configurações de ferramentas MCP e habilidades específicas isoladas e customizadas para Gemini e Claude.
- **Fase 22 — Detecção Inteligente de Instalação**: Comando `/upgrade-sdd` detecta automaticamente se o CLI foi instalado via NPX ou Brew, instruindo a forma certa de atualizar.
- **Fase 21 — Instalação de Skills via URL**: Novo comando `/install-skill` para baixar e adaptar skills de IA diretamente de links públicos do GitHub para o seu projeto.
- **Fase 20 — Refinamento C4 Model**: Diagramas Mermaid C4 simplificados e padronizados para garantir perfeita compatibilidade de renderização visual.
- **Fase 19 — Integração de Arquitetura C4**: Capacidade de descrever arquiteturas complexas usando diagramas de fluxo e sequência C4 Model representados via Mermaid.
- **Fase 18 — Protocolo de Handoff**: Mecanismo de transição estruturado entre fases de desenvolvimento para prevenir perda de contexto e memória entre agentes.
- **Fase 17 — Onboarding Simplificado**: Guias rápidos ("Getting Started") reestruturados para facilitar o onboarding em repositórios novos e já existentes.
- **Fase 16 — Comando de Constituição Automática**: Comando `/constitution` para varredura do codebase do projeto, gerando regras e o stack tecnológico automaticamente.
- **Fase 15 — Guardrails de SDD**: Validações automáticas para escrita de especificações e features mantendo a qualidade técnica das fases.
- **Fase 14 — Comando /discovery**: Introdução do `/discovery` para criar automaticamente especificações de produto e critérios técnicos refinados a partir de breves ideias.
- **Fase 13 — Documentação Multi-Agente**: Atualização dos READMEs principais e do wrapper NPM para documentar e apoiar o uso do Gemini, Claude e Copilot.
- **Fase 12 — Deploy Automatizado da Landing Page**: Pipeline automatizado no Vercel e NPM que garante a publicação contínua da landing page e a atualização do pacote NPM.
- **Fase 11 — Estruturas Unificadas de Agentes**: Padronização dos diretórios de MCP, prompts e skills entre Gemini, Claude e Copilot.
- **Fase 10 — Landing Page Oficial**: Criação da landing page oficial em `site/` para visualização online do forge-sdd e seus comandos.
- **Fase 9 — Suporte Multi-Agente**: Possibilidade de escolher Claude, Gemini ou GitHub Copilot durante o scaffold inicial para gerar configurações focadas no agente escolhido.
- **Fase 8 — Execução via NPX**: Wrapper NPM que possibilita executar o CLI via `npx forge-sdd init` sem necessidade de ambiente local com Go pré-instalado.
- **Fase 7 — Compilação Multi-Plataforma**: Automação com GoReleaser para gerar binários e publicar em GitHub Releases a partir de tags Git.
- **Fase 6 — Teste de Conformidade (Self-test)**: Cobertura de testes automatizados com golden files para atestar que atualizações não quebram o layout do scaffolding.
- **Fase 5 — Arquivos de Configuração de Versão**: Geração e controle da versão via `.sdd-version` e configurações do projeto no arquivo `.sddrc`.
- **Fase 4 — Pré-visualização com Dry-run**: Flag `--dry-run` para listar quais arquivos seriam gerados no disco sem de fato gravá-los.
- **Fase 3 — Execução Silenciosa com Flags**: Flag `--yes` para execução não interativa e simplificada do CLI ideal para automações.
- **Fase 2 — Prompt Interativo Moderno**: Interface amigável baseada em TUI (usando a biblioteca `huh`) para coletar parâmetros durante o `init`.
- **Fase 1 — Templates Embutidos no Executável**: Uso de `embed.FS` para carregar todos os templates no binário compilado eliminando dependência externa de assets.
- **Fase 0 — Base Estrutural**: Estruturação inicial do projeto Go, CLI integrado via Cobra CLI e definição dos stubs internos.
