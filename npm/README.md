# @nathanramorim/forge-sdd

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

Você pode rodar diretamente via `npx` ou instalar globalmente para ter o comando simplificado `forge` à disposição:

### Via Instalação Global (Recomendado)
```bash
# Instale globalmente no seu sistema
npm install -g @nathanramorim/forge-sdd

# Inicialize de forma simples em qualquer projeto
forge init
```

### Via Execução Direta (npx)
```bash
# Inicialização interativa (escolha os agentes no menu)
npx @nathanramorim/forge-sdd@latest init

# Inicialização rápida no diretório atual
npx @nathanramorim/forge-sdd@latest init . --agent copilot,gemini --name meu-projeto
```

## ✨ Agent Rules e Branch por Feature (v2.2.0-beta) — uma fonte, três agentes, sem duplicar

*   **`.agent/rules/` — regras de domínio compartilhadas:** declare design system, arquitetura, acessibilidade e outras convenções do seu projeto em `.agent/rules/*.md`, uma pasta neutra fora de `.claude/`/`.gemini/`/`.github/`. Qualquer agente configurado consulta o mesmo arquivo — nada para copiar entre eles.
*   **`.agent/commands/` — comandos com corpo único, não mais triplicado:** os 13 comandos SDD (`/discovery`, `/proxima-feature`, etc.) agora têm um corpo de instrução canônico em `.agent/commands/`. `.claude/commands/`, `.gemini/prompts/` e `.github/prompts/` continuam existindo (é assim que cada ferramenta descobre o comando), mas viram adaptadores finos que apontam para o corpo compartilhado — uma mudança de comportamento passa a ser editada uma vez, não três.
*   **`forge-sdd update` migra sem perder nada:** projetos existentes ganham `.agent/rules/` e `.agent/commands/` automaticamente ao atualizar, sem tocar em `sdd/features/`, `sdd/discovery/`, `sdd/fix/*` ou `progress.md`. Regras já criadas em `.agent/rules/` nunca são sobrescritas.
*   **Branch única por feature quebrada em subpastas:** ao trabalhar em uma feature dividida em subtarefas (`sdd/features/feat-XX-nome/`), os agentes (`/nova-feature`, `/proxima-feature`, `/novo-fix`) agora tratam a pasta inteira como uma única branch — em vez de uma por subtarefa.
*   **Pergunta obrigatória de branch de partida e retomada:** antes de criar uma branch, os agentes perguntam de onde partir (default `main`) e verificam se já existe uma branch da mesma feature/fix para continuar, em vez de recriar do zero.

📢 [Ver todas as entregas desta versão](https://github.com/nathanramorim/forge-sdd/blob/main/sdd/releases/history.md)

---

## 📢 Novidades da Versão Anterior (v2.0.0-beta — Forge-SDD Slim)

A metodologia continua completa — só ficou mais fácil de confiar nela.

*   **Telemetria que não falha mais em silêncio:** a gravação de métricas de sessão deixou de depender de um passo tardio de um prompt longo — agora é um comando determinístico do próprio binário, disparado em todos os pontos onde uma sessão pode terminar.
*   **Os agentes aprendem com os próprios fixes:** novo `sdd/memory/lessons.md` registra automaticamente padrões de erro já corrigidos e é consultado por Builder/Revisor antes de implementar ou revisar.
*   **MCPs e VCS configuráveis por projeto:** declare na Constituição se o seu projeto usa GitHub, Azure DevOps ou nenhum VCS automatizado, e quais MCPs realmente respondem — sem mais suposições incondicionais.
*   **Um fluxo, uma fonte da verdade:** o pipeline por feature agora vive em um único lugar (`sdd/FLOW.md`), não em três descrições que podiam divergir.
*   **Menos duplicação nos bastidores:** a lógica de nomenclatura, antes copiada em ~14 lugares, agora vive em um único arquivo — mesma capacidade, menos superfície para manter.

---

## 📢 Novidades da Versão Anterior (v1.9.4)

*   **Telemetria funcionando para todos os agentes:** Antes, a gravação de métricas de uso ao final de uma sessão só acontecia de fato ao usar o agente Gemini — Claude e Copilot silenciosamente deixavam de registrar telemetria mesmo com a opção ativada no projeto. Agora os três agentes gravam a telemetria corretamente ao concluir uma feature ou correção.

---

## 📢 Novidades da Versão Anterior (v1.9.3)

Release apenas de documentação: sincroniza a seção "Novidades" com o estado real do projeto, sem alteração de código/comportamento.

---

## 📢 Novidades da Versão Anterior (v1.9.1)

Esta release estável consolidou todo o ciclo de betas desde a v1.7.0:

*   **Convenção de Nomenclatura Configurável:** Escolha entre nomenclatura `sequencial`, `hash` ou `workitem` no `init`, com auto-healing e detecção de deriva de convenção no `doctor` para projetos existentes.
*   **Onboarding Mais Simples:** `init` agora imprime um cheat-sheet dos comandos disponíveis, `/status` sempre sugere o próximo passo, e um novo comando `/tutorial` guia um ciclo SDD completo fictício para quem está começando.
*   **Modo Iniciante:** `/constitution` pode gerar explicações em linguagem simplificada, com exemplos no lugar de jargão técnico.
*   **Comando `forge-sdd autopilot`:** Ativa o modo autopilot somente após um número mínimo de ciclos completos registrados em telemetria, com bypass consciente via flag.
*   **Telemetria Mais Confiável e Diagnósticos Adicionais no `doctor`:** Estimativa real de tokens de entrada/saída, ativação/desativação dinâmica baseada no `.sddrc`, e novas checagens de nome padrão de projeto e caminho de métricas.

---

## 📢 Novidades da Versão Anterior (v1.9.0-beta)

Esta versão trouxe o atalho global de execução `forge`, uma interface de onboarding pós-instalação e melhorias focadas em reduzir a curva de aprendizado do Forge-SDD:

*   **Atalho Global (`forge`) e Onboarding Pós-Instalação:** Agora você pode acionar todos os comandos da CLI simplesmente usando `forge` (ex: `forge init`, `forge doctor`). Ao instalar o pacote globalmente, uma tela de boas-vindas interativa e instrutiva é exibida com o guia dos comandos.
*   **Cheat-Sheet de Comandos:** O `forge init` agora imprime a lista completa dos comandos SDD disponíveis para os agentes escolhidos ao final da inicialização.
*   **`/status` Prescritivo:** O comando `/status` (Copilot, Claude e Gemini) agora sempre sugere o próximo comando a ser executado, com base no estado real do progresso do projeto.
*   **Diagnóstico de Deriva de Nomenclatura:** O comando `doctor` passa a detectar quando um projeto mistura a nomenclatura sequencial (`feat-NN`) com a nomenclatura por hash (`feat-xxxx`), alertando o usuário sobre a inconsistência.
*   **Comando `autopilot`:** Novo comando CLI que ativa o modo autopilot somente após um número mínimo de ciclos completos registrados em telemetria, com bypass consciente disponível via flag.

---

## Links Úteis

- [GitHub Repository](https://github.com/nathanramorim/forge-sdd)
- [Landing Page](https://forge-sdd.vercel.app)
- [Metodologia SDD](https://github.com/nathanramorim/forge-sdd/blob/main/docs/metodologia-sdd.md)
- [Telemetria e Métricas](https://github.com/nathanramorim/forge-sdd/blob/main/docs/telemetria.md)
- [Licença MIT](https://github.com/nathanramorim/forge-sdd/blob/main/LICENSE)
