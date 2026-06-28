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

## Publicar uma release

O goreleaser faz tudo automaticamente ao receber uma tag `v*`. Ele:
1. Compila para 5 plataformas (linux/darwin × amd64/arm64 + windows/amd64)
2. Cria o GitHub Release com os archives e `checksums.txt` diretamente neste repositório
3. O NPM downloader (`npm/bin/run.js`) busca os binários automaticamente a partir do GitHub Release

### Pré-requisitos

Nenhum segredo ou PAT adicional é necessário. O workflow usa o `GITHUB_TOKEN` nativo do repositório.

### Fluxo de release

```bash
# commitar tudo que está pronto
git add .
git commit -m "chore: prepara v1.x.0"

# criar e enviar a tag → dispara release.yml
git tag v1.x.0
git push origin main
git push origin v1.x.0
```

### Pré-release (beta) — testar antes de aprovar o PR

Use quando quiser validar uma feature branch via npx antes de mergear para main.
A tag deve conter `-` (ex: `v1.6.0-beta.0`) — o workflow detecta isso e publica com
`--tag beta` no npm, sem tocar o canal `@latest`.

```bash
# na feature branch, com tudo commitado
git tag v1.6.0-beta.0
git push origin v1.6.0-beta.0
# → CI compila + publica @nathanramorim/forge-sdd@beta
```

Testando como usuário final:
```bash
npx @nathanramorim/forge-sdd@beta update --yes --agent claude
```

Quando o PR for aprovado e mergear para main, publique a versão estável normalmente:
```bash
git tag v1.6.0
git push origin v1.6.0
# → publica em @latest, sobrepõe o beta
```

> **Convenção de numeração:** use o mesmo número da próxima versão estável com sufixo
> `-beta.N`. Assim `v1.6.0-beta.0` → `v1.6.0` é o ciclo completo de uma feature.

### Diferenças entre `@latest` e `@beta`

| | `@latest` | `@beta` |
|---|---|---|
| **Quem recebe** | Todos os usuários do `npx` | Apenas quem passa `@beta` explicitamente |
| **Quando é atualizado** | Push de tag sem `-` (ex: `v1.6.0`) | Push de tag com `-` (ex: `v1.6.0-beta.0`) |
| **Estabilidade** | Aprovado e mergeado em `main` | Feature branch em validação |
| **Cache local** | `~/.cache/forge-sdd/1.6.0/` | `~/.cache/forge-sdd/1.6.0-beta.0/` |
| **Binário no GitHub** | Release público, sem label | Release marcado como *pre-release* |

**Regra prática:** um push de `@beta` nunca afeta `@latest`. São canais independentes no npm registry. O `@latest` só avança quando você publica uma tag estável.

### Recriar uma tag já publicada (correção)

```bash
git tag -d v1.x.0                  # apaga local
git tag v1.x.0                     # recria no commit atual
git push origin :v1.x.0            # apaga no remote
git push origin v1.x.0             # sobe novamente → dispara o workflow
```

### Usuários na versão antiga (sem o comando `update`)

O cache do npx é por versão (`~/.cache/forge-sdd/<versão>/`). Quando uma nova versão
estável é publicada, o npx baixa o novo binário automaticamente na próxima execução.

Para forçar a migração sem esperar:
```bash
# seguro: shouldPreserve protege tudo em sdd/ (features, spec, memory)
npx @nathanramorim/forge-sdd@latest init . --yes --agent <agentes-atuais>
```

Após isso o binário da nova versão estará no cache e `forge-sdd update` ficará disponível.

### Instalação pelo usuário final (após o primeiro release)

```bash
npx @nathanramorim/forge-sdd@latest init
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
