# Fix 115-01 — Encerrar workers ao drenar a iniciativa

## Problema

`forge up` abre Builder e Reviewer que ficam vivos consultando o barramento compartilhado. A tela de status mostra a fila da iniciativa da branch atual, mas `worker` reivindica jobs de qualquer iniciativa no mesmo projeto. Assim, a fila visível pode estar vazia enquanto um Builder trabalha em outra iniciativa ou fica ocioso aguardando jobs futuros.

Na reprodução de 21/11/2026, a branch atual era `feat/115-02-remove-tmux` e sua fila estava vazia. `forge ps --json` mostrou o worker do Builder executando `goal-feat-97-02-97-finalizar-implementacoes-pendentes`, com mais jobs de outras iniciativas aguardando. O usuário decidiu que cada worker deve ficar limitado à iniciativa em que foi aberto e encerrar quando o trabalho dessa iniciativa terminar. O job já ativo deve concluir; não interromper jobs em andamento.

## Feature Correlacionada

- **Origem:** iniciativa 115, `sdd/discovery/115-obsolescencia-do-tmux/plan.md`, tasks 1.2 e 2.2, que levantam o acoplamento de `worker.go` e adaptam o pipeline após a saída do tmux.
- **Causa-raiz:** `internal/pipeline/worker.go` continua em loop quando `WorkerOptions.Once` é falso; `cmd/forge-sdd/worker.go` define `--once=false` por padrão. `forge up` inicia workers persistentes e `bus.Claim` não recebe filtro de iniciativa.
- **Evidência:** status da iniciativa 115 sem tasks enfileiradas; `forge ps --json` identificou um job da iniciativa 97 em execução no Builder. O barramento do projeto também continha jobs pendentes das iniciativas 88, 97, 98, 99 e 100.

## Modelo do módulo

**Enxuto** (`ARQ-11`): estender os contratos existentes em `internal/pipeline`, `internal/bus` e `internal/layout`; `cmd/` adapta flags e encaminha a iniciativa resolvida. Não criar pacote nem dependência externa.

## Decisões do projeto aplicadas

- **ADR-0002 — Layout e Estado por Iniciativa:** `internal/layout` é dono da descoberta da iniciativa atual pela branch e do estado da fila.
- **ADR-0004 — Pipeline Multi-Sessão:** o worker Go reivindica jobs do barramento com lease; Builder e Reviewer mantêm papéis distintos.
- **ADR-0016 — Lock por Iniciativa e Estado do Laço:** o estado e o ciclo de vida do trabalho são isolados por iniciativa.
- **ADR-0023 — Estado isolado:** o estado de execução não deve misturar o trabalho de features/iniciativas paralelas.
- **Decisão do dono neste fix:** um worker acompanha apenas a iniciativa selecionada; jobs de outras iniciativas permanecem intactos. O worker termina depois de concluir o job atual e quando não há mais jobs pendentes ou em processamento em nenhum dos dois papéis daquela iniciativa, preservando a possibilidade de rework do Reviewer.
- **Resolução proposta:** `forge up` deriva a iniciativa da branch atual e a passa a Builder e Reviewer. `forge up` e `forge worker` aceitam `--initiative` para seleção explícita; sem iniciativa resolvível, falham com orientação em vez de usar a fila global.
- **Registro arquitetural:** documentar o escopo e o ciclo de vida do worker na ADR-0024 planejada pela iniciativa 115 antes da implementação (`FLX-05`, `GOV-01`); não é necessária mudança na Constitution.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Ciclo de vida do worker | `internal/pipeline/worker.go` | `RunWorker` repete claim e polling; `Once` encerra só quando a fila consultada está vazia | **adaptar** para escopo de iniciativa e encerramento após drenar a iniciativa inteira | Evitar workers presos aguardando jobs de outras branches |
| Claim de jobs | `internal/bus` | `Claim` reivindica o próximo job por papel no barramento do projeto | **estender** com filtro de iniciativa sem alterar leases ou atomicidade | Não mover nem reivindicar jobs alheios |
| Inicialização de sessões | `cmd/forge-sdd/up.go`, `internal/pipeline/launcher.go` | O launcher inicia workers sem identidade de iniciativa | **adaptar** para passar a iniciativa da branch ou flag explícita | `forge up` deve abrir workers do contexto atual |
| Resolução da iniciativa | `internal/layout` | `Current` e `FromBranch` resolvem o contexto da branch | **reutilizar** | Um único dono para o conceito, conforme `DUP-01` |
| Status e monitoramento | `internal/pipeline/status_live.go`, `internal/monitor/ps.go` | O status da fila é por iniciativa; os processos são do projeto | **alinhar** a exibição ao escopo do worker | Explicar por que um worker permanece ativo e qual iniciativa atende |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-02`, `ARQ-06`, `ARQ-11`, `COD-03`, `DUP-01`, `FLX-01`, `FLX-02`, `FLX-04`, `FLX-05`, `GOV-01`, `SEG-01`, `TST-01`, `TST-03`, `TST-05`.

Rules: `sdd-arquitetura`, `sdd-fluxo-e-portao`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis

1. **Isolamento:** teste do barramento prova que um worker com iniciativa `115` reivindica somente jobs `115`; jobs de `97`, `98` e outras iniciativas continuam no mesmo estado.
2. **Encerramento:** teste do worker prova que ele sai após concluir o job atual e não haver jobs pendentes ou em processamento da iniciativa alvo em Builder ou Reviewer.
3. **Ciclo de revisão:** teste prova que o Builder permanece aguardando enquanto há job da mesma iniciativa em processamento pelo Reviewer e volta a reivindicar um rework dessa iniciativa.
4. **Contexto de `up`:** `forge up --dry-run` numa branch de iniciativa passa o ID aos workers; quando não houver iniciativa resolvível, o comando pede uma seleção explícita ou falha com orientação, sem cair no barramento global.
5. **Job ativo:** teste verifica que o worker conclui o job que já reivindicou antes de sair; a mudança de escopo não cancela trabalho em andamento.
6. **Verificação:** `go vet ./...` e `go test ./...` terminam com código `0`.
