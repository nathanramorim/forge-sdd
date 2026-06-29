# feat/update-agent-prompts

**Branch:** `feat/update-agent-prompts`
**Fase:** 34
**Depende de:** `feat-33-lp-wiki-sync`
**Status:** `done`

## Objetivo

Migrar o frontmatter das diretivas de prompts do GitHub Copilot da chave antiga `mode` para `agent` (ex: `agent: agent` ou `agent: ask`). Alterar também os prompts `status` e `doctor` para usar `agent: agent` (em vez de `ask`) para habilitar permissões de execução de scripts bash/comandos no terminal do Copilot.

## Critério de conclusão

```bash
# 1. Nenhum arquivo .prompt.md or .prompt.md.tmpl deve conter a chave "mode:" no frontmatter
! grep -r "mode:" .github/prompts/ internal/scaffold/templates/.github/prompts/ internal/scaffold/testdata/golden/.github/prompts/

# 2. Os prompts 'status' e 'doctor' devem estar configurados como "agent: agent" no workspace e nos templates
grep -q "agent: agent" .github/prompts/status.prompt.md
grep -q "agent: agent" .github/prompts/doctor.prompt.md
grep -q "agent: agent" internal/scaffold/templates/.github/prompts/status.prompt.md.tmpl
grep -q "agent: agent" internal/scaffold/templates/.github/prompts/doctor.prompt.md.tmpl

# 3. Testes do projeto devem passar com sucesso
go test ./...
go vet ./...
```

## Tarefas

- [x] **34-1** Criar especificação da feature em `sdd/features/feat-34-update-agent-prompts.md`
- [x] **34-2** Atualizar `sdd/memory/progress.md` e `sdd/features/index.md`
- [x] **34-3** Atualizar prompts em `internal/scaffold/templates/.github/prompts/`
- [x] **34-4** Atualizar prompts no workspace `.github/prompts/`
- [x] **34-5** Atualizar prompts em `internal/scaffold/testdata/golden/.github/prompts/`
- [x] **34-6** Executar testes unitários e de integração (`go test ./...` e `go vet ./...`)
