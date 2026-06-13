# @nathanramorim/forge-sdd

> CLI que scaffolda a estrutura **Forge-SDD** v1.1.0 em segundos. Suporte total para **GitHub Copilot, Claude e Gemini**.

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

## Links

- [GitHub Repository](https://github.com/nathanramorim/forge-sdd)
- [Metodologia SDD](https://github.com/nathanramorim/forge-sdd#readme)
