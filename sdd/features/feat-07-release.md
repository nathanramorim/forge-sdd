# feat/release

**Branch:** `feat/release`
**Fase:** 7
**Depende de:** `feat/self-test` (mergeada)
**Status:** `todo`

## Objetivo
Binários multi-OS gerados via goreleaser e publicados automaticamente em GitHub Releases ao criar tag `v*`.

## Critério de conclusão
```bash
goreleaser build --snapshot --clean
# → binários em dist/ para linux/amd64, darwin/amd64, darwin/arm64, windows/amd64
# → Exit 0
ls dist/
# → deve conter 4 diretórios de plataforma
```

## Tarefas
- [ ] **07-1** Criar `.goreleaser.yaml` com targets: `linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- [ ] **07-2** Adicionar `ldflags` com `-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}`
- [ ] **07-3** Criar `.github/workflows/release.yml` acionado em `push: tags: ['v*']`
- [ ] **07-4** Testar `goreleaser build --snapshot --clean` localmente (consultar context7 para goreleaser v2)
- [ ] **07-5** Adicionar instruções de instalação no README do projeto

## Arquivos gerados
```
.goreleaser.yaml
.github/workflows/release.yml
README.md                    (instruções de instalação)
```

## Skills relevantes
(consultar `skills/index.md`)
