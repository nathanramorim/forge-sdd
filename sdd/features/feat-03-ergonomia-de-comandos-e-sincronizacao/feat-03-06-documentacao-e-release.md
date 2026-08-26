# feat/03-ergonomia-de-comandos-e-sincronizacao (03-06)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-06
**Depende de:** feat-03-01, feat-03-02, feat-03-03, feat-03-04, feat-03-05 (todas `done`)
**Status:** `todo`

## Objetivo

Fechar o pacote: atualizar documentação e (se aplicável) release notes citando a correção de nome do adaptador Claude, a nova etapa de `/status`, a nova regra de clarify e a nova pergunta de confirmação de subagente — respeitando Regra 12 da Constituição (agente/comando identificado explicitamente).

## Critérios de Aceitação Executáveis

1. `README.md`/`npm/README.md` refletem: comando Claude sem sufixo `.prompt` corrigido, nova etapa de `/status`, novo passo de clarify, nova pergunta de subagente.
2. Se a mudança justificar bump de versão (`sdd/.sddrc`, `internal/config/config.go`, `npm/package.json`), release notes em `sdd/releases/history.md` com bullets reais do que foi melhorado (Regra 12 da Constituição — falha se cair no texto genérico de fallback), identificando agente/comando explicitamente.
3. `sdd/features/index.md` e `sdd/memory/progress.md` atualizados refletindo status `done` de todas as subtarefas 03-01 a 03-05.
4. `go build ./...`, `go vet ./...` e `go test ./...` passam.

## Handoff

Última etapa do pacote — só inicia depois de 03-01 a 03-05 estarem `done`.
