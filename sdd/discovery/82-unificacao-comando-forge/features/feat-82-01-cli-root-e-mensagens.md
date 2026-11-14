# feat-82-01 — Padronização no CLI Go e Monitor

**Branch:** `feat/82-unificacao-comando-forge`
**Fase:** 82-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `complete`

## Contexto & Motivação
A CLI Forge-SDD hoje ainda exibe o nome `forge-sdd` no comando raiz Cobra (`rootCmd.Use = "forge-sdd"` em `cmd/forge-sdd/main.go`), além de mensagens hardcoded em dicas e avisos (como no monitor de processos `internal/monitor/ps.go`). Esta task padroniza o comando raiz para `forge` (`rootCmd.Use = "forge"`), ajusta as dicas do CLI (`forge ps --kill-zombies`) e atualiza a suíte de testes do CLI e monitor.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica comandos centrais da CLI Go e do pacote de monitoramento (`ARQ-01`, `COD-01`).

## Decisões do projeto aplicadas
- `ADR-0005`: Menu Navegável e Agente por Sessão (`forge` sem argumentos como entrada).
- `ADR-0006`: Visibilidade e monitoramento canônico via `forge`.
- `D1 / D2`: Go + Cobra como CLI central.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Comando raiz Cobra | `cmd/forge-sdd/main.go` | `var rootCmd = &cobra.Command{ Use: "forge-sdd", ... }` | estender | Alterar para `Use: "forge"` para que help e usage usem `forge`. |
| Dicas do Monitor | `internal/monitor/ps.go` | `[Dica: Use 'forge-sdd ps --kill-zombies'...]` | estender | Padronizar string para `[Dica: Use 'forge ps --kill-zombies'...]`. |
| Testes de Monitor | `internal/monitor/ps_test.go` | Asserções de string formatada | estender | Ajustar asserções para o novo padrão. |
| Testes do CLI | `cmd/forge-sdd/` | Testes do rootCmd | estender | Adicionar teste validando `rootCmd.Use == "forge"`. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`: cmd/ apenas adapta CLI, regras em internal/.
- `COD-01`, `COD-03`: Nomes e mensagens claras e coesas.
- `DUP-01`: Um conceito, uma casa: o comando canônico é `forge`.
- `TST-01`: Testes unitários cobrindo o novo usage e formato de mensagens.

## Critérios de Aceitação Executáveis
- [x] 1. `forge --help` (ou `rootCmd.UsageString()`) informa `Usage: forge [flags]` e `forge [command]`.
- [x] 2. `internal/monitor/ps.go` exibe a dica `[Dica: Use 'forge ps --kill-zombies' para limpar processos orfãos ou presos]`.
- [x] 3. Testes unitários em `cmd/forge-sdd` e `internal/monitor` passam com 100% de sucesso (`go test -v ./cmd/forge-sdd/... ./internal/monitor/...`).
