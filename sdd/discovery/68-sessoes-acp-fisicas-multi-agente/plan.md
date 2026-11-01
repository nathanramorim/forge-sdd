# Plano de Discovery 68 — Sessões ACP Físicas Multi-Agente

Execução em uma única branch `feat/68-sessoes-acp-fisicas-multi-agente` (`FLX-02`). Pasta de execução: `sdd/discovery/68-sessoes-acp-fisicas-multi-agente/features/`. Rule of 7: ver abaixo. Status: FLX-01 aprovado; fatiada em 8 tasks.

## Ordem e dependências (após fatiamento aprovado)
`68-01 -> 68-06 -> 68-06b`, `68-01 -> 68-02`, `68-02 + 68-06 -> 68-03 -> 68-03b`, `68-03 + 68-06 -> 68-04`, `68-03b + 68-04 -> 68-05`.
Ordem de fila: 68-01, 68-06, 68-02, 68-03, 68-03b, 68-06b, 68-04, 68-05.

| Task | Tam. | Depende | Escopo | Rules/IDs | Skills |
|---|---|---|---|---|---|
| 68-01 | M | — | Bloco `acp.*` em `.sddrc`, presets embutidos (`requires`, `versionArgs`, `install`), `forge-sdd config list/get/set/unset` só `acp.*` | ARQ-01, ARQ-03, ARQ-06, SEG-01, TST-02 | `novo-subcomando-cli-sdd` |
| 68-06 | M | 68-01 | `internal/preflight`: `Check`/`Plan` (LookPath, versão, Node/npx), relatório, uso no doctor e aviso no Studio | SEG-01, SEG-03, DUP-03, TST-04 | — |
| 68-02 | M | 68-01 | `ProcessRunner`: spawn, PID, stdin/stdout, health check (`initialize`/`liveness`), `Exited()` | ARQ-04, SEG-03, SEG-05, COD-03, TST-04 | — |
| 68-03 | M | 68-02, 68-06 | Pool: Warmup físico, falha parcial, `enabledAgents`, `Status` conectadas/total | ARQ-11, DUP-03, COD-01, TST-03 | — |
| 68-03b | M | 68-03 | Restart com backoff exponencial, estado `restarting`, `FailureHistory` | ARQ-11, SEG-05, COD-01, TST-04 | — |
| 68-06b | M | 68-06, 68-01 | Seleção interativa em init/update, instalação com confirmação, `--yes`, grava `enabledAgents` | SEG-01, SEG-03, ARQ-05, TST-04 | — |
| 68-04 | M | 68-03, 68-06 | Studio: boot não bloqueante, chip "x de N conectadas", estação em erro, aviso de preflight, modo leitura | ARQ-10, FLX-04, TST-05 | — |
| 68-05 | M | 68-03b, 68-04 | Painel de falhas: `GET /api/acp/failures` + UI pt-BR | ARQ-03, ARQ-10, SEG-05, COD-05 | — |

Rule of 7: 8 subtarefas atômicas em duas fases lógicas; a Governança pode exigir agrupamento por fase no Portão. Status: FLX-01 aprovado.

## Critério final
`go build ./... && go vet ./... && go test -race ./...`

## Pós-Portão
Fila alimentada em ordem de dependência; Builder por subagentes, Governança (FLX-06). Sem PR até o humano retornar.
