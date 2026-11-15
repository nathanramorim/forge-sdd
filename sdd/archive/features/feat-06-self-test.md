# feat/self-test

**Branch:** `feat/self-test`
**Fase:** 6
**Depende de:** `feat/versioning` (mergeada)
**Status:** `todo`

## Objetivo
Golden test automatizado valida que `init --yes` produz exatamente a árvore esperada. CI passa com `go test + golangci-lint`.

## Critério de conclusão
```bash
go test ./... -v
# → PASS em todos os pacotes, incluindo TestGoldenInit, Exit 0
```

## Tarefas
- [ ] **06-1** Criar `tests/fixtures/expected/` com a árvore completa esperada de `forge-sdd init --yes --stack go --db postgres`
- [ ] **06-2** Escrever `TestGoldenInit` em `internal/scaffold/golden_test.go`: roda `scaffold.Run` em tmpdir, faz diff recursivo contra fixture
- [ ] **06-3** Criar `.github/workflows/ci.yml`: `go test ./...` + `golangci-lint run` em push/PR para `main`
- [ ] **06-4** Garantir que todos os testes de fases anteriores ainda passam

## Arquivos gerados
```
tests/fixtures/expected/         (árvore completa de referência)
internal/scaffold/golden_test.go
.github/workflows/ci.yml
```

## Skills relevantes
(consultar `skills/index.md`)
