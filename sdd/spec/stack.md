# Stack — forge-sdd

## Dependências

| Camada | Lib | Versão | Import path |
|--------|-----|--------|-------------|
| Runtime | Go | 1.22+ | `GOTOOLCHAIN=local` |
| CLI | cobra | v1.8.x | `github.com/spf13/cobra` |
| Prompts | huh | v0.3.x | `github.com/charmbracelet/huh` |
| Templates | embed.FS | stdlib | `//go:embed templates/**` |
| Render | text/template | stdlib | `text/template` |
| Testes | testify | v1.9.x | `github.com/stretchr/testify` |
| Lint | golangci-lint | v1.57+ | via GitHub Actions |
| Release | goreleaser | v2.x | `.goreleaser.yaml` |

## Layout do projeto
```
cmd/forge-sdd/
  main.go                  # entrypoint — registra cobra root + init cmd
internal/
  config/config.go         # struct Config, Defaults(), FromFlags()
  scaffold/scaffold.go     # Walk + Run (render + write)
  scaffold/scaffold_test.go
  survey/survey.go         # formulário huh, retorna Config
templates/                 # embed.FS — todos os artefatos Forge-SDD como *.tmpl
  .github/
  sdd/
  .vscode/
tests/
  fixtures/expected/       # golden test: saída esperada de init --yes
go.mod
go.sum
.goreleaser.yaml
.github/workflows/
  ci.yml
  release.yml
```

## Características do binário
- Estático (`CGO_ENABLED=0`)
- Tamanho estimado: ~8 MB
- Sem acesso à rede em runtime
- Targets: linux/amd64, darwin/amd64, darwin/arm64, windows/amd64
