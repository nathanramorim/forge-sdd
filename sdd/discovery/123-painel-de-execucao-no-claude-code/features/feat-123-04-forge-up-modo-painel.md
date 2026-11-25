# feat-123-04 — forge up --terminal painel

Task 2.5. Trilha C. Depende de feat-123-02.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/config (TerminalAuto/Tmux/Native), internal/pipeline/launcher.go, cmd/forge-sdd/up.go.

## Decisões do projeto aplicadas
ADR-0026 (tmux opção), RN-13 e RN-14 da spec, ARQ-05.

## O que já existe
Terminais hoje: `auto`, `tmux`, `native`. Estender com `painel`: workers em segundo plano (processo destacado, log em `.forge/runtime/logs/`), nenhuma janela nem tmux; recusa sem o mod instalado e indica `forge skills instalar plugin forge-painel`. `--dry-run` só imprime.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-05`, `ARQ-06`, `SEG-03`, `TST-02`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/ -run Painel ./cmd/forge-sdd/ -run Up
go build ./... && go vet ./...
```
