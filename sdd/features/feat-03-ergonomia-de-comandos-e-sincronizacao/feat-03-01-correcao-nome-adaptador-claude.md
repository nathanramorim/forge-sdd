# feat/03-ergonomia-de-comandos-e-sincronizacao (03-01)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-01
**Depende de:** — (sem dependência; menor risco do pacote)
**Status:** `todo`

## Objetivo

Corrigir o nome dos 12 arquivos de comando do adaptador Claude: `.claude/commands/<comando>.prompt.md` → `.claude/commands/<comando>.md`. O Claude Code descobre slash commands pelo nome do arquivo sem a extensão `.md` — hoje `nova-feature.prompt.md` registra `/nova-feature.prompt`, não `/nova-feature`, quebrando todo handoff que recomenda o comando sem sufixo.

## Critérios de Aceitação Executáveis

1. Os 12 arquivos em `.claude/commands/` (repositório) e o template-fonte correspondente (`internal/scaffold/templates/agents/claude/.claude/commands/*.md.tmpl`, hoje `*.prompt.md.tmpl`) são renomeados para `<comando>.md`, preservando o conteúdo (referência a `.agents/commands/<comando>.md`).
2. Gemini (`.gemini/prompts/*.prompt.md`) e Copilot (`.github/prompts/*.prompt.md`) permanecem inalterados — a extensão `.prompt.md` é a convenção correta de cada uma dessas ferramentas.
3. `forge-sdd init` em projeto novo escaffolda `.claude/commands/*.md` (sem sufixo `.prompt`).
4. Golden fixtures de `internal/scaffold` (`internal/scaffold/testdata/golden/.claude/commands/*`) regeneradas refletindo o novo nome.
5. `go build ./...`, `go vet ./...` e `go test ./...` passam.

## Handoff

Pré-requisito da feature 03-02 (migração em `forge-sdd update` depende do novo template já existir).
