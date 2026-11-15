# Feature 02-07 — Versão, Documentação e Release `v2.2.0-beta`

Fecha o pacote. Só inicia depois das demais features desta subpasta estarem `done`, seguindo o padrão dos pacotes anteriores (feat-01, discovery-5ae2). Depende de feat-02-01 a feat-02-06.

## Critérios de Aceitação Executáveis

1. **Confirmar com o usuário antes de codar**: o pulo de `2.0.0-beta` (versão atual em `sdd/.sddrc`/`internal/config/config.go`) direto para `2.2.0-beta`, sem passar por `2.1.0-beta`, é definitivo — ou o usuário prefere `2.1.0-beta` seguindo o padrão incremental já usado no projeto.
2. `sdd/.sddrc` e `internal/config/config.go` (`SddVersion`) atualizados para a versão confirmada; `npm/package.json` sincronizado.
3. Release notes em `sdd/releases/history.md` citam `.agent/` (rules + commands), o novo modelo de adaptador fino por agente, e a nova pergunta de branch de partida/retomada — identificando explicitamente o agente/comando afetado (Regra 12 da Constituição).
4. Documentação (README) atualizada citando a nova convenção `.agent/`.
5. `go build`, `go vet ./...` e `go test ./...` passam; golden fixtures finais regeneradas.

## Status: done

`2.2.0-beta` confirmado — pedido explícito do usuário na abertura da discovery-02 ("quero criar ... uma nova entrega da versão beta 2.2.0"), reafirmado ao acionar a execução completa do pacote; não houve necessidade de nova confirmação. `cmd/forge-sdd/main.go` (`var version`), `internal/config/config.go` (`Defaults().SddVersion`) e `npm/package.json` atualizados para `2.2.0-beta`, no mesmo padrão do bump anterior (v2.0.0-beta, commit `3105010`). `sdd/.sdd-version`/`sdd/.sddrc` deste repositório já refletem `2.2.0-beta` via dogfood real (`forge-sdd update --yes --version 2.2.0-beta`). `README.md` e `npm/README.md` ganharam a seção "✨ Agent Rules e Branch por Feature (v2.2.0-beta)", com a versão anterior movida para "Novidades da Versão Anterior". `sdd/releases/history.md` documenta as quatro entregas do pacote. `go build`, `go vet ./...` e `go test ./...` passam; golden fixtures regeneradas (feat-02-01).
