# feat-77-03 — Despachante de entrega

**Branch:** `feat/77-sessoes-com-cli-aberta`
**Fase:** 77-03
**Tamanho:** `G`
**Depende de:** 77-02
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
Implementar a flag `forge-sdd worker --deliver <target>`: consome jobs da fila, aguarda o painel ficar ocioso, entrega o prompt usando colagem segura via buffer do tmux (`load-buffer` + `paste-buffer`), monitora a conclusão do job por estado (`verifyHandoff`) e efetua a limpeza do contexto entre execuções.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Manipulação de processos e buffers tmux, concorrência, laço de leases e handoff de jobs (`ARQ-11`, `COD-01`, `SEG-01`).

## Decisões do projeto aplicadas
- `D-2 (ADR-0007)`: Entrega segura por colagem (`load-buffer -b <nome> -` + `paste-buffer -p -d -b <nome> -t <painel>` + `send-keys -t <painel> Enter`). Sem shell intermediário.
- `D-3 (ADR-0007)`: Painel ocioso medido via `capture-pane -p -t <painel>` estável por `ready_quiet_seconds`.
- `D-4 (ADR-0007)`: Conclusão por estado compartilhando lógica de `verifyHandoff`.
- `D-5 (ADR-0007)`: Limpeza de contexto com `reset_command` quando `context == clear`.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Laço de claim e lease | `internal/pipeline/worker.go` | `RunWorker`, `bus.Claim`, `bus.RenewEvery` | reusar | O despachante mantém o mesmo ciclo de vida e recuperação de leases. |
| Verificação de handoff | `internal/pipeline/worker.go` | `verifyHandoff` | extrair e reusar | Compartilhar lógica de término por estado entre headless e live. |
| Tmux runner | `internal/pipeline/tmux.go` | `TmuxRunner` | estender | Adicionar métodos para `LoadBuffer`, `PasteBuffer`, `SendKeys` e `CapturePane`. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-05`, `DUP-01`, `COD-01`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`, `sdd-reuso-e-duplicacao.md`.

## Escopo
1. Adicionar flag `--deliver <target>` no worker (`cmd/forge-sdd/worker.go`).
2. Implementar detecção de painel ocioso (`isPaneIdle`) usando `capture-pane`.
3. Implementar colagem de prompt segura sem interpolação de shell, com timeout e retry único no Enter.
4. Compartilhar `verifyHandoff` para checagem periódica do estado do PR (Builder) ou veredito (Reviewer).
5. Limpar contexto enviando `reset_command` após o término do job quando `context == clear`.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/pipeline -run Deliver
go test -v ./internal/pipeline -run DeliverEstado
go test -v ./internal/pipeline -run DeliverContexto
```
- Testes com runner falso garantem colagem intacta com aspas/quebras de linha, conclusão por estado e limpeza de contexto.
