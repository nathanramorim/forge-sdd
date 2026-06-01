# feat/npx

**Branch:** `feat/npx`
**Fase:** 8
**Depende de:** `feat/release` (mergeada — binários publicados em GitHub Releases)
**Status:** `done`

## Objetivo
Publicar um pacote npm (`forge-sdd`) que permita execução via `npx forge-sdd init` sem instalar Go.

O pacote age como **thin wrapper**: detecta plataforma (OS + arch), baixa o binário correspondente do GitHub Releases, e executa com os args passados.

## Critério de conclusão
```bash
npx forge-sdd@latest init --yes --dry-run
# → exibe o preview dos arquivos que seriam criados
# → Exit 0
```

## Estratégia técnica

### Estrutura do pacote npm
```
npm/
  package.json      # name: forge-sdd, bin: { forge-sdd: "bin/run.js" }
  bin/
    run.js          # wrapper Node.js: detecta plataforma → baixa binário → executa
  README.md
```

### Detecção de plataforma
| `process.platform` | `process.arch` | Asset GitHub                          |
|--------------------|----------------|---------------------------------------|
| linux              | x64            | forge-sdd_linux_amd64.tar.gz          |
| linux              | arm64          | forge-sdd_linux_arm64.tar.gz          |
| darwin             | x64            | forge-sdd_darwin_amd64.tar.gz         |
| darwin             | arm64          | forge-sdd_darwin_arm64.tar.gz         |
| win32              | x64            | forge-sdd_windows_amd64.zip           |

### Download do binário
- URL: `https://github.com/nathanramorim/homebrew-forge-sdd/releases/download/vVERSION/ASSET`
- Versão: lida de `package.json#version` (mantida em sincronia com tags Go)
- Cache: `~/.cache/forge-sdd/VERSION/` (evita re-download)
- Verificação: SHA256 via `checksums.txt` publicado pelo goreleaser

## Tarefas
- [ ] **08-1** Criar `npm/package.json` com `name: forge-sdd`, `bin`, `version` inicial `1.1.0`
- [ ] **08-2** Criar `npm/bin/run.js` com lógica de detecção, cache e execução do binário
- [ ] **08-3** Adicionar `.github/workflows/npm-publish.yml` acionado em `push: tags: ['v*']` após goreleaser (usa `NPM_TOKEN`)
- [ ] **08-4** Adicionar `npm/` ao `.goreleaser.yaml` como `after hook` para `npm publish` (ou usar workflow separado)
- [ ] **08-5** Testar localmente: `node npm/bin/run.js init --yes --dry-run`
- [ ] **08-6** Publicar v1.1.0 manualmente na primeira vez: `cd npm && npm publish --access public`
- [ ] **08-7** Atualizar README com seção `npx` ao lado da seção Homebrew

## Arquivos gerados / alterados
```
npm/
  package.json
  bin/run.js
  README.md
.github/workflows/npm-publish.yml
README.md                           (seção npx adicionada)
```

## Segurança
- `run.js` valida SHA256 antes de executar o binário baixado
- Binário é armazenado em diretório do usuário, nunca em `node_modules`
- Não executa código de terceiros; apenas o binário oficial do release

## Skills relevantes
(consultar `skills/index.md`)
