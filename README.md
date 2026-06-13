# forge-sdd

[![NPM Version](https://img.shields.io/npm/v/@nathanramorim/forge-sdd)](https://www.npmjs.com/package/@nathanramorim/forge-sdd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

CLI que scaffolda a estrutura **Forge-SDD** em qualquer projeto — pronta para uso com **GitHub Copilot, Claude e Gemini**.

🚀 **Landing Page:** [forge-sdd.vercel.app](https://forge-sdd.vercel.app)

## O que ele gera

Ao rodar `forge-sdd init`, o CLI cria a árvore completa da **Metodologia SDD v1.3.0** (aprox. 40 arquivos):

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

## Como interagir com os Agentes

Após o `init`, você terá acesso a comandos poderosos que automatizam o fluxo de desenvolvimento.

### 🤖 GitHub Copilot (VS Code Chat)
O Copilot utiliza **Chatmodes** para definir o comportamento e **Prompts** para comandos rápidos.
- **Modo Orquestrador:** "Ative o chatmode orquestrador para iniciar a sessão."
- **Comandos Slash:**
  - `/status` — Diagnóstico do projeto e progresso atual.
  - `/proxima-feature` — Inicia automaticamente a próxima tarefa do `index.md`.
  - `/revisar` — Chama o Revisor para validar o código contra o critério de conclusão.
  - `/nova-feature "descrição"` — O Specifier cria os arquivos e atualiza o mapa.

### ♊ Gemini (Google AI Studio / CLI)
O Gemini utiliza **Skills** (carregadas automaticamente via `GEMINI.md`) e **Prompts** de referência.
- **Interação:** Basta pedir ao Gemini para executar uma das tarefas definidas em `.gemini/prompts/`.
- **Comandos Disponíveis:**
  - `status.prompt.md` — Visão geral das fases e bloqueios.
  - `proxima-feature.prompt.md` — Orquestração da próxima tarefa.
  - `doctor.prompt.md` — Verifica integridade dos arquivos e budgets de tokens.
  - `archive.prompt.md` — Limpa o `progress.md` e move histórico para o log.

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
