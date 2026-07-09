# Release Notes — forge-sdd

Este arquivo registra o histórico de entregas de produto deste projeto.

## Entregas

### Versão 1.7.1-beta.1 (Beta)

Esta release beta traz a flexibilização do padrão de nomenclatura de especificações no SDD, permitindo o desenvolvimento assíncrono e paralelo sem colisões.

*   **Identificação com Workitems:** Os prompts de IA (`/nova-feature` e `/discovery`) agora solicitam ativamente o ID do Workitem de referência (como Jira ou ClickUp) para usar no prefixo do nome de arquivos e branches.
*   **Geração de Hash Hexadecimal:** Caso o ID do Workitem de referência seja omitido, o agente gera um hash de 4 casas (ex: `3ec4`) em substituição à numeração sequencial rígida (`01`, `02`, etc.), eliminando gargalos de concorrência.

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
