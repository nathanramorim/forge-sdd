# feat/multi-agent

**Branch:** `feat/multi-agent`
**Fase:** 9
**Depende de:** `feat/npx` (mergeada)
**Status:** `todo`

## Objetivo
Suportar múltiplos agentes de IA no scaffolding: além do **GitHub Copilot** (atual), permitir escolher **Claude (Anthropic)** e **Gemini (Google)** — via formulário interativo ou flag `--agent`.

Cada agente gera arquivos específicos ao seu ecossistema (chatmodes, prompts, instruções) no lugar dos arquivos Copilot-only atuais.

## Critério de conclusão
```bash
# interativo — exibe seleção de agentes
forge-sdd init

# flag direta
forge-sdd init --agent copilot
forge-sdd init --agent claude
forge-sdd init --agent gemini

# múltiplos agentes
forge-sdd init --agent copilot,claude

# dry-run
forge-sdd init --agent claude --yes --dry-run
# → arquivos CLAUDE.md, .claude/commands/*.md no preview
```

## Estratégia técnica

### Novo campo em `Config`
```go
// config.go
Agents []string // ex: ["copilot"], ["claude"], ["gemini"], ["copilot","claude"]
```

### Flag e survey
- Flag: `--agent <agent1,agent2>` (default: `copilot`)
- Survey: `huh.NewMultiSelect` com opções Copilot, Claude, Gemini

### Estrutura de templates por agente

**copilot** (atual — sem mudança):
```
.github/
  copilot-instructions.md
  chatmodes/*.chatmode.md
  prompts/*.prompt.md
.vscode/mcp.json
```

**claude**:
```
CLAUDE.md                        → instruções globais para o Claude
.claude/
  commands/
    nova-feature.md              → equivalente a nova-feature.prompt.md
    proxima-feature.md
    status.md
    revisar.md
```

**gemini**:
```
GEMINI.md                        → instruções globais para o Gemini
.gemini/
  system_instructions.md         → contexto do projeto
```

### Templates
Cada agente tem seu próprio subdiretório em `internal/scaffold/templates/agents/`:
```
agents/
  copilot/           → templates já existentes (mover de .github/)
  claude/
    CLAUDE.md.tmpl
    .claude/commands/nova-feature.md.tmpl
    .claude/commands/proxima-feature.md.tmpl
    .claude/commands/status.md.tmpl
    .claude/commands/revisar.md.tmpl
  gemini/
    GEMINI.md.tmpl
    .gemini/system_instructions.md.tmpl
```

### Scaffold
`scaffold.Run()` itera sobre `cfg.Agents` e renderiza o diretório de cada agente selecionado.

## Tarefas
- [ ] **09-1** Adicionar campo `Agents []string` em `config.go` + default `["copilot"]`
- [ ] **09-2** Adicionar flag `--agent` em `cmd/forge-sdd/main.go` (aceita csv)
- [ ] **09-3** Adicionar `huh.NewMultiSelect` no survey com opções Copilot / Claude / Gemini
- [ ] **09-4** Criar templates `agents/claude/` (CLAUDE.md + 4 commands)
- [ ] **09-5** Criar templates `agents/gemini/` (GEMINI.md + system_instructions.md)
- [ ] **09-6** Refatorar `scaffold.go` para iterar sobre agentes e renderizar por pasta
- [ ] **09-7** Atualizar testes (`scaffold_test.go`) para cobrir `--agent claude` e `--agent gemini`
- [ ] **09-8** Atualizar `--help` e README com a nova flag

## Arquivos gerados / alterados
```
internal/config/config.go              (campo Agents + FromFlags)
internal/survey/survey.go              (MultiSelect agentes)
cmd/forge-sdd/main.go                  (flag --agent)
internal/scaffold/scaffold.go          (iterar agentes)
internal/scaffold/scaffold_test.go     (novos casos)
internal/scaffold/templates/agents/
  claude/
    CLAUDE.md.tmpl
    .claude/commands/nova-feature.md.tmpl
    .claude/commands/proxima-feature.md.tmpl
    .claude/commands/status.md.tmpl
    .claude/commands/revisar.md.tmpl
  gemini/
    GEMINI.md.tmpl
    .gemini/system_instructions.md.tmpl
README.md                              (seção --agent)
```

## Notas de compatibilidade
- Projetos existentes (Copilot) não são afetados — `default: copilot` preserva comportamento atual
- Múltiplos agentes podem coexistir no mesmo projeto (pastas separadas, sem conflito)

## Skills relevantes
(consultar `skills/index.md`)
