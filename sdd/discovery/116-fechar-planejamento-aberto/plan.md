# Plano — 116 fechar o planejamento aberto

Objetivo: levar ao fim as specs e a integração que ficaram abertas em 22/11/2026. Quem decide o que sobe para a `main` é a pessoa; as etapas só preparam, conferem e entregam.

| # | Etapa | **Depende de:** | Rules |
|---|---|---|---|
| 1 | Reviewer confere a spec da 115 (#211), incluindo a feat-115-04 | — | FLX-01, FLX-03 |
| 2 | Reviewer confere a spec da 113 (#175) com `forge queue card` | — | FLX-01, ARQ-06 |
| 3 | Spec da 98 (#214): separar spec de entrega e reconferir | — | FLX-01, FLX-03, TST-01 |
| 4 | Integrar a 5.1 na main (#226), com merge commit | — | FLX-07 |
| 5 | Gate 1 das specs 113, 115 e 98 e `/split-features` | 1, 2, 3 | FLX-01 |
| 6 | Builder implementa 115-04, 113 e 98 em uma branch por feature | 5 | FLX-02, TST-03 |

Etapas 1 a 4 são independentes. A 4 e o merge das specs exigem consentimento (`requisitos.json`).
