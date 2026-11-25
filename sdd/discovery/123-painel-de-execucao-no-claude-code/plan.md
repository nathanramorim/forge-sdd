# Plano — 123

## Fatias e validação
- **Piloto:** onda 1 entrega status line e alertas, que só leem e não gravam nada. Valida o carregamento do mod no projeto.
- **Validação:** CA-01 a CA-03 e uso real numa iniciativa.
- **Reversão:** remover `.agents/plugins/forge-painel/`; nada no core depende do mod.

## Onda 1 — Contrato, carregamento e leitura (1 dia)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 1.1 | Spike: carregar plugin do projeto sem `/plugin install`, no terminal, no desktop e no `claude -p` | — | SEG-07 |
| 1.2 | ADR-0030: contrato de mod (manifesto, permissões, versões mínimas) | 1.1 | FLX-05, GOV-01 |
| 1.3 | Mod `forge-painel`: status line e alertas de janela | 1.2 | SEG-05, SEG-08 |
| 1.4 | `forge skills instalar plugin` valida o manifesto do mod | 1.2 | SEG-01, SEG-02, TST-03 |

## Onda 2 — Telemetria e agentes (1 a 2 dias)
| # | Task | Depende de | Rules/IDs |
|---|---|---|---|
| 2.1 | Telemetria automática da sessão interativa e subagentes, dono único (mudo com `FORGE_WORKER`) | 1.3 | FLX-04, DUP-04 |
| 2.2 | Painel lateral da esteira: jobs do `forge up` lidos do barramento e dos logs, mais subagentes da sessão | 1.3 | SEG-08, DUP-01 |
| 2.3 | Aviso de orçamento em `agent.spawn` | 1.3 | — |
| 2.4 | `/gastos` por iniciativa e estação | 2.1 | DUP-01 |

Com a 122 em `granularidade: task`, a 2.1 passa a registrar um Builder por task sem mudança.
