# forge-sdd

[![NPM Version](https://img.shields.io/npm/v/@nathanramorim/forge-sdd)](https://www.npmjs.com/package/@nathanramorim/forge-sdd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> CLI open source que instala em qualquer projeto a **Metodologia Forge-SDD** — um framework de desenvolvimento guiado por IA que elimina a repetição de instruções, garante padrões arquiteturais e traz a expertise de um engenheiro sênior para o seu fluxo diário.

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

## 📢 Novidades da Versão (v1.9.0-beta)

Esta versão de testes beta é focada em reduzir a curva de aprendizado do Forge-SDD para quem está começando, na era dos agentes de IA autônomos:

*   **Cheat-Sheet de Comandos:** O `forge-sdd init` agora imprime a lista completa dos comandos SDD disponíveis para os agentes escolhidos ao final da inicialização.
*   **`/status` Prescritivo:** O comando `/status` (Copilot, Claude e Gemini) agora sempre sugere o próximo comando a ser executado, com base no estado real do progresso do projeto.
*   **Diagnóstico de Deriva de Nomenclatura:** O comando `forge-sdd doctor` passa a detectar quando um projeto mistura a nomenclatura sequencial (`feat-NN`) com a nomenclatura por hash (`feat-xxxx`), alertando o usuário sobre a inconsistência.
*   **Onboarding Guiado:** Novo comando `/tutorial` (Copilot, Claude e Gemini) guia o usuário por um ciclo SDD completo e fictício, isolado dos dados reais do projeto.
*   **Comando `forge-sdd autopilot`:** Novo comando CLI que ativa o modo autopilot somente após um número mínimo de ciclos completos registrados em telemetria, com bypass consciente disponível via flag.
*   **Modo Iniciante:** O comando `/constitution` (Copilot, Claude e Gemini) passa a perguntar, opcionalmente, se o usuário prefere explicações em linguagem simplificada, com exemplos concretos no lugar de jargão técnico.

---

## 📢 Novidades da Versão Anterior (v1.7.1-beta.5)

Esta versão de testes beta trouxe correções e alinhamento no fluxo de descobertas (discovery):

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

## Links Úteis

- [Landing Page](https://forge-sdd.vercel.app)
- [Metodologia SDD](docs/metodologia-sdd.md)
- [Telemetria e Métricas](docs/telemetria.md)
- [Guia de Contribuição](CONTRIBUTING.md)
- [Licença MIT](LICENSE)
