# Plano de Discovery 81 — Monitor de Tokens por Janela e Reset Suave de Sessão

Branch única `feat/81-tokens-por-janela-e-reset-suave` (`FLX-02`), partindo de `feat/80-status-executivo-triplice-baixo-token`. Pasta de execução: `sdd/discovery/81-tokens-por-janela-e-reset-suave/features/`.

## Ordem e dependências
- **Onda 1 (Paralelo):** `81-01` e `81-02`
- **Onda 2:** `81-03` (depende de 81-02)
- **Onda 3:** `81-04` (fechamento)

## Quebra de Features
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 81-01 Contabilidade de Tokens por Janela | M | — | Em `internal/monitor/` e `cmd/forge-sdd/ps.go`, calcular tokens de entrada/saída por papel (Discovery, Builder, Reviewer) a partir da telemetria de sessão. Atualizar `BuildTmuxPlan` para exibir badge de tokens nas bordas dos painéis e rodapé. | `sdd-arquitetura`, `sdd-codigo-e-testes`; `ARQ-01`, `ARQ-04`, `COD-01`, `TST-04` | — |
| 81-02 Reset Suave In-Place e Card de Handoff | M | — | Em `internal/pipeline/deliver.go`, implementar o fluxo in-place: envio do comando de reset (`/clear` ou `/new`), verificação de tela ociosa e injeção automática do Card de Handoff Compacto (< 800 tokens) para a próxima task, sem fechar nem reabrir janelas. | `sdd-seguranca`, `sdd-codigo-e-testes`; `SEG-01`, `COD-01`, `COD-03`, `TST-04` | — |
| 81-03 Gatilho de Auto-Archive de Features | M | 81-02 | Integrar o término de tasks aprovadas com a rotina de auto-archive de artefatos finalizados para manter o repositório e o contexto enxutos antes do reset. | `sdd-fluxo-e-portao`, `sdd-reuso-e-duplicacao`; `FLX-01`, `FLX-06`, `DUP-02`, `GOV-01` | — |
| 81-04 Documentação, LEIA-ME e Validação Final | P | 81-01, 81-03 | Atualizar `.forge/LEIA-ME.md.tmpl`, documentar o ciclo de auto-reset e métricas de tokens; rodar suíte completa (`go test ./...`, `go vet`, `gate`). | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `GOV-01`, `TST-05` | — |

## Critério Final
Aprovação dos critérios CA-01 a CA-06 de `criteria.md`. Verificação de contagem de tokens por painel e transição transparente de tarefas sem fechamento de terminal.

## Pós-Portão
Submissão para aprovação humana (Gate 1 - `FLX-01`). Após aprovação, executar `/split-features` para gerar as tasks em `features/` e entregar ao Builder.
