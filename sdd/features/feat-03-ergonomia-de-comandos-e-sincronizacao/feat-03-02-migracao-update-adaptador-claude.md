# feat/03-ergonomia-de-comandos-e-sincronizacao (03-02)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-02
**Depende de:** feat-03-01 (novo template `.claude/commands/*.md` precisa existir)
**Status:** `todo`

## Objetivo

Estender `forge-sdd update` para migrar projetos já escaffoldados com o nome antigo (`.claude/commands/*.prompt.md`) para o novo (`*.md`), de forma aditiva e idempotente, sem apagar customização do usuário — mesmo princípio já usado para a migração `.agent/` → `.agents/`.

## Critérios de Aceitação Executáveis

1. `forge-sdd update` em projeto existente detecta `.claude/commands/*.prompt.md` remanescente e, se o conteúdo bater com o template anterior conhecido (sem customização detectada), renomeia para `*.md`.
2. Se o conteúdo divergir do template conhecido (customização manual detectada), o arquivo antigo é preservado com o conteúdo intacto e o usuário é avisado — nunca sobrescrito ou apagado silenciosamente (decisão resolvida da Constituição).
3. Migração é idempotente — rodar `update` duas vezes seguidas não duplica nem quebra arquivos.
4. `forge-sdd update --dry-run` reporta a renomeação/migração na árvore impressa, sem escrever nenhum arquivo real (Regra 9 da Constituição).
5. Novo teste cobrindo o cenário de migração (`.prompt.md` legado → `.md`, com e sem customização detectada, incluindo idempotência).
6. `go build ./...`, `go vet ./...` e `go test ./...` passam.

## Handoff

Fecha a correção de nome do adaptador Claude (03-01 + 03-02). Próxima etapa do pacote: 03-03 (sincronização remota em `/status`), sem dependência técnica com esta.
