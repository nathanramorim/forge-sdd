# Plano — 122

## Fatias e validação
- **Piloto:** onda 1 entrega job por task com worktree efêmero em sequência (1 par), atrás de `builder.granularidade: task`. Já reduz contexto e disco sem paralelo.
- **Validação:** CA-01, CA-03, CA-04, CA-08 e uma iniciativa pequena real nos dois modos.
- **Reversão:** voltar a chave para `feature`; nenhum dado migra.

## Onda 1 — Job por task, worktree efêmero, em sequência (1 dia)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 1.1 | ADR-0029: job por task, worktree efêmero, pares Builder e Reviewer, junção contínua, teto por disco | — | FLX-05, GOV-01 |
| 1.2 | Tipo `task_build` no barramento e chaves `builder.granularidade` e `builder.limite_disco` | 1.1 | ARQ-01, SEG-01, DUP-04, TST-03 |
| 1.3 | `dispatch` enfileira uma task por job com o pacote de `forge context`; remove `BuilderHints` de paralelo nesse modo | 1.2 | DUP-03, COD-04 |
| 1.4 | Worktree efêmero: cria, commit, apaga; órfão vira menu | 1.2 | SEG-02, SEG-03, TST-04 |
| 1.5 | Telemetria por task e papel com uso real | 1.2 | FLX-04, SEG-05 |
| 1.6 | Contrato compacto entre agentes (prompt, handoff, retorno, parecer) e saída em tópicos para o humano gerada pelo Go | 1.2 | DUP-04, SEG-08, COD-01 |

## Onda 2 — Pares em paralelo com teto por disco (2 dias)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 2.1 | Medição de disco por sistema (stdlib) e tamanho do worktree guardado por máquina | 1.4 | ARQ-04 |
| 2.2 | Despachante: vagas por disco, largura da onda, janela de uso; recalcula a cada vaga liberada | 2.1 | ARQ-02, DUP-01 |
| 2.3 | Reviewer por Builder ativo, lendo só o diff da task | 1.3 | FLX-06, FLX-08, SEG-06 |
| 2.4 | Junção contínua na branch da feature, em série, na ordem do plano | 2.3 | SEG-03, TST-03 |

## Onda 3 — Status e medição (1 a 2 dias)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 3.1 | `forge status` mostra pares ativos, vagas de disco e tokens por task | 2.2 | COD-01 |
| 3.2 | PR em draft da feature ao fim das tasks; squash opcional no merge | 2.4 | FLX-03 |
| 3.4 | Retomada após queda: lease vencido, worktree sujo vira retomar, handoff compacto, checkpoint `wip` (RN-16 a RN-19) | 2.4 | SEG-02, TST-03 |
| 3.3 | Medição nos dois modos e decisão do padrão (RN-12) | 3.1 | FLX-04 |

Depende da 115-03 (tier por task) só para o modelo; sem ela, vale o modelo do papel.
