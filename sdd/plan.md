# Plano — forge-sdd

## Fases

| Fase | Feature | Entrega | Critério resumido |
|------|---------|---------|-------------------|
| 0 | feat-00-foundation | Repo Go compilando com cobra | `go build ./... && ./forge-sdd --help` |
| 1 | feat-01-templates-embed | Templates embutidos via embed.FS | `go test ./internal/scaffold/... -run TestWalkTemplates` |
| 2 | feat-02-init-interactive | Survey + scaffold funcional | `forge-sdd init` cria árvore completa |
| 3 | feat-03-init-flags | Modo não-interativo (`--yes`) | `forge-sdd init --yes` sem prompts |
| 4 | feat-04-dry-run | Preview sem criar arquivos | `forge-sdd init --dry-run` → Exit 0, sem arquivos criados |
| 5 | feat-05-versioning | `.sdd-version` + `.sddrc` corretos | `cat sdd/.sdd-version` → `1.1.0` |
| 6 | feat-06-self-test | Golden test + CI | `go test ./...` passa |
| 7 | feat-07-release | Binários multi-OS | `goreleaser build --snapshot --clean` |

## Paralelismo
Nenhuma feature pode ser paralelizada — pipeline sequencial (cada fase depende da anterior).

## Dependências externas a instalar
```bash
go get github.com/spf13/cobra@v1.8.1
go get github.com/charmbracelet/huh@v0.3.0
go get github.com/stretchr/testify@v1.9.0
```
> Consultar context7 com versão exata antes de cada `go get`.
