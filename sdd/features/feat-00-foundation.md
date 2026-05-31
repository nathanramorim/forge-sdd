# feat/foundation

**Branch:** `feat/foundation`
**Fase:** 0
**Depende de:** —
**Status:** `todo`

## Objetivo
Repositório Go compilando com layout de pastas correto, cobra registrado e stubs dos módulos internos.

## Critério de conclusão
```bash
go build -o forge-sdd ./cmd/forge-sdd && ./forge-sdd --help
# → exibe help do cobra sem erros, Exit 0
go vet ./...
# → Exit 0, sem erros
```

## Tarefas
- [ ] **00-1** `go mod init github.com/forge-sdd/cli`
- [ ] **00-2** Criar `cmd/forge-sdd/main.go` com cobra root command e `init` subcommand (stub)
- [ ] **00-3** Criar `internal/config/config.go` com struct `Config` e `Defaults()`
- [ ] **00-4** Criar `internal/scaffold/scaffold.go` (stub: Walk e Run retornam nil)
- [ ] **00-5** Criar `internal/survey/survey.go` (stub: Run retorna `Config{}`)
- [ ] **00-6** `go vet ./...` passa sem erros

## Arquivos gerados
```
go.mod
cmd/forge-sdd/main.go
internal/config/config.go
internal/scaffold/scaffold.go
internal/survey/survey.go
```

## Skills relevantes
(consultar `skills/index.md`)
