# Plano — Iniciativa 115 (pipeline e roteamento de modelos)

A prioridade é manter o pipeline de agentes e atribuir o modelo certo por complexidade. O tmux é opcional e não bloqueia estas ondas. Limite de 7 tasks por onda.

## Onda 1 — Contratos

| # | Task | Depende de | Rules |
|---|---|---|---|
| 1.1 | ADR-0024 registra workers por iniciativa, política de complexidade/modelo e tmux como opção. | — | FLX-05, GOV-01 |
| 1.2 | Fix 115-01: worker filtra jobs por iniciativa, preserva review/rework e encerra ao drenar. | 1.1 | ARQ-01, DUP-01, TST-03 |
| 1.3 | feat-115-03-01: contratos de tier por task e configuração validada, mantendo feat-90-07 como fallback. | 1.1 | ARQ-11, SEG-01, DUP-01 |

## Onda 2 — Despacho de modelos

| # | Task | Depende de | Rules |
|---|---|---|---|
| 2.1 | feat-115-03-02: dispatch propaga tier e worker resolve modelo pela CLI antes do fallback por papel; mantém Reviewer após Camada A/PR. | 1.3 | FLX-01, SEG-01, SEG-03, TST-03, TST-05 |

## Onda 3 — Orquestração e validação

| # | Task | Depende de | Rules |
|---|---|---|---|
| 3.1 | feat-115-03-03: atualizar skills comuns com rubrica, classificação por task e Advisor condicional. | 1.3 | COD-01, COD-05, DUP-01, FLX-01 |
| 3.2 | Validar roteamento multi-CLI e handoff Builder→Reviewer após Camada A; manter tmux opcional. | 2.1, 3.1 | FLX-03, FLX-05, TST-03, TST-05 |
| 3.3 | feat-115-04: o tmux volta como opção (a 115-02, #224, o removeu do código); o padrão segue em janelas do sistema. | 3.2 | ARQ-11, DUP-01, TST-03, TST-05 |

## Caminho crítico

1.1 → (1.2 || 1.3) → 2.1 → 3.1 → 3.2 → 3.3

A 115-02 (#224) removeu o tmux do código; a decisão do dono é que ele continua opcional, e a feat-115-04 o devolve como opção.
