# Feature 02-01 — `.agent/rules/` — Convenção e Scaffold em `init`

Base do pacote, sem risco de conteúdo pré-existente para migrar (pasta nova). Introduz `.agent/rules/` como pasta neutra na raiz do projeto para regras de domínio livres (design system, frontend, arquitetura, acessibilidade, ...), consultável por qualquer agente sem duplicação.

## Critérios de Aceitação Executáveis

1. `internal/scaffold/templates/.agent/rules/` existe com um `README.md` explicando a convenção (o que é, como nomear arquivos, que o conteúdo é livre do usuário) e um arquivo de amostra comentado (ex: `design-system.md.example`) demonstrando o formato esperado.
2. `forge-sdd init` em projeto novo escaffolda `.agent/rules/` com esses arquivos, embutidos via `embed.FS` (Regra 3 da Constituição).
3. `forge-sdd init --dry-run` lista `.agent/rules/` na árvore impressa sem criar arquivo real (Regra 9 da Constituição).
4. `go build`, `go vet ./...` e `go test ./...` passam; golden fixtures de `internal/scaffold` regeneradas cobrindo `.agent/rules/`.

## Status: done

Criado `internal/scaffold/templates/.agent/rules/README.md.tmpl` (explica a convenção) e `example.md.example.tmpl` (amostra de design system, extensão `.example` para não ser lido como regra real). `scaffold.go` passou a incluir `templates/.agent` em `globalRoots` (renderizado sempre, independente de agente selecionado). `shouldPreserve()` ganhou um caso especial: qualquer caminho sob `.agent/rules/` é preservado uma vez criado (mesmo espírito de preservação de domínio já usado em `sdd/`) — nunca sobrescrito em `update`/`init` subsequentes. `--dry-run` já respeita isso via o mecanismo existente de `renderDir`. Testes: `TestWalkTemplates` (novos paths essenciais) e `TestUpgradePreservesAgentRulesButRegeneratesCommands` (novo, cobre também feat-02-03/04). Golden fixtures regeneradas. `go build`, `go vet ./...`, `go test ./...` passam.
