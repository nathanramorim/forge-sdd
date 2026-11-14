# feat-81-01 — Contabilidade de Tokens por Janela

**Branch:** `feat/81-tokens-por-janela-e-reset-suave`
**Fase:** 81-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `ready` (FLX-01 aprovado)

## Contexto & Motivação
Calcular e exibir o consumo acumulado de tokens por janela (Discovery, Builder, Reviewer) a partir dos registros de telemetria da sessão, enriquecendo o `forge-sdd ps` e os títulos das bordas dos painéis tmux.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica `internal/monitor/ps.go`, `cmd/forge-sdd/ps.go` e `internal/pipeline/tmux.go` (`ARQ-01`).

## Decisões do projeto aplicadas
- `ADR-0003`: Métricas de Tokens e Orçamento.
- `ADR-0006`: Visibilidade nos painéis tmux.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Monitor de processos | `cmd/forge-sdd/ps.go`, `internal/monitor/ps.go` | `ListProcessTree` | estender | Adicionar colunas de contabilidade de tokens por papel. |
| Telemetria de sessão | `sdd/.metrics/` | Arquivos JSON de sessão | reusar | Ler tokens consumidos por sessão. |
| Configuração de tmux | `internal/pipeline/tmux.go` | `BuildTmuxPlan` | estender | Formatar bordas dos painéis com indicador de tokens. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `COD-01`, `TST-04`.

## Critérios de Aceitação Executáveis
1. `forge-sdd ps` exibe os tokens de entrada e saída consumidos por painel.
2. Formatação dos painéis tmux reflete os tokens consumidos quando disponíveis.
3. Testes unitários com mock validam o cálculo e formatação.
