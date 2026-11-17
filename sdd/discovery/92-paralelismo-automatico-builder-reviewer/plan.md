# Plano — Iniciativa 92 (de baixo para cima)

**Pré-requisito:** 91-07 (um job por task, na branch e no PR da iniciativa). Se a 92 for construída antes da 91, a 91-07 passa a ser a primeira task daqui.

## Onda 1: Fundação
- **92-01 ADR-0017** — paralelismo decidido pelo Go: vagas num worker só, cópia por job, integração só pelo worker, revisão por task; estende as ADRs 0004, 0006 e 0014; registra que as branches `forge/…` são locais e temporárias (Regra 15). Rules: FLX-05, GOV-01. Deps: —. CA: arquivo presente + `forge gate`.
- **92-02 Raiz do projeto vista de uma cópia** — `internal/layout` resolve configuração, fila e logs do projeto principal a partir de um worktree. Rules: DUP-01, DUP-04, SEG-01, TST-04. Deps: —. CA: CA-01.
- **92-03 `**Toca:**` e plano de execução** — leitor do cabeçalho, checagem antes do build, plano com motivo de espera em `dispatch --dry-run` e `bus plan`, regra na skill do split. Rules: DUP-01, DUP-03, SEG-01, TST-03. Deps: —. CA: CA-02, CA-03.
- **92-04 Reivindicação elegível** — `bus.Claim` com condição (dependência integrada, arquivo em comum, retrabalho primeiro), reusando o resolvedor de dependências da fila. Rules: DUP-01, ARQ-02, TST-03. Deps: 92-03. CA: CA-04.

## Onda 2: Builder em paralelo
- **92-05 Cópia do job e entrega** — ciclo de vida do worktree pelo Go (criar, retomar, remover com segurança, `list` e `prune`), agente rodando na cópia, guardas da entrega em `forge pipeline pr`. Rules: SEG-01, SEG-02, SEG-03, COD-04, TST-04. Deps: 92-02. CA: CA-05 e as recusas da CA-06.
- **92-06 Integração pelo worker** — merge um por vez na branch da iniciativa, Camada A, push, PR único, registro e revisão enfileirada; conflito vira retrabalho; adiamento com aviso quando o checkout principal impede. Rules: FLX-02, FLX-06, SEG-02, SEG-03, COD-03, TST-03. Deps: 92-05. CA: CA-06.
- **92-07 Vagas do Builder e freio** — worker com N vagas, lease e tetos por vaga, freio por limite de uso e por conflito, uma vaga para `live` e `/toca-ai`. Rules: ARQ-01, COD-01, COD-03, TST-04. Deps: 92-04, 92-06. CA: CA-07, CA-10.

## Onda 3: Reviewer e acabamento
- **92-08 Revisão por task em paralelo** — intervalo de commits no job de revisão, cópia destacada, vagas do Reviewer (`reviewer.max_parallel_reviews`), aviso único de PR pronto depois da Camada A final. Rules: SEG-06, FLX-06, DUP-01, TST-03. Deps: 92-06. CA: CA-08, CA-09.
- **92-09 Painel, Settings, relatório e instruções** — vagas no `forge status` e nas linhas do worker, chaves no Settings, tempo de relógio e conflitos no relatório, `worker-job.md`, `BuilderHints` e `pr-review.md` sem paralelismo pelo agente, docs. Rules: ARQ-01, DUP-10, COD-04, TST-02. Deps: 92-07, 92-08. CA: CA-11, CA-12, CA-13.

## Fechamento (Líder)
CA-14; uma iniciativa real com 3 vagas e a comparação com a 90 no `status.md`.

Paralelismo: na onda 1, 92-01, 92-02 e 92-03 tocam arquivos diferentes (docs, `internal/layout`, `internal/queue`). Na onda 2 as três são em sequência (cada uma usa a anterior). Na onda 3, 92-08 e 92-09 se cruzam em `internal/pipeline`: 92-09 depois de 92-08.

Ordem de entrega recomendada: beta com a 90, depois a 91 e então esta. A própria 92 será construída em fila, pela esteira de hoje.
