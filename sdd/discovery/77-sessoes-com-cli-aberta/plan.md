# Plano de Discovery 77 — Sessões com a CLI Aberta

Branch única `feat/77-sessoes-com-cli-aberta` (`FLX-02`), em worktree `../forge-sdd-77`, partindo da `feat/72`. Pasta de execução: `sdd/discovery/77-sessoes-com-cli-aberta/features/`. Depende do **fix-76** (classificador de agente indisponível) e toca `tmux.go` como a Discovery 75: mesclar a 75 antes, ou resolver o conflito na 77-02.

## Ordem e dependências
`77-01 -> 77-02 -> 77-03 -> 77-04 -> 77-05`. Sequencial: todas tocam `internal/pipeline` e o mesmo contrato de configuração.

## Quebra preliminar (5 features)
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 77-01 Modo de sessão e perfis | M | — | `session_mode`, `context`, `live.*`, `reset_command` por perfil; padrão `headless` para config antiga; argumentos interativos de Builder e Reviewer; menu e validação | `sdd-codigo-e-testes`, `sdd-arquitetura`; `COD-01`, `COD-03`, `TST-04`, `ARQ-01` | — |
| 77-02 Painéis ao vivo no tmux | M | 77-01 | `BuildTmuxPlan` em `live` (CLI nos painéis 1 e 2, janela `forge-ctl`), contingência para `headless`, `up --dry-run`, rótulo no menu | `sdd-seguranca`, `sdd-codigo-e-testes`; `SEG-01`, `SEG-03`, `COD-03`, `TST-04` | — |
| 77-03 Despachante de entrega | G | 77-02 | `worker --deliver`: claim e lease, painel ocioso, colagem por buffer, término por estado (`verifyHandoff` compartilhado), timeout, limpeza de contexto | `sdd-seguranca`, `sdd-codigo-e-testes`; `SEG-01`, `SEG-05`, `DUP-01`, `COD-01`, `TST-03`, `TST-04` | — |
| 77-04 Falhas lidas da tela | M | 77-03, fix-76 | `capture-pane` sobre limite de uso, agente indisponível e painel fechado; `Requeue`, pausa, avisos no inbox | `sdd-codigo-e-testes`, `sdd-seguranca`; `SEG-05`, `COD-03`, `TST-03`, `TST-04` | — |
| 77-05 Docs, ADR, E2E e release | M | 77-04 | ADR nova (substitui parcialmente a 0004), guia, `LEIA-ME`, `concepts.md`, release notes, E2E com tmux real e agente falso | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-05` | — |

## Critério final
CA-01 a CA-09 de `criteria.md`; E2E com tmux real: dois jobs seguidos entregues ao agente falso, contexto limpo entre eles, término por estado.

## Medição
Anotar no PR: tempo entre o job entrar na fila e o primeiro texto aparecer no painel (meta: abaixo de 5 s com painel ocioso) e quantas entregas precisaram de reenvio.

## Pós-Portão
Após `FLX-01`: `/split-features` grava as specs e enfileira. Camada B por task, só sobre o diff. Release: bump a decidir junto da 75 e do fix-76; tag na branch só com autorização (Regra 11).
