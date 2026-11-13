# feat/03-ergonomia-de-comandos-e-sincronizacao (03-06)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-06
**Depende de:** feat-03-01, feat-03-02, feat-03-03, feat-03-04, feat-03-05 (todas `done`)
**Status:** `done`

## Objetivo

Fechar o pacote: atualizar documentação e (se aplicável) release notes citando a correção de nome do adaptador Claude, a nova etapa de `/status`, a nova regra de clarify e a nova pergunta de confirmação de subagente — respeitando Regra 12 da Constituição (agente/comando identificado explicitamente).

## Critérios de Aceitação Executáveis

1. `README.md`/`npm/README.md` refletem: comando Claude sem sufixo `.prompt` corrigido, nova etapa de `/status`, novo passo de clarify, nova pergunta de subagente.
2. Se a mudança justificar bump de versão (`sdd/.sddrc`, `internal/config/config.go`, `npm/package.json`), release notes em `sdd/releases/history.md` com bullets reais do que foi melhorado (Regra 12 da Constituição — falha se cair no texto genérico de fallback), identificando agente/comando explicitamente.
3. `sdd/features/index.md` e `sdd/memory/progress.md` atualizados refletindo status `done` de todas as subtarefas 03-01 a 03-05.
4. `go build ./...`, `go vet ./...` e `go test ./...` passam.

## Handoff

Implementado: `README.md`/`npm/README.md` ganharam a seção "Ergonomia de Comandos e Sincronização (v2.3.0-beta)" com os 4 destaques de feat-03; `sdd/releases/history.md` ganhou a seção "Versão 2.3.0-beta" com bullets reais cobrindo fix-52 + as 5 entregas técnicas de feat-03, identificando agente/comando explicitamente (Regra 12). Bump de versão em `sdd/.sddrc`, `sdd/.sdd-version`, `npm/package.json`, `internal/config/config.go` e `cmd/forge-sdd/main.go` (`2.2.0-beta` → `2.3.0-beta`). `sdd/features/index.md` e `sdd/memory/progress.md` atualizados refletindo `done` em 03-01 a 03-06. Tag `v2.3.0-beta` criada na própria branch (Regra 11 — beta não mescla em `main`), acumulando feat-02 + fix-52 + feat-03. `go build/vet/test` passam.
