# @nathanramorim/forge-sdd

[![NPM Version](https://img.shields.io/npm/v/@nathanramorim/forge-sdd)](https://www.npmjs.com/package/@nathanramorim/forge-sdd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> CLI open source que instala em qualquer projeto a **Metodologia Forge-SDD** v1.7.1-beta.5 — um framework de desenvolvimento guiado por IA que elimina a repetição de instruções, garante padrões arquiteturais e traz a expertise de um engenheiro sênior para o seu fluxo diário.

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

## 📢 Novidades da Versão (v1.7.1-beta.5)

Esta versão de testes beta traz correções e alinhamento no fluxo de descobertas (discovery):

*   **Plano de Discovery Padronizado:** O prompt de `/discovery` de todos os agentes (Copilot, Claude e Gemini) foi atualizado e alinhado para gerar obrigatoriamente o arquivo `plan-XX-*.md` na pasta `sdd/discovery/` contendo o roadmap e sugestão de quebra de tarefas/features.
*   **Agrupamento Opcional de Releases Beta:** A política de releases beta (Regra 11 da Constituição) foi atualizada para explicitar que a publicação e geração de tags beta pode, opcionalmente, agrupar e acumular múltiplos fixes ou features antes de lançar novas tags ou bumps de versão, otimizando o fluxo.
*   **Diagnóstico de Nome Padrão:** O comando `doctor` e os prompts `/doctor` dos agentes passam a verificar se o projeto ainda utiliza o nome genérico `"meu-projeto"`, listando as correções e os cabeçalhos de agentes recomendados a serem renomeados.
*   **Métricas Granulares e Robustas:** Implementado mapeamento e telemetria incondicional para caminhos aninhados de subpastas de features e discoveries.
*   **Prompt do Orquestrador Refatorado:** Instruções do Orquestrador atualizadas para impor os novos guardrails de close em todos os agentes suportados.

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
