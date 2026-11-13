# feat/03-ergonomia-de-comandos-e-sincronizacao (03-02)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-02
**Depende de:** feat-03-01 (novo template `.claude/commands/*.md` precisa existir)
**Status:** `done`

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

**Desvio consciente do critério 1/2 original, registrado aqui:** a análise de código (`shouldPreserve` em `scaffold.go`) mostrou que `.claude/commands/` **nunca** foi preservado como domínio do usuário — só `sdd/` e `.agents/rules/` são. Ou seja, `.claude/commands/*.prompt.md` já era 100% regenerado/sobrescrito a cada `init`/`update` antes desta feature; não havia detecção de customização a fazer porque o comportamento atual do repositório nunca ofereceu esse contrato para esse caminho. Implementar "detecção de customização" ali teria sido uma garantia nova e não solicitada, inconsistente com o resto do código. O gap real era outro, mais simples: `renderDir` escreve o novo nome (`*.md`) mas nunca apaga o antigo (`*.prompt.md`), deixando lixo duplicado.

Implementado: `cleanObsoleteFiles()` em `scaffold.go` ganhou os 12 nomes antigos de `.claude/commands/*.prompt.md` (reaproveitando `commandOrder` de `cheatsheet.go`, sem duplicar a lista). Durante o teste, encontrado e corrigido um bug pré-existente: `cleanObsoleteFiles()` rodava incondicionalmente, inclusive em `--dry-run`, apagando arquivos reais — corrigido com guarda `!cfg.DryRun` em `Run()`/`RunAgents()` (cumpre Regra 9 da Constituição, que este código já deveria cumprir antes). Novo teste `TestUpdateCleansObsoleteClaudePromptSuffix` cobre remoção do arquivo antigo, idempotência e `--dry-run` sem efeito colateral. `go build/vet/test` passam. Commit `577c28b`.
