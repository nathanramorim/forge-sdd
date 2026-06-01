# feat/templates-embed

**Branch:** `feat/templates-embed`
**Fase:** 1
**Depende de:** `feat/foundation` (mergeada)
**Status:** `done`

## Objetivo
Todos os artefatos Forge-SDD embutidos no binário via `embed.FS` e acessíveis por walk em `scaffold.Walk()`.

## Critério de conclusão
```bash
go test ./internal/scaffold/... -run TestWalkTemplates -v
# → lista todos os arquivos de templates/ (≥25 arquivos), Exit 0
```

## Tarefas
- [x] **01-1** Criar pasta `templates/` com todos os artefatos Forge-SDD como `*.tmpl` (variáveis `{{.Project}}`, `{{.Stack}}`, etc.)
- [x] **01-2** Adicionar `//go:embed templates/**` em `internal/scaffold/scaffold.go`
- [x] **01-3** Implementar `scaffold.Walk() []string` que retorna lista de paths via embed.FS
- [x] **01-4** Escrever `TestWalkTemplates` em `internal/scaffold/scaffold_test.go` validando contagem e paths esperados

## Arquivos gerados
```
templates/
  .github/copilot-instructions.md.tmpl
  .github/chatmodes/orquestrador.chatmode.md.tmpl
  .github/chatmodes/specifier.chatmode.md.tmpl
  .github/chatmodes/builder.chatmode.md.tmpl
  .github/chatmodes/revisor.chatmode.md.tmpl
  .github/chatmodes/archivist.chatmode.md.tmpl
  .github/chatmodes/migrator.chatmode.md.tmpl
  .github/prompts/proxima-feature.prompt.md.tmpl
  .github/prompts/nova-feature.prompt.md.tmpl
  .github/prompts/revisar.prompt.md.tmpl
  .github/prompts/status.prompt.md.tmpl
  .github/prompts/archive.prompt.md.tmpl
  .github/prompts/doctor.prompt.md.tmpl
  .github/prompts/upgrade-sdd.prompt.md.tmpl
  .vscode/mcp.json.tmpl
  sdd/.sdd-version.tmpl
  sdd/.sddrc.tmpl
  sdd/README.md.tmpl
  sdd/memory/constitution.md.tmpl
  sdd/memory/progress.md.tmpl
  sdd/memory/progress-log.md.tmpl
  sdd/memory/mcps.md.tmpl
  sdd/spec/overview.md.tmpl
  sdd/spec/stack.md.tmpl
  sdd/spec/modules.md.tmpl
  sdd/spec/flows.md.tmpl
  sdd/spec/decisions.md.tmpl
  sdd/plan.md.tmpl
  sdd/features/index.md.tmpl
  sdd/features/feat-00-foundation.md.tmpl
  sdd/skills/index.md.tmpl
  sdd/.metrics/schema.json.tmpl
internal/scaffold/scaffold_test.go
```

## Skills relevantes
(consultar `skills/index.md`)
