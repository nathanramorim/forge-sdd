# Guia do desenvolvedor — forge-sdd

## Estrutura do projeto

```
cmd/forge-sdd/main.go          → entrypoint cobra: flags, survey→scaffold, cmd version
internal/
  config/config.go             → struct Config, Defaults(), FromFlags()
  survey/survey.go             → formulário huh (5 campos interativos)
  scaffold/
    scaffold.go                → Walk() + Run(): renderiza embed.FS e escreve em disco
    scaffold_test.go           → TestWalkTemplates, TestDryRunNoFiles
    scaffold_integration_test.go → TestRunIntegration
    golden_test.go             → TestGoldenInit (compara contra testdata/golden/)
    testdata/golden/           → árvore de referência (32 arquivos renderizados)
    templates/                 → 32 templates *.tmpl embutidos via embed.FS
      .github/
      .vscode/
      sdd/
sdd/                           → documentação Forge-SDD deste próprio projeto
.goreleaser.yaml               → config de release (5 plataformas)
.github/workflows/
  ci.yml                       → go test + golangci-lint em push/PR
  release.yml                  → goreleaser em push v*
```

---

## Adicionar ou modificar um template

Todos os templates ficam em `internal/scaffold/templates/`. São arquivos Go `text/template` com as variáveis de `config.Config`:

| Variável | Tipo | Descrição |
|----------|------|-----------|
| `{{.Project}}` | string | Nome do projeto |
| `{{.Stack}}` | string | Runtime principal |
| `{{.DB}}` | string | Banco de dados |
| `{{.Telemetry}}` | bool | Telemetria habilitada |
| `{{.Lang}}` | string | Idioma (pt-BR / en) |
| `{{.SddVersion}}` | string | Versão Forge-SDD |

### Passos para adicionar um template
1. Crie o arquivo em `internal/scaffold/templates/<caminho>.tmpl`
2. O nome do arquivo gerado será o caminho sem o sufixo `.tmpl`
3. Rode `go test ./internal/scaffold/... -run TestWalkTemplates` — deve passar com contagem atualizada
4. Atualize o golden: `go test ./internal/scaffold/... -run TestGoldenInit -update`
5. Commit: `git add internal/scaffold/templates/ internal/scaffold/testdata/golden/`

---

## Adicionar um campo ao formulário

1. **`internal/config/config.go`** — adicione o campo em `Config` e valor padrão em `Defaults()`; adicione leitura em `FromFlags()` se quiser suporte via flag
2. **`internal/survey/survey.go`** — adicione o campo `huh.NewInput/Select/Confirm` ao grupo do formulário
3. **`cmd/forge-sdd/main.go`** — registre a nova flag em `initCmd.Flags()`
4. Atualize os templates que precisam da nova variável
5. Atualize o golden: `-run TestGoldenInit -update`

---

## Comandos do dia a dia

```bash
# compilar com versão injetada
go build -ldflags "-X main.version=1.1.0" -o /tmp/forge-sdd ./cmd/forge-sdd

# rodar todos os testes
go test ./...

# rodar apenas os testes do scaffold com verbose
go test ./internal/scaffold/... -v

# regenerar golden fixtures (após mudar templates)
go test ./internal/scaffold/... -run TestGoldenInit -update

# verificar lint
golangci-lint run

# build snapshot para todas as plataformas (requer goreleaser instalado)
goreleaser build --snapshot --clean
# → binários em dist/

# publicar release (cria tag e faz push)
git tag v1.2.0
git push origin v1.2.0
# → CI dispara release.yml → goreleaser → GitHub Releases
```

---

## Fluxo de uma nova feature

```
1. Ler sdd/memory/progress.md
2. git checkout -b feat/<nome>
3. Implementar tarefas do feat-XX-*.md
4. go test ./... → PASS
5. go test ./internal/scaffold/... -run TestGoldenInit -update  (se mudou templates)
6. git commit -m "feat(XX): ..."
7. Atualizar sdd/memory/progress.md
8. git commit -m "chore: atualiza progress.md"
```

---

## Dependências principais

| Lib | Versão | Função |
|-----|--------|--------|
| `cobra` | v1.9.1 | CLI framework |
| `huh` | v0.3.0 | Formulário interativo no terminal |
| `testify` | v1.9.0 | Assertions nos testes |
| `goreleaser` | v2 | Build e release multi-plataforma |

> Antes de adicionar qualquer lib nova, consultar **context7** com o nome e versão exatos.
