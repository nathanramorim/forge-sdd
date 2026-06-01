# feat/versioning

**Branch:** `feat/versioning`
**Fase:** 5
**Depende de:** `feat/dry-run` (mergeada)
**Status:** `todo`

## Objetivo
`.sdd-version` gerado com a versão Forge-SDD correta; `.sddrc` gerado como JSON válido com configuração do projeto.

## Critério de conclusão
```bash
mkdir /tmp/test-version
forge-sdd init --yes /tmp/test-version
cat /tmp/test-version/sdd/.sdd-version
# → "1.1.0"
python3 -m json.tool /tmp/test-version/sdd/.sddrc
# → JSON válido com campos version, telemetry, lang; Exit 0
```

## Tarefas
- [ ] **05-1** Adicionar `SddVersion string` em `Config`; `Defaults()` retorna `"1.1.0"`
- [ ] **05-2** Embutir versão no binário via `ldflags` em `go build`: `-X main.version=1.1.0`
- [ ] **05-3** Template `sdd/.sdd-version.tmpl` renderiza `{{.SddVersion}}`
- [ ] **05-4** Template `sdd/.sddrc.tmpl` renderiza JSON com `Project`, `SddVersion`, `Telemetry`, `Lang`

## Arquivos gerados
```
internal/config/config.go         (atualizado: SddVersion)
templates/sdd/.sdd-version.tmpl   (atualizado: {{.SddVersion}})
templates/sdd/.sddrc.tmpl         (atualizado: JSON completo)
cmd/forge-sdd/main.go             (atualizado: var version injetada via ldflags)
```

## Skills relevantes
(consultar `skills/index.md`)
