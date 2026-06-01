# feat/dry-run

**Branch:** `feat/dry-run`
**Fase:** 4
**Depende de:** `feat/init-flags` (mergeada)
**Status:** `todo`

## Objetivo
`forge-sdd init --dry-run` imprime a árvore que seria criada sem gravar nenhum arquivo no disco.

## Critério de conclusão
```bash
cd /tmp && forge-sdd init --dry-run --yes
# → imprime lista de arquivos prefixados com [DRY], Exit 0
ls /tmp/sdd 2>/dev/null | wc -l
# → deve ser 0 (nenhum arquivo criado)
go test ./internal/scaffold/... -run TestDryRunNoFiles -v
# → Exit 0
```

## Tarefas
- [ ] **04-1** Adicionar `DryRun bool` em `internal/config/Config`
- [ ] **04-2** Registrar flag `--dry-run` no cobra
- [ ] **04-3** Em `scaffold.Run`: se `cfg.DryRun`, imprimir `[DRY] <destPath>` em vez de criar arquivo
- [ ] **04-4** Escrever `TestDryRunNoFiles` em `scaffold_test.go` verificando que tmpdir permanece vazio

## Arquivos gerados
```
internal/config/config.go            (atualizado: DryRun field)
internal/scaffold/scaffold.go        (atualizado: branch DryRun)
internal/scaffold/scaffold_test.go   (atualizado: TestDryRunNoFiles)
cmd/forge-sdd/main.go                (atualizado: flag --dry-run)
```

## Skills relevantes
(consultar `skills/index.md`)
