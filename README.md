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

## ✨ Forge-SDD Slim (v2.0.0-beta) — mais simples, e os agentes agora aprendem com os próprios erros

A metodologia continua completa — só ficou mais fácil de confiar nela.

*   **Telemetria que não falha mais em silêncio:** a gravação de métricas de sessão deixou de depender de um passo tardio de um prompt longo — agora é um comando determinístico do próprio binário, disparado em todos os pontos onde uma sessão pode terminar.
*   **Os agentes aprendem com os próprios fixes:** novo `sdd/memory/lessons.md` registra automaticamente padrões de erro já corrigidos e é consultado por Builder/Revisor antes de implementar ou revisar.
*   **MCPs e VCS configuráveis por projeto:** declare na Constituição se o seu projeto usa GitHub, Azure DevOps ou nenhum VCS automatizado, e quais MCPs realmente respondem — sem mais suposições incondicionais.
*   **Um fluxo, uma fonte da verdade:** o pipeline por feature agora vive em um único lugar (`sdd/FLOW.md`), não em três descrições que podiam divergir.
*   **Menos duplicação nos bastidores:** a lógica de nomenclatura, antes copiada em ~14 lugares, agora vive em um único arquivo — mesma capacidade, menos superfície para manter.

📢 [Ver todas as entregas desta versão](https://github.com/nathanramorim/forge-sdd/blob/main/sdd/releases/history.md)

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

- [Landing Page](https://forge-sdd.vercel.app)
- [Metodologia SDD](docs/metodologia-sdd.md)
- [Telemetria e Métricas](docs/telemetria.md)
- [Guia de Contribuição](CONTRIBUTING.md)
- [Licença MIT](LICENSE)
