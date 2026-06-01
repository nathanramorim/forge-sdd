# forge-sdd

> CLI que scaffolda a estrutura **Forge-SDD** em qualquer projeto — pronta para uso com GitHub Copilot.

## Uso rápido

```bash
npx forge-sdd@latest init
```

Apresenta um formulário interativo e cria **32 arquivos** com memória de projeto, spec, chatmodes, prompts e configuração de MCPs.

---

## Requisitos

- Node.js ≥ 18
- Acesso à internet (primeiro uso baixa o binário ~5 MB)

O binário Go é baixado automaticamente do GitHub Releases, validado por SHA256 e cacheado em `~/.cache/forge-sdd/`. Execuções seguintes são instantâneas.

---

## Opções

```bash
# Modo interativo (padrão)
npx forge-sdd@latest init

# Pular formulário, usar valores padrão
npx forge-sdd@latest init --yes

# Especificar diretório destino
npx forge-sdd@latest init /caminho/do/projeto

# Preview sem criar arquivos
npx forge-sdd@latest init --yes --dry-run

# Ver versão
npx forge-sdd@latest version
```

---

## O que é gerado

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
    feat-00-foundation.md
    index.md
  skills/index.md
  plan.md, README.md

.github/
  copilot-instructions.md     → instruções globais para o Copilot
  chatmodes/                  → 6 modos de agente (orquestrador, builder, revisor…)
  prompts/                    → 7 prompts reutilizáveis

.vscode/
  mcp.json                    → configuração dos MCPs (context7, git)
```

---

## Instalação permanente (opcional)

Se preferir ter o comando disponível globalmente sem `npx`:

```bash
npm install -g forge-sdd
forge-sdd init
```

Ou via Homebrew (macOS/Linux):

```bash
brew install nathanramorim/forge-sdd/forge-sdd
```

---

## Links

- [Repositório do projeto](https://github.com/nathanramorim/homebrew-forge-sdd)
- [Releases / Changelog](https://github.com/nathanramorim/homebrew-forge-sdd/releases)
- [Metodologia Forge-SDD](https://github.com/nathanramorim/homebrew-forge-sdd#readme)
