# Módulos — forge-sdd

## cmd/forge-sdd
**Responsabilidade:** entrypoint; registra comando `init` no cobra; orquestra survey → scaffold.

- [ ] Parsear flags globais (`--stack`, `--db`, `--telemetry`, `--lang`, `--version`, `--yes`, `--dry-run`)
- [ ] Aceitar diretório alvo como argumento posicional (default: `.`)
- [ ] Se `--yes`: usar `config.Defaults()` + flags, pular survey
- [ ] Chamar `scaffold.Run(config, targetDir)`
- [ ] Imprimir próximos passos ao finalizar

## internal/config
**Responsabilidade:** struct `Config` e lógica de preenchimento.

- [ ] `Config{Project, Stack, DB, Telemetry, Lang, SddVersion, DryRun}`
- [ ] `Defaults()` — valores padrão (`Lang: "pt-BR"`, `SddVersion: "1.1.0"`, etc.)
- [ ] `FromFlags(cmd *cobra.Command) Config`

## internal/survey
**Responsabilidade:** formulário interativo com huh; retorna `Config`.

- [ ] 5 campos: nome do projeto, stack, db, telemetria (bool), idioma
- [ ] Retorna `Config` preenchido
- [ ] Skip automático se `Config.DryRun || --yes`

## internal/scaffold
**Responsabilidade:** renderizar templates e escrever (ou imprimir) a árvore.

- [ ] `Walk() []string` — lista paths de `templates/` via embed.FS
- [ ] `Run(cfg Config, targetDir string) ([]string, error)` — walk + render + write
- [ ] Se `cfg.DryRun`: imprimir path com prefixo `[DRY]`, sem criar arquivos
- [ ] Não sobrescrever arquivo existente (erro com lista de conflitos)
- [ ] Renderizar cada arquivo com `text/template` usando `cfg`

## templates/
**Responsabilidade:** artefatos Forge-SDD embutidos via `//go:embed templates/**`.

- [ ] 1 arquivo `.tmpl` por artefato do §3 da metodologia
- [ ] Variáveis de template: `{{.Project}}`, `{{.Stack}}`, `{{.DB}}`, `{{.Lang}}`, `{{.SddVersion}}`
- [ ] Total esperado: ~30 arquivos
