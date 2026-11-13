# fix/rename-agent-dir-para-agents

**Branch:** `claude/agent-folder-rename-forge-ww8abl`
**Fase:** 52
**Depende de:** feat-02-agent-rules-e-branch-por-feature (`.agent/` — não mesclada em `main`)
**Status:** `done`

## Objetivo

Renomear a pasta de fonte única de agente de `.agent/` para `.agents/`
(pedido explícito do usuário), garantindo que `forge-sdd update` migre
projetos já escaffoldados com o nome anterior sem apagar `rules/` já
criadas pelo usuário.

## Causa raiz / contexto

`.agent/` foi introduzida no commit `ae906d7` (branch
`feat/02-agent-rules-e-branch-por-feature`, nunca mesclada em `main`), com
duas subpastas: `.agent/rules/` (regras de domínio do usuário, preservadas
em update via `shouldPreserve`) e `.agent/commands/` (corpo canônico dos
13 comandos SDD, sempre regenerado). Este fix parte dessa branch inteira
(13 commits à frente de `main`) e renomeia o caminho antes de qualquer
merge em `main`, evitando que o nome `.agent` chegue a ser publicado.

## Critérios de Aceitação Executáveis

1. Todo caminho gerado/embutido usa `.agents/` — templates-fonte
   (`internal/scaffold/templates/.agents/**`), fixtures golden
   (`internal/scaffold/testdata/golden/.agents/**`), a pasta dogfooded na
   raiz deste repo, os adaptadores finos (`.claude/commands/*.prompt.md`,
   `.gemini/prompts/*.prompt.md`, `.github/prompts/*.prompt.md`) e a
   documentação (`README.md`, `npm/README.md`, `CLAUDE.md`, `GEMINI.md`).
2. `internal/scaffold/scaffold.go`: `globalRoots` aponta para
   `templates/.agents`; `shouldPreserve` preserva `.agents/rules/` (não
   mais `.agent/rules/`).
3. `forge-sdd update` em um projeto que ainda tenha `.agent/` (nome
   anterior) migra automaticamente para `.agents/` sem apagar nenhum
   arquivo do usuário: `.agent/rules/*` sobrevive com conteúdo intacto,
   `.agent/` deixa de existir após a migração (renomeada, não duplicada),
   e `.agents/commands/*` é regenerado a partir do template atual — mesmo
   quando o `update` é executado mais de uma vez seguida (idempotente).
4. `forge-sdd init` em projeto novo nunca cria `.agent/` — gera direto
   `.agents/`.
5. Registros históricos (`sdd/discovery/*-02-*`, `sdd/features/feat-02-*`,
   `sdd/memory/progress-log.md`, `sdd/releases/history.md`) permanecem
   citando `.agent/` como estava — documentam o que foi decidido/construído
   na época; a renomeação é registrada aqui, não reescrita retroativamente
   neles.
6. `go build ./...`, `go vet ./...` e `go test ./...` passam, incluindo um
   novo teste cobrindo o cenário de migração (`.agent/` legado → `.agents/`
   preservando `rules/` do usuário e regenerando `commands/`).

## Handoff

Implementado: `git mv` das três pastas físicas (`.agent/`,
`internal/scaffold/templates/.agent/`,
`internal/scaffold/testdata/golden/.agent/`) para `.agents/`; substituição
textual de `.agent/` → `.agents/` em adaptadores/docs (excluindo os
registros históricos citados no critério 5); nova função
`migrateLegacyAgentDir` em `internal/scaffold/scaffold.go` (mesmo estilo de
`cleanObsoleteFiles`), chamada no início de `Run()`/`RunAgents()` (só fora
de `--dry-run`) — rename atômico quando `.agents/` ainda não existe, merge
arquivo-a-arquivo sem sobrescrever quando já existe (idempotência); novo
teste `TestUpdateMigratesLegacyAgentDirToAgents` em
`internal/scaffold/scaffold_test.go`. Pronto para `/proxima-feature`
(commit, push e fechamento do lifecycle).
