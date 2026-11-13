# feat/deprecate-homebrew

**Branch:** `feat/remove-homebrew`
**Fase:** 29
**Depende de:** `feat-28-dev-flow-guide`
**Status:** `done`

## Objetivo
Remover o suporte e a publicação do CLI via Homebrew, consolidando a distribuição do `forge-sdd` exclusivamente via `npx` (NPM). Isso simplifica o pipeline de CI/CD, remove dependências de PATs (Personal Access Tokens) customizados no GitHub Actions e centraliza as releases no próprio repositório principal `forge-sdd`.

## Critério de conclusão
```bash
# O arquivo .goreleaser.yaml não deve conter a seção brews nem apontar para o repo homebrew-forge-sdd
! grep -q "brews:" .goreleaser.yaml && ! grep -q "name: homebrew-forge-sdd" .goreleaser.yaml

# O workflow release.yml não deve conter variáveis ou segredos do Homebrew
! grep -q "HOMEBREW_TAP_GITHUB_TOKEN" .github/workflows/release.yml

# O README.md não deve fazer referências a instalação via brew
! grep -q "brew install" README.md

# Os testes de unidade do scaffold devem passar limpos
go test ./...
```

## Tarefas
- [x] **29-1** Remover a seção `brews` e apontar o destino das releases para `forge-sdd` no `.goreleaser.yaml`.
- [x] **29-2** Simplificar o workflow `.github/workflows/release.yml` utilizando apenas o `GITHUB_TOKEN` nativo.
- [x] **29-3** Remover menções à instalação do Homebrew no `README.md` raiz.
- [x] **29-4** Atualizar o script de download do npm `npm/bin/run.js` para buscar binários no repo `forge-sdd` e usar o mapeamento de arquivos correto (prefixo `forge-sdd_` em vez de `homebrew-forge-sdd_`).
- [x] **29-5** Atualizar a `homepage` e o link do repositório em `npm/package.json` para apontar para `forge-sdd`.
- [x] **29-6** Remover o suporte a Homebrew nas skills de migração (`migrator.chatmode.md.tmpl` nos templates e no codebase).
- [x] **29-7** Atualizar documentações internas (`docs/metodologia-sdd.md` e `CONTRIBUTING.md`) removendo seções sobre Homebrew.
- [x] **29-8** Atualizar golden files de teste de scaffold (`go test ./internal/scaffold -update`).

## Arquivos gerados/modificados
```
.goreleaser.yaml
.github/workflows/release.yml
README.md
CONTRIBUTING.md
docs/metodologia-sdd.md
npm/bin/run.js
npm/package.json
internal/scaffold/templates/.github/chatmodes/migrator.chatmode.md.tmpl
internal/scaffold/templates/agents/gemini/.gemini/skills/migrator.chatmode.md.tmpl
```
