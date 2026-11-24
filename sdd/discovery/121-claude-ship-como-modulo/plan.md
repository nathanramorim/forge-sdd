# Plano: 121 Claude Ship como módulo opcional

Pré-requisito duro: 109 até a 109-05 (execução de módulo) e 109-12 (ADR). `pendentes` depende da 107-13. Nenhuma task roda no core além da P-01.

## Onda 0 (dono, fora da esteira)
| Task | O quê | IDs | Depende |
|---|---|---|---|
| P-00 | Dono cria `forge-sdd-ship` privado e instala o forge-sdd estável com a esteira (RN-10 e RN-11 da 109) | FLX-02 | Gate 1 |
| P-01 | Core: linha `ship` em "Outros candidatos a módulo" da 109 (só texto) | DUP-01 | Gate 1 |

## Onda 1 (módulo, base)
| Task | O quê | IDs | Depende |
|---|---|---|---|
| 121-01 | `modulo.json`, `internal/contrato` (envelope, códigos de saída) e `gh` falso | ARQ-01, SEG-01, TST-04 | P-00 |
| 121-02 | `internal/ghcli`: leitura de checks, labels e comentários com argumentos explícitos e saída saneada | SEG-03, SEG-08 | 121-01 |
| 121-03 | `forge ship init` e leitura de `sdd/modulos/ship.json` (padrão `observar`) | SEG-01 | 121-01 |

## Onda 2 (módulo, comportamento)
| Task | O quê | IDs | Depende |
|---|---|---|---|
| 121-04 | `internal/janela`: regra do Steward (retrabalho em draft) e `@claude stop` antes do `[Continuar]` | FLX-06 | 121-02, 121-03 |
| 121-05 | `internal/evento`: roteamento dos 5 eventos, idempotente, nunca bloqueia | TST-03 | 121-04, 109-05 |
| 121-06 | `internal/politica`: proposta de `approvals.yml` a partir da Constituição, sempre `observe` | GOV-01, SEG-02 | 121-03 |

## Onda 3 (módulo, operação e docs)
| Task | O quê | IDs | Depende |
|---|---|---|---|
| 121-07 | capacidades `doctor` (Ship no repositório sem módulo, Steward sem limite, `/ship` ligado) e `pendentes` | SEG-05 | 121-05, 107-13 |
| 121-08 | `docs/migracao.md` (M0 a M5) e `docs/governanca.md` | COD-05 | 121-06 |
| 121-09 | prova ponta a ponta num repositório de teste, em `observar` (tarefa do dono) | TST-01 | 121-07 |

Fatias verticais (120): a onda 1 sozinha já entrega `forge ship doctor` em modo leitura, útil para o caso M2; reverter é tirar `ship` de `Módulos permitidos`.
