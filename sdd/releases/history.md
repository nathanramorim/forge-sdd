# Release Notes — forge-sdd

Este arquivo registra o histórico de entregas de produto deste projeto.

## Entregas

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
