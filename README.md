# forge-sdd

[![NPM Version](https://img.shields.io/npm/v/@nathanramorim/forge-sdd)](https://www.npmjs.com/package/@nathanramorim/forge-sdd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

CLI que scaffolda a estrutura **Forge-SDD** em qualquer projeto — pronta para uso com **GitHub Copilot, Claude e Gemini**.

🚀 **Landing Page:** [forge-sdd.vercel.app](https://forge-sdd.vercel.app)

## Por que o forge-sdd? (Open Source)

O **forge-sdd** nasceu para **mudar a dinâmica de desenvolvimento orientado a IA**. No cenário atual, com novas ideias, ferramentas e informações surgindo diariamente no ecossistema de inteligência artificial, manter a consistência e a velocidade do desenvolvimento pode ser um desafio. 

Este projeto foi criado para resolver esses problemas através de quatro pilares fundamentais:
- 🤖 **Controle de Fluxo com IA:** Fornece um framework estruturado de ciclos de desenvolvimento onde a IA atua de forma guiada e previsível (Orquestrador, Builder, Revisor).
- 🧠 **Fim da Replicação de Instruções:** Chega de copiar e colar regras de arquitetura, preferências de estilo ou restrições do projeto a cada nova sessão de chat. O forge-sdd centraliza a memória técnica do projeto em arquivos locais (`progress.md`, `constitution.md`).
- 🔄 **Padrões & Reutilização Incremental:** Garante que cada evolução do código siga padrões arquiteturais sólidos e que novas informações sejam agregadas de maneira incremental e reutilizável.
- 🎓 **Expertise Sênior Integrada:** Injeta no ciclo de vida do projeto processos e metodologias com a maturidade de engenheiros de software seniores (fases de Discovery, especificações claras e revisões de código automatizadas).

Como um projeto **Open Source** licenciado sob a licença MIT, o forge-sdd foi aberto para que a comunidade possa evoluir e adaptar o fluxo de scaffolding de SDDs para qualquer stack de desenvolvimento.

## O que ele gera

Ao rodar `forge-sdd init`, o CLI cria a árvore completa da **Metodologia SDD v1.5.2

```
sdd/                          → memória e especificação do projeto
  memory/
    progress.md               → estado ativo (leia primeiro a cada sessão)
    constitution.md           → regras imutáveis do projeto
    mcps.md                   → MCPs configurados
    progress-log.md           → histórico de sessões
  spec/
    overview.md, stack.md, modules.md, flows.md, decisions.md
  features/
    feat-00-foundation.md     → feature inicial
    index.md                  → mapa de features
  skills/index.md
  plan.md, README.md
  .sdd-version                → versão Forge-SDD usada
  .sddrc                      → config JSON do projeto
  .metrics/schema.json

# Agente Selecionado (Ex: Gemini)
.gemini/
  system_instructions.md
  skills/*.chatmode.md        → 6 papéis (orquestrador, builder, revisor…)
  prompts/*.prompt.md         → 7 comandos (status, proxima-feature…)
GEMINI.md                     → Instruções de contexto

.vscode/
  mcp.json                    → configuração dos MCPs (context7, git)
```

---

## Instalação

### npx (recomendado)
```bash
npx @nathanramorim/forge-sdd@latest init
```
O binário Go é baixado automaticamente e cacheado em `~/.cache/forge-sdd/`.

### Homebrew
```bash
brew install nathanramorim/forge-sdd/forge-sdd
```

---

## Uso

### Modo interativo (padrão)
```bash
cd meu-projeto
forge-sdd init
```
Apresenta um formulário no terminal para configurar:
- **Nome do projeto**
- **Stack principal** (go, node, python, rust, other)
- **Banco de dados** (postgres, sqlite, mongo, none)
- **Agente(s) de IA** — GitHub Copilot, Claude, Gemini (multi-select)
- **Telemetria local**
- **Idioma** (pt-BR ou en)

### Modo não-interativo (`--yes`)
```bash
# Gemini (Google)
forge-sdd init --yes --agent gemini --name meu-servico

# Claude (Anthropic)
forge-sdd init --yes --agent claude --name meu-servico

# GitHub Copilot
forge-sdd init --yes --agent copilot --name meu-servico

# Múltiplos agentes simultâneos
forge-sdd init --yes --agent copilot,claude,gemini
```

### Agentes suportados

Cada agente possui uma interface otimizada para a Metodologia SDD:

| Agente | Flag | Estrutura de Comandos | Como Usar |
|--------|------|-----------------------|-----------|
| **GitHub Copilot** | `copilot` | `.github/chatmodes/` & `.github/prompts/` | Use `/` no chat (ex: `/status`) |
| **Claude** | `claude` | `CLAUDE.md` & `.claude/commands/` | Mencione o comando (ex: `/revisar`) |
| **Gemini** | `gemini` | `GEMINI.md`, `.gemini/skills/` & `.gemini/prompts/` | Peça o comando ou anexe a skill |

---

## 🚀 Guia de Início Rápido

Dependendo do seu cenário, siga uma das trilhas lógicas abaixo para extrair o máximo do SDD:

### 🆕 Trilha para Novos Projetos (Do Zero)
1. `npx @nathanramorim/forge-sdd@latest init`
2. `/constitution` → Define arquitetura e regras base.
3. `/discovery "sua ideia"` → **Discovery de Produto:** Onde o agente define o "que" e o "para quem" será construído.
4. `/nova-feature` → Mapeia o roadmap de tarefas baseado no discovery.
5. `/proxima-feature` → Inicia a implementação.

### 🏗️ Trilha para Projetos Existentes (Adoção)
1. `npx @nathanramorim/forge-sdd@latest init --yes`
2. `/constitution` → O agente faz o **scan do seu codebase** e aprende suas regras.
3. `/status` → Sincroniza o estado atual do projeto.
4. `/nova-feature` → Mapeia a próxima evolução ou correção necessária.

---

## Como interagir com os Agentes

A Metodologia Forge-SDD v1.5.2

### 🛠️ Comandos Universais (Prompts)

| Comando | O que faz | Agente (Ativação) |
|---------|-----------|-------------------|
| `/status` | Diagnóstico rápido do progresso e fases. | Copilot, Claude, Gemini |
| `/proxima-feature` | Inicia a próxima tarefa (cria branch, delega). | Copilot, Claude, Gemini |
| `/nova-feature` | Specifier cria nova feature e atualiza índice. | Copilot, Claude, Gemini |
| `/revisar` | Revisor valida código e critério de conclusão. | Copilot, Claude, Gemini |
| `/doctor` | Check-up de integridade, MCPs e budgets. | Copilot, Claude, Gemini |
| `/archive` | Compacta `progress.md` movendo para o log. | Copilot, Claude, Gemini |
| `/upgrade-sdd` | Migra a estrutura para uma nova versão. | Copilot, Claude, Gemini |
| `/discovery` | Processo de Discovery (Produto + Engenharia Sênior). | Copilot, Claude, Gemini |
| `/constitution` | Analisa codebase e alinha arquitetura/regras no SDD. | Copilot, Claude, Gemini |
| `/c4-architecture` | Gera diagramas C4 Model (Contexto, Container, Seq) em Mermaid. | Copilot, Claude, Gemini |
| `/split-features` | Quebra um plano de discovery em múltiplas features independentes. | Copilot, Claude, Gemini |
| `/install-skill` | Importa e adapta uma skill a partir de uma URL do GitHub/Raw. | Copilot, Claude, Gemini |

---

### 🤖 GitHub Copilot (VS Code Chat)
- **Interface:** Use comandos slash diretamente no chat (ex: `/status`).
- **Modos:** Você pode pedir explicitamente: *"Ative o chatmode orquestrador"* para gerenciar a sessão.
- **Configuração:** Localizada em `.github/chatmodes/` e `.github/prompts/`.

### ♊ Gemini (Google AI Studio / CLI / Code Assist)
- **Interface:** Peça pelo nome do comando ou por uma frase de ação (ex: "rodar o status" ou "desenhar arquitetura").
- **Skills:** O Gemini carrega automaticamente suas habilidades de Orquestrador, Builder, etc.
- **Configuração:** Localizada em `.gemini/skills/` e `.gemini/prompts/`.

### 📝 Claude (Claude.ai / Desktop / Dev)
- **Interface:** Mencione o comando no chat (ex: `/revisar`).
- **Comandos:** Segue as instruções em `CLAUDE.md`.
- **Configuração:** Localizada em `.claude/commands/`.

---

## Primeiros passos após o `init`

1. Abra o projeto no **VS Code**
2. Aceite as extensões recomendadas (Copilot, MCP)
3. Leia `sdd/memory/progress.md` — é o ponto de entrada de cada sessão
4. Use o comando `/status` no seu agente para ver o estado inicial do projeto.

---

## Contribuição

Veja [CONTRIBUTING.md](CONTRIBUTING.md) para detalhes sobre como contribuir para o projeto.

## Licença

Distribuído sob a licença MIT. Veja `LICENSE` para mais informações.
