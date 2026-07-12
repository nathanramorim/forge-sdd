# Feature 5ae2-01 — Cheat-Sheet de Comandos SDD no `init`

Fecha a fratura entre o CLI Go (que só expõe `init/doctor/destroy/update`) e os 11 slash-commands que hoje só existem como arquivos escondidos dentro de `.claude/`, `.gemini/`, `.github/`.

## Critérios de Aceitação Executáveis

1. Ao final de uma execução bem-sucedida de `forge-sdd init` (não em `--dry-run`), o binário deve imprimir no terminal a lista ordenada dos 11 comandos SDD (`/constitution`, `/discovery`, `/split-features`, `/nova-feature`, `/proxima-feature`, `/revisar`, `/status`, `/doctor`, `/archive`, `/upgrade-sdd`, `/c4-architecture`), cada um com uma linha de descrição.
2. A descrição de cada comando deve ser extraída dos próprios arquivos `.prompt.md.tmpl` já embutidos via `embed.FS` (linha "Uso:" ou primeira linha descritiva) — sem duplicar a fonte da verdade em uma nova estrutura de dados.
3. `go build` e `go vet ./...` devem passar; deve existir teste cobrindo a saída do cheat-sheet em `cmd/forge-sdd/commands_test.go` ou equivalente.
