# Plano — 111 (painel de métricas e aprendizado; módulo da 109)

Ordem de baixo para cima. Até 7 tasks por onda. As tasks 111-xx rodam no repositório do módulo (`forge-sdd-metricas`); só a onda 4 toca o core. Cada task traz teste negativo (`TST-03`), sem rede (`TST-04`).

## Onda 0 — Bootstrap
| # | Task | Depende de | Rules |
|---|---|---|---|
| P-00 | Repositório privado, diagram-studio, readme-studio (plano ao dono antes de gravar), forge-sdd estável mais recente, `forge-sdd init`, `modulo.json` | decisão 5 | `FLX-01`, `RN-8 da 110` |

## Onda 1 — Primeira entrega: Visão geral e Erros e retornos (dados que já existem)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 111-01 | Esquema `forge-metricas/1` (JSON Schema) e tipos; teste de paridade com a 110 | P-00 | 111-02 | `DUP-01`, `DUP-04` |
| 111-02 | Leitores: sessões (`sdd/.metrics`), vereditos (`verdicts/*.json`), fila (`queue.json`), com contagem de linha ruim | 111-01 | com 111-01 | `SEG-01`, `COD-03` |
| 111-03 | Agregação: retornos por regra, papel e desfecho; medianas com `n` e rótulo | 111-02 | — | `TST-03` |
| 111-04 | `forge metricas gerar`: grava `data/*.js` e `data/*.json`, idempotente, gravação segura | 111-03 | com 111-05 | `SEG-02`, `ARQ-05` |
| 111-05 | `index.html` e `assets/` (landmarks, tokens da LP, tabela com ordenar, buscar, filtrar, paginar) | P-00 | com 111-04 | acessibilidade, `TST-03` |
| 111-06 | Abas Visão geral e Erros e retornos ligadas aos dados; abas restantes com "sem registro" | 111-04, 111-05 | — | `TST-03` |
| 111-07 | Testes estáticos de HTML, CSS, JS (sem rede, sem http externo, contraste, paridade da lista); `forge metricas abrir` | 111-06 | — | `TST-04` |

## Onda 2 — Depreciados e Tokens
| # | Task | Depende de | Rules |
|---|---|---|---|
| 111-08 | Leitor do `depreciados.csv` (105-07) e do catálogo; aba Depreciados | **105-07 mesclada**, 111-02 | `SEG-01` |
| 111-09 | Economia estimada com método e faixa, rótulo `estimado` | 111-08 | `DUP-01` |
| 111-10 | Aba Tokens: por papel, agente, modelo, ferramenta, iniciativa; mediana, faixa, `n`; estimado x real da 110 | 111-02, **110** | `TST-03` |
| 111-11 | Adaptador de `state/<sid>.json` com versão de layout | **106/0.7** | `SEG-01` |
| 111-12 | Capacidade `doctor` e `forge metricas resumo` | **109** | `ARQ-01` |

## Onda 3 — Lições (aprender com os erros)
| # | Task | Depende de | Rules |
|---|---|---|---|
| 111-13 | Leitor de `achados.csv` (105-08) e dos 19 achados históricos | **105-08**, 111-03 | `SEG-01` |
| 111-14 | Reincidência (N = 3; 2 iniciativas; tentativa ≥ 2) | 111-13 | `TST-03` |
| 111-15 | Gerador determinístico de proposta de lição com filtro de afrouxamento | 111-14 | `FLX-08` |
| 111-16 | Aba Lições e menu de aprovação (aprovar, ajustar, recusar, adiar); gravação pelo core | 111-15, **106/RN-8** | `GOV-01`, `RN-8` |
| 111-17 | Medição de efeito antes e depois, reversão | 111-16 | `TST-03` |

## Onda 4 — Core (mínimo, aditivo, só com módulo ativo)
| # | Task | Depende de | Rules |
|---|---|---|---|
| 111-C1 | Evento `sessao.registrada` e campo opcional `tool` (com a 110) | **109** | `ARQ-02`, `TST-02` |
| 111-C2 | Evento `volta.registrada` (Discovery, nova-feature, novo-fix) | 109 | `FLX-06`, `TST-02` |
| 111-C3 | Suíte de byte-identidade estendida: sem módulo, saída igual | 109-01 | `TST-02` |

## Paralelismo e riscos de ordem
Ondas 1 e 0 não dependem de 109, 110, 105-07/08 nem 106 (comando solto lendo arquivos). Onda 2 espera 105-07; Onda 3, 105-08. Onda 4 só depois da 109.
