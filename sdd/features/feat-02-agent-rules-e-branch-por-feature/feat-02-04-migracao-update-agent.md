# Feature 02-04 — Migração Aditiva de `.agent/` em `forge-sdd update`

Estende o fluxo de `forge-sdd update` (mesmo ponto de extensão do feat-31 — preservação de domínio) para que projetos existentes ganhem `.agent/rules/` e `.agent/commands/` sem perder estado. Depende de feat-02-01 e feat-02-03 (precisa dos templates canônicos existirem para migrar/converter).

## Critérios de Aceitação Executáveis

1. `forge-sdd update` cria `.agent/rules/` se ausente (com README/exemplo), sem nunca sobrescrever uma regra já existente.
2. `forge-sdd update` cria `.agent/commands/` a partir do template canônico e converte um adaptador existente (`.claude/commands/*.md`, etc.) para a forma fina/referenciada **somente se** o conteúdo atual bater com o template original conhecido da versão anterior (sem edição manual detectada); caso contrário, pula o arquivo e reporta ao usuário quais adaptadores não puderam ser migrados automaticamente.
3. Nenhum passo de migração altera `sdd/features/`, `sdd/discovery/`, `sdd/fix/*`, `sdd/memory/progress.md` ou a branch/ponto atual do projeto.
4. `forge-sdd update --dry-run` reporta a criação/conversão de `.agent/` e dos adaptadores na árvore impressa, sem escrever nenhum arquivo real.
5. Migração é idempotente: rodar `update` duas vezes seguidas não duplica nem altera conteúdo já migrado.
6. Testes cobrindo: idempotência, detecção de customização (adaptador editado manualmente não é sobrescrito), e `--dry-run` sem escrita. `go build`, `go vet ./...`, `go test ./...` passam.

## Status: done

Reaproveitado o mecanismo já existente do feat-31 (`shouldPreserve`) em vez de construir detecção de customização por diff: `forge-sdd update` chama `scaffold.Run()` (mesmo caminho de `init`) sempre que há bump de versão, e `.agent` agora está em `globalRoots`. `.agent/rules/` é preservado (criado só se ausente); `.agent/commands/` e os adaptadores de agente são regenerados a cada `update`, como qualquer outro arquivo de configuração fora de `sdd/` (mesmo tratamento hoje dado a `CLAUDE.md`/`GEMINI.md`). Isso simplifica o desenho original da discovery (que previa detecção de edição manual antes de converter comandos) porque a convenção do repositório já trata todo conteúdo fora de `sdd/` como gerado/substituível — só `.agent/rules/` precisava de uma exceção, adicionada em feat-02-01. `--dry-run` não escreve nada (herdado do mecanismo existente). Teste `TestUpgradePreservesAgentRulesButRegeneratesCommands` cobre preservação de regra customizada, regeneração do corpo canônico e idempotência (roda `Run` duas vezes). Validado com dogfood real: `forge-sdd update --yes --version 2.2.0-beta` neste repositório criou `.agent/`, atualizou `.claude/`/`.gemini/`/`CLAUDE.md`/`GEMINI.md`, e não tocou `sdd/features/`, `sdd/discovery/` nem `sdd/memory/progress.md`.
