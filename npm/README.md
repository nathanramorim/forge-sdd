# @nathanramorim/forge-sdd

> CLI que scaffolda a estrutura **Forge-SDD** v1.4.3 em segundos. Suporte total para **GitHub Copilot, Claude e Gemini**.

🚀 **Landing Page:** [forge-sdd.vercel.app](https://forge-sdd.vercel.app)

## Uso rápido (npx)

```bash
npx @nathanramorim/forge-sdd@latest init
```

Apresenta um formulário interativo e cria toda a infraestrutura de memória de projeto, especificações, agentes e MCPs.

---

## Requisitos

- Node.js ≥ 18
- O binário Go é baixado automaticamente e cacheado em `~/.cache/forge-sdd/`.

---

## Opções CLI

```bash
# Modo interativo (padrão)
npx @nathanramorim/forge-sdd@latest init

# Gemini (Google) - não interativo
npx @nathanramorim/forge-sdd@latest init --yes --agent gemini

# Claude (Anthropic)
npx @nathanramorim/forge-sdd@latest init --yes --agent claude

# Preview sem criar arquivos
npx @nathanramorim/forge-sdd@latest init --yes --dry-run
```

---

## O que é gerado

Uma árvore de aproximadamente 40 arquivos baseada na **Metodologia Forge-SDD**, organizada para minimizar o consumo de tokens e maximizar a precisão da IA:

- `sdd/memory/`: Estado ativo e histórico.
- `sdd/spec/`: Especificação particionada (overview, stack, modules...).
- `sdd/features/`: Roadmap e tarefas executáveis.
- **Agentes:** Arquivos de instrução e comandos customizados para o agente escolhido.
- `.vscode/mcp.json`: Integração com MCP (Context7, Git).

---

## 🚀 Guia de Início Rápido

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

A Metodologia Forge-SDD v1.4.3

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
| `/c4-architecture` | Gera diagramas C4 Model em Mermaid. | Copilot, Claude, Gemini |

### ♊ Gemini (Google AI Studio / CLI)
Peça pelo nome do comando ou por uma frase de ação (ex: "rodar o status" ou "iniciar próxima tarefa"). O Gemini carrega automaticamente suas habilidades de Orquestrador, Builder, etc., localizadas em `.gemini/`.

### 🤖 GitHub Copilot (VS Code Chat)
Use comandos slash diretamente no chat (ex: `/status`). Configuração em `.github/chatmodes/` e `.github/prompts/`.

---

## Links

- [GitHub Repository](https://github.com/nathanramorim/forge-sdd)
- [Metodologia SDD](https://github.com/nathanramorim/forge-sdd#readme)
