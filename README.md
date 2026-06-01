# forge-sdd

CLI que scaffolda a estrutura **Forge-SDD** em qualquer projeto — pronta para uso com GitHub Copilot.

## O que ele gera

Ao rodar `forge-sdd init`, o CLI cria **32 arquivos** divididos em três áreas:

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

.github/
  copilot-instructions.md     → instruções globais para o Copilot
  chatmodes/                  → 6 modos de agente (orquestrador, builder, revisor…)
  prompts/                    → 7 prompts reutilizáveis (status, proxima-feature…)

.vscode/
  mcp.json                    → configuração dos MCPs (context7, git)
```

---

## Instalação

### npx (sem instalação)
```bash
npx forge-sdd@latest init
```
Funciona em qualquer máquina com Node.js ≥ 18. O binário Go é baixado automaticamente e cacheado em `~/.cache/forge-sdd/`.

### Homebrew
```bash
brew install nathanramorim/forge-sdd/forge-sdd
```

### Download direto
Baixe o binário para seu sistema em [Releases](../../releases):

| Plataforma | Arquivo |
|------------|---------|
| macOS Apple Silicon | `forge-sdd_darwin_arm64.tar.gz` |
| macOS Intel | `forge-sdd_darwin_amd64.tar.gz` |
| Linux x86-64 | `forge-sdd_linux_amd64.tar.gz` |
| Windows x86-64 | `forge-sdd_windows_amd64.zip` |

```bash
tar xzf forge-sdd_darwin_arm64.tar.gz
sudo mv forge-sdd /usr/local/bin/
```

### Compilar do código-fonte
```bash
git clone https://github.com/forge-sdd/cli
cd cli
go build -ldflags "-X main.version=1.1.0" -o forge-sdd ./cmd/forge-sdd
sudo mv forge-sdd /usr/local/bin/
```

---

## Uso

### Modo interativo (padrão)
```bash
cd meu-projeto
forge-sdd init
```
Apresenta um formulário no terminal com 6 campos:
- **Nome do projeto**
- **Stack principal** — go, node, python, rust, other
- **Banco de dados** — postgres, sqlite, mongo, none
- **Agente(s) de IA** — GitHub Copilot, Claude, Gemini (multi-select)
- **Telemetria local** — sim/não
- **Idioma** — pt-BR ou en

### Modo não-interativo (`--yes`)
```bash
# GitHub Copilot (default)
forge-sdd init --yes --name meu-servico --stack go

# Claude
forge-sdd init --yes --agent claude --name meu-servico

# Gemini
forge-sdd init --yes --agent gemini --name meu-servico

# Múltiplos agentes
forge-sdd init --yes --agent copilot,claude --name meu-servico
```
Pula todos os prompts e usa as flags passadas (o resto usa defaults).

### Flags disponíveis
| Flag | Descrição | Default |
|------|-----------|---------|
| `--yes` | Pula prompts interativos | `false` |
| `--name` | Nome do projeto | `meu-projeto` |
| `--stack` | Stack principal | `go` |
| `--db` | Banco de dados | `none` |
| `--lang` | Idioma dos templates | `pt-BR` |
| `--agent` | Agente(s) de IA: `copilot`, `claude`, `gemini` (csv) | `copilot` |
| `--version` | Versão Forge-SDD | `1.1.0` |
| `--no-telemetry` | Desabilita telemetria | `false` |
| `--dry-run` | Lista arquivos sem criar | `false` |

### Agentes suportados
| Agente | Flag | Arquivos gerados |
|--------|------|-----------------|
| GitHub Copilot | `copilot` | `.github/copilot-instructions.md`, chatmodes, prompts |
| Claude | `claude` | `CLAUDE.md`, `.claude/commands/*.md` |
| Gemini | `gemini` | `GEMINI.md`, `.gemini/system_instructions.md` |

### Dry-run (visualizar sem criar)
```bash
forge-sdd init --dry-run --yes
# imprime [DRY] /caminho/do/arquivo para cada um dos 32 arquivos
# nenhum arquivo é criado em disco
```

### Ver versão
```bash
forge-sdd version
# → 1.1.0
```

---

## Primeiros passos após o `init`

1. Abra o projeto no **VS Code**
2. Aceite as extensões recomendadas (Copilot, MCP)
3. Leia `sdd/memory/progress.md` — é o ponto de entrada de cada sessão
4. Configure o Copilot com o chatmode `orquestrador` para começar a primeira feature
