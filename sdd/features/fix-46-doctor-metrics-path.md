# Fix 46 — `doctor` Verifica `.metrics/schema.json` na Raiz em Vez de `sdd/.metrics/`

**Branch:** `fix/update-beta-version-detection` (agrupada com a fix-45, mesma leva/PR)
**Depende de:** —

## Descrição
Os prompts `/doctor` (Copilot, Claude, Gemini) instruem o agente a verificar a presença de `.metrics/schema.json` na raiz do projeto. O arquivo, no entanto, é sempre escaffoldado em `sdd/.metrics/schema.json` (`internal/scaffold/templates/sdd/.metrics/schema.json.tmpl`), e é para lá que o Orquestrador grava as métricas de sessão (`sdd/.metrics/session-*.json`, conforme `orquestrador.chatmode.md`). Como resultado, o `doctor` sempre reporta o item de métricas como ausente (✗), mesmo em projetos saudáveis.

## Critérios de Aceitação Executáveis

1. Os templates de prompt `/doctor` para Copilot (`.github/prompts/doctor.prompt.md.tmpl`), Claude (`.claude/commands/doctor.prompt.md.tmpl`) e Gemini (`.gemini/prompts/doctor.prompt.md.tmpl`) devem referenciar `sdd/.metrics/schema.json`, não `.metrics/schema.json`.
2. As cópias já renderizadas na própria estrutura `sdd/` deste repositório (dogfooding) devem ser atualizadas para refletir o caminho correto.
3. Os fixtures golden de teste (`internal/scaffold/testdata/golden/`) devem ser regenerados e permanecer consistentes com os templates corrigidos.

## Status: done

Corrigido nos 3 templates (`internal/scaffold/templates/.github/prompts/doctor.prompt.md.tmpl`, `.../agents/claude/.claude/commands/doctor.prompt.md.tmpl`, `.../agents/gemini/.gemini/prompts/doctor.prompt.md.tmpl`) e na cópia dogfooded `.github/prompts/doctor.prompt.md`. Golden fixture regenerado via `go test ./internal/scaffold/... -run TestGoldenInit -update`. `go build ./...` e `go test ./...` confirmados sem regressões.
