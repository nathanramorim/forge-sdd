# feat/lp-wiki-sync

**Branch:** `feat/sync-wiki-documentation`
**Fase:** 33
**Depende de:** `feat-32-product-release-notes`
**Status:** `done`

## Objetivo

Automatizar a atualização da documentação wiki localizada no repositório privado `nathanramorim/lp-forge-sdd` (diretório `wiki/` na branch `main`) sempre que uma nova release/tag `v*` do `forge-sdd` for publicada.

## Critério de conclusão

```bash
# 1. O arquivo release.yml deve conter o step com nome contendo "Wiki" ou "Sync"
grep -q -i -E "sync|wiki" .github/workflows/release.yml

# 2. O script do workflow deve referenciar a variável de segredo LP_FORGE_SDD_TOKEN
grep -q "LP_FORGE_SDD_TOKEN" .github/workflows/release.yml
```

## Tarefas

- [x] **33-1** Criar especificação da feature em `sdd/features/feat-33-lp-wiki-sync.md`
- [x] **33-2** Atualizar `sdd/memory/progress.md` e `sdd/features/index.md`
- [x] **33-3** Atualizar o workflow `.github/workflows/release.yml` com o passo de sincronização
- [x] **33-4** Testar localmente a consistência do shell script utilizado
