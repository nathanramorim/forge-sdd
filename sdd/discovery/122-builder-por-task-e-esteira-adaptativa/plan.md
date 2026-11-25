# Plano — 122

## Fatias e validação
- **Piloto:** onda 1 entrega o job por task em sequência (paralelo 1), atrás de `builder.granularidade: task`. Já reduz o contexto sem tocar na junção.
- **Validação:** CA-01, CA-03, CA-06 e uma iniciativa pequena real rodada nos dois modos.
- **Reversão:** voltar a chave para `feature`; nada migra dados.

## Onda 1 — Job por task em sequência (1 dia)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 1.1 | ADR-0029: job por task, estados compartilhados com o laço, Camada A por task e parecer por feature | — | FLX-05, GOV-01 |
| 1.2 | Tipo `task_build` no barramento e chave `builder.granularidade` | 1.1 | ARQ-01, SEG-01, DUP-04, TST-03 |
| 1.3 | `dispatch` enfileira uma task por job com o pacote de `forge context`; remove `BuilderHints` de paralelo nesse modo | 1.2 | DUP-03, COD-04 |
| 1.4 | Telemetria por task com uso real | 1.2 | FLX-04, SEG-05 |

## Onda 2 — Ondas paralelas com leque (2 dias)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 2.1 | Despachante: `PlanWaves` + `PodeComecar` + paralelo adaptativo (RN-4, RN-5) | 1.3 | ARQ-02, DUP-01 |
| 2.2 | Leque ligado ao worker: worktree por task, junção ao fim da onda, bloqueio com menu | 2.1 | SEG-02, SEG-03, TST-04 |
| 2.3 | N workers do papel builder na mesma iniciativa (ADR-0024) | 2.1 | ARQ-06 |

## Onda 3 — Revisão em dois níveis e medição (1 a 2 dias)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 3.1 | Camada A por task; review LLM nasce uma vez por feature, após junção | 2.2 | FLX-06, FLX-08 |
| 3.2 | `forge status` mostra tokens e estado por task | 1.4 | COD-01 |
| 3.3 | Medição nos dois modos e decisão do padrão (RN-8) | 3.1 | FLX-04 |

Depende da 115-03 (tier por task) só para o modelo; sem ela, vale o modelo do papel.
