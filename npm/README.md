# @nathanramorim/forge-sdd

[![NPM Version](https://img.shields.io/npm/v/@nathanramorim/forge-sdd)](https://www.npmjs.com/package/@nathanramorim/forge-sdd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> CLI open source que instala em qualquer projeto a **Metodologia Forge-SDD** v1.8.1-beta — um framework de desenvolvimento guiado por IA que elimina a repetição de instruções, garante padrões arquiteturais e traz a expertise de um engenheiro sênior para o seu fluxo diário.

🚀 **Landing Page Oficial:** [forge-sdd.vercel.app](https://forge-sdd.vercel.app)
📖 **Wiki & Documentação Completa:** [Wiki do Forge-SDD](https://forge-sdd.vercel.app/wiki#introducao)

---

## ⚡ O que é o forge-sdd?

O **forge-sdd** nasceu para **mudar a dinâmica de desenvolvimento orientado a IA**. Ele resolve o desafio de manter a consistência, qualidade e velocidade do desenvolvimento através de agentes de IA locais especializados (Orquestrador, Builder, Revisor, etc.) operando sob a metodologia SDD (Software Design Doc).

Para conhecer todos os comandos disponíveis (`/status`, `/discovery`, `/nova-feature`, etc.), guias rápidos de início do zero e adoção em projetos existentes, acesse a nossa **[Landing Page Oficial](https://forge-sdd.vercel.app)** ou a **[Wiki do Projeto](https://forge-sdd.vercel.app/wiki#introducao)**.

---

## 💻 Instalação & Inicialização Rápida

Inicialize a estrutura em segundos. O instalador gerará as pastas de memória, especificações e configurações específicas para a sua IA preferida (Copilot, Gemini ou Claude):

```bash
# Inicialização interativa (escolha os agentes no menu)
npx @nathanramorim/forge-sdd@latest init

# Inicialização rápida no diretório atual
npx @nathanramorim/forge-sdd@latest init . --agent copilot,gemini --name meu-projeto
```

## 📢 Novidades da Versão (v1.8.1-beta)

Esta versão de testes beta traz a proteção de customizações locais durante upgrades e a automação contínua de sessões:

*   **Backup Automático no Upgrade:** Toda vez que você rodar um upgrade ou update via CLI, um backup físico de arquivos de configuração e prompts de agentes customizados no seu projeto (como `.github/copilot-instructions.md`, `CLAUDE.md`, prompts e skills) será salvo automaticamente em `sdd/backups/upgrade-<timestamp>/` antes de qualquer sobrescrita, eliminando o risco de perda de lógica do projeto.
*   **Modo Piloto Automático (Auto-Pilot):** Os prompts e chatmodes do Orquestrador foram atualizados para ler o arquivo flag `.sdd-auto-pilot` e ignorar a confirmação humana do `PLAN`. O CLI agora também scaffolda o script daemon padrão `sdd/scripts/sdd-daemon.sh` e o guia técnico de automação `sdd/docs/autonomia-autopilot.md`.
*   **Convenção de Versionamento Simplificada:** Tags do canal de homologação passam a usar apenas o sufixo simples `-beta` (ex: `1.8.1-beta`), permitindo o acúmulo de múltiplos fixes ou features antes de lançar a estável oficial `1.8.0` correspondente (Regra 15).
*   **Métricas Granulares e Robustas:** Mapeamento incondicional de telemetria mesmo para sessões parciais, abortadas ou com timeouts.
*   **Diagnóstico de Nome Padrão:** O comando `doctor` alerta caso o projeto use o nome padrão `"meu-projeto"` no `.sddrc` ou nos cabeçalhos dos arquivos de agentes.organização.

---

## 📢 Novidades da Versão Anterior (v1.6.1-beta.0)

A versão beta trouxe detecção inteligente de projetos existentes e upgrade aprimorado:

- **Fase 38 — Detecção Inteligente e Upgrade no CLI**: O CLI agora detecta automaticamente se o diretório já possui a metodologia estruturada e redireciona para a interface de upgrade.
- **Integração NPM Registry**: O CLI consulta dinamicamente as versões oficiais (`latest`) e de teste (`beta`) publicadas no NPM Registry para que o usuário escolha para qual deseja atualizar.

---

## 📢 Novidades da Versão Anterior (v1.6.0)

A versão estável v1.6.0 trouxe grandes evoluções de arquitetura e consistência no desenvolvimento:

- **Fase 36 — Documentação Estruturada**: Estruturação de diagramas conceituais Mermaid detalhados para visualização clara de fluxos e arquitetura.
- **Fase 35 — Remoção de Skill Incompleta**: Limpeza completa da funcionalidade de instalação de skill (`/install-skill`) em todos os prompts, agentes e documentação.
- **Fase 34 — Reestruturação de Prompts por Agente**: Migração dos prompts do Copilot e unificação dos prompts de sistema e habilidades customizadas específicas para Gemini, Claude e Copilot.
- **Fase 33 — Sincronização Automática da Wiki**: A documentação oficial (diretório docs/ e README) agora é enviada e atualizada automaticamente na wiki do repositório lp-forge-sdd a cada nova tag publicada.
- **Fase 32 — Geração de Release Notes**: Agora o agente documenta as entregas de cada feature de forma automática e em linguagem clara de produto ao encerrar as tarefas.
- **Fase 31 — Preservação de Domínio no Upgrade**: As atualizações do Forge-SDD ou do CLI protegem e preservam todas as suas especificações e arquivos de domínio do usuário.
- **Fase 30 — Correções de Fluxo de Branches**: Ajustado o fluxo de desenvolvimento automático do agente para garantir criação prévia de branches e integração robusta de PRs via `gh` CLI.

---

## Links Úteis

- [GitHub Repository](https://github.com/nathanramorim/forge-sdd)
- [Landing Page](https://forge-sdd.vercel.app)
- [Metodologia SDD](https://github.com/nathanramorim/forge-sdd/blob/main/docs/metodologia-sdd.md)
- [Telemetria e Métricas](https://github.com/nathanramorim/forge-sdd/blob/main/docs/telemetria.md)
- [Licença MIT](https://github.com/nathanramorim/forge-sdd/blob/main/LICENSE)
