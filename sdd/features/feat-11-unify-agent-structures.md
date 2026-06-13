# feat/unify-agent-structures

**Branch:** `feat/unify-agent-structures`
**Fase:** 11
**Depende de:** `feat-09-multi-agent`
**Status:** `done`

## Objetivo

Garantir que a estrutura de agentes (Gemini, Claude, Copilot) seja consistente e funcional em todos os métodos de distribuição (Homebrew, npx, binário). O foco é replicar a excelência do Copilot para o Gemini, garantindo que MCPs, skills e prompts sejam instalados corretamente.

## Critério de conclusão

```bash
# 1. Criar um projeto demo com agente gemini
./dist/forge-sdd_darwin_arm64/forge-sdd init demo-gemini --yes --agent gemini

# 2. Validar existência de arquivos críticos
test -f demo-gemini/GEMINI.md
test -f demo-gemini/.gemini/system_instructions.md
test -f demo-gemini/.gemini/skills/orquestrador.md
test -f demo-gemini/.gemini/prompts/status.md
test -f demo-gemini/.vscode/mcp.json

# 3. Validar conteúdo (ex: MCPs configurados)
grep "context7" demo-gemini/.vscode/mcp.json
```

## Tarefas

- [x] **11-1** Revisar templates do Gemini para paridade total com o Lifecycle v1.1.0 do Copilot
- [x] **11-2** Garantir que o `mcp.json` seja gerado corretamente para todos os agentes
- [x] **11-3** Unificar nomenclatura de skills e prompts entre agentes (ex: `orquestrador.chatmode.md` vs `orquestrador.md`)
- [x] **11-4** Validar se o build via `goreleaser` e o pacote `npm` incluem todos os subdiretórios de templates (skills/prompts)
- [x] **11-5** Adicionar teste de integração que simula um `init` completo e checa a árvore de arquivos de cada agente

## Arquivos gerados/modificados

```
internal/scaffold/templates/agents/gemini/
internal/scaffold/templates/agents/claude/
internal/scaffold/templates/.github/
internal/scaffold/templates/.vscode/
internal/scaffold/scaffold_integration_test.go
```

## Skills relevantes

- `builder.md` (Go)
- `specifier.md` (Metodologia SDD)
