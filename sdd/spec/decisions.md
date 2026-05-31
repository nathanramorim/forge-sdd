# Decisões — forge-sdd

## Resolvidas

| # | Decisão | Resolução | Motivo |
|---|---------|-----------|--------|
| D1 | Linguagem do CLI | Go | Binário estático, sem runtime externo |
| D2 | Framework CLI | cobra v1.8 | Padrão Go, flags + help automático |
| D3 | Prompts interativos | huh v0.3 (charmbracelet) | Sem CGO, TUI moderno, sem deps pesadas |
| D4 | Render de templates | text/template (stdlib) | Zero dependência, suficiente para o caso de uso |
| D5 | Embed de templates | embed.FS (stdlib) | Binário único, sem assets externos em runtime |
| D6 | Não sobrescrever arquivos | Erro com lista de conflitos | Evitar perda de dados silenciosa |
| D7 | Subcomandos de runtime | Nenhum (chatmodes Copilot) | CLI mínimo, sem manutenção dupla |
| D8 | Idioma padrão | pt-BR | Público-alvo primário; `--lang en` para inglês |
| D9 | Distribuição | goreleaser v2 + GitHub Releases | Multi-OS automatizado, zero manual |

## Abertas

| # | Questão |
|---|---------|
| D10 | [ ] Suportar `--lang en` com templates em inglês? Duplicar templates ou sistema de i18n? |
| D11 | [ ] `forge-sdd init` em diretório não-vazio: bloquear tudo ou listar conflitos e continuar? |
| D12 | [ ] Embutir versão da metodologia no binário via `ldflags` ou constante em `internal/config`? |
