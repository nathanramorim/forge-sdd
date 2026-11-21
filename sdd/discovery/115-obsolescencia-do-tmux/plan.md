# Plano — Iniciativa 115 (Obsolescência do tmux)

Ordem: primeiro o levantamento (o que depende de tmux de forma não óbvia), depois a ADR, depois a remoção, e por último a documentação e a nota de release. Até 7 tasks por onda.

## Onda 1 — Decidir e levantar

| # | Task | Depende de | Rules |
|---|---|---|---|
| 1.1 | **ADR-0024: o tmux sai; o chat é o único caminho de orquestração.** Fecha a ADR-0015 como aceita e substitui a ADR-0008. Diz o que sai, o que fica e como é a transição. | — | `FLX-05`, `GOV-01` |
| 1.2 | **Levantamento de acoplamento:** ler `deliver.go`, `worker.go`, `status_live.go` e `preflight.go` e escrever, em `acoplamento.md`, o que quebra sem tmux e o que não depende dele. Fecha as dúvidas do `## A confirmar`. | — | `DUP-01` |

## Onda 2 — Remover

| # | Task | Depende de | Rules |
|---|---|---|---|
| 2.1 | Remover `tmux.go`, `tmuxinstall.go` e os três testes de tmux. Adaptar `preflight.go` e `menu.go` ao modo chat. | 1.2 | `COD-04`, `ARQ-01` |
| 2.2 | Adaptar `deliver.go` e `status_live.go` ao que sobra do levantamento: sem `exec.Command("tmux")`, mantendo as fontes que não são tmux. | 1.2 | `COD-04`, `TST-05` |
| 2.3 | Configuração: `terminal: tmux` recusado com mensagem de transição; `auto` e `native` deixam de existir como escolha se não houver terminal a escolher. Teste negativo `TestTerminalTmuxRecusado`. | 2.1 | `SEG-01`, `COD-03` |

## Onda 3 — Fechar a frente

| # | Task | Depende de | Rules |
|---|---|---|---|
| 3.1 | `forge doctor` deixa de falar de tmux e passa a diagnosticar o que o modo chat precisa (agente de IA configurado, barramento legível). | 2.3 | `ARQ-01` |
| 3.2 | Documentação: README, `docs/`, `docs/comandos.md`, `/tutorial` e a skill `sdd-operacao`. Seção "o que saiu" para quem procura. | 2.3 | `COD-05`, `GOV-04` |
| 3.3 | Nota de release na `history.md` e checklist de versão: a remoção entra numa **beta** antes de qualquer estável (`RN-4`). | 3.2 | `FLX-07` |

## Caminho crítico

`1.2 → 2.1 → 2.3 → 3.1/3.2 → 3.3`

A 1.1 (ADR) pode correr em paralelo com 1.2, mas precisa estar aceita antes do merge da 2.x, porque muda arquitetura (`GOV-01`).

## Observações de risco

- **2.2 é a task de risco:** `deliver.go` pode estar no caminho do modo `worker` headless, que não é tmux. Por isso a 1.2 vem antes: sem o levantamento, a remoção cega quebra o worker.
- **Transição:** a recomendação é a versão de aviso (uma beta com `terminal: tmux` ainda aceito + aviso), e a remoção do campo na beta seguinte. Confirmar no Gate 1.