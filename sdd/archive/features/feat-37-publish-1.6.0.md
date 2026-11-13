# feat/publish-1.6.0

**Branch:** `feat/publish-1.6.0`
**Fase:** 37
**Depende de:** `feat-36-structured-documentation`
**Status:** `done`

## Objetivo

Atualizar todas as referências de versão do projeto de `1.5.3` para `1.6.0` (estável) e documentar as entregas recentes no histórico de releases, preparando o repositório para a publicação oficial da versão 1.6.0 estável via tags do GitHub e npm (com a tag default `latest`).

## Critério de conclusão

```bash
# 1. A versão atualizada deve estar presente nos arquivos principais do CLI e metadados
grep -q "1.6.0" GEMINI.md
grep -q "1.6.0" cmd/forge-sdd/main.go
grep -q "1.6.0" docs/metodologia-sdd.md
grep -q "1.6.0" internal/config/config.go
grep -q "1.6.0" npm/package.json
grep -q "1.6.0" npm/README.md

# 2. Arquivos de scaffolding e golden files devem refletir a versão 1.6.0
grep -q "1.6.0" internal/scaffold/testdata/golden/sdd/.sdd-version
grep -q "1.6.0" internal/scaffold/testdata/golden/sdd/.sddrc

# 3. Todos os testes unitários e de integração devem passar sem quebras
go test ./...
go vet ./...
```

## Tarefas

- [x] **37-1** Criar especificação da feature em `sdd/features/feat-37-publish-1.6.0.md`
- [x] **37-2** Atualizar `sdd/features/index.md` e `sdd/memory/progress.md`
- [x] **37-3** Atualizar referências de versão em metadados (`GEMINI.md`, `npm/package.json`, `npm/README.md`)
- [x] **37-4** Atualizar versão no código fonte (`cmd/forge-sdd/main.go`, `internal/config/config.go`)
- [x] **37-5** Atualizar versão em templates, testes e golden files (`internal/scaffold/scaffold_test.go`, `internal/scaffold/golden_test.go`, `internal/scaffold/scaffold_integration_test.go`, e pasta `testdata/golden/`)
- [x] **37-6** Atualizar documentação metodológica (`docs/metodologia-sdd.md`)
- [x] **37-7** Adicionar notas de release da versão 1.6.0 em `sdd/releases/history.md` (englobando as fases 30 a 36)
- [x] **37-8** Rodar suite de testes e validar conformidade (`go test ./...` e `go vet ./...`)
