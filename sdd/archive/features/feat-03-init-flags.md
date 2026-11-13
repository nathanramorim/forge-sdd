# feat/init-flags

**Branch:** `feat/init-flags`
**Fase:** 3
**Depende de:** `feat/init-interactive` (mergeada)
**Status:** `todo`

## Objetivo
`forge-sdd init --yes` cria a árvore com valores default + flags, sem nenhum prompt interativo.

## Critério de conclusão
```bash
mkdir /tmp/test-flags
forge-sdd init --yes --stack go --db postgres /tmp/test-flags
grep -i "go" /tmp/test-flags/sdd/memory/constitution.md
# → deve conter "go" ou "Go", Exit 0, nenhum prompt exibido
```

## Tarefas
- [ ] **03-1** Registrar flags cobra: `--stack`, `--db`, `--telemetry`, `--lang`, `--version`, `--yes`
- [ ] **03-2** Implementar `config.FromFlags(cmd *cobra.Command) Config` preenchendo Config das flags
- [ ] **03-3** Se `--yes`: pular `survey.Run()`, usar `config.Defaults()` sobrescrito pelas flags
- [ ] **03-4** Aceitar diretório alvo como argumento posicional `[targetDir]` (default: `.`)
- [ ] **03-5** Garantir que modo interativo ainda funciona sem `--yes`

## Arquivos gerados
```
internal/config/config.go  (atualizado: FromFlags)
cmd/forge-sdd/main.go      (atualizado: flags registradas + lógica --yes)
```

## Skills relevantes
(consultar `skills/index.md`)
