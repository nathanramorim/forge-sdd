# feat-81-02 — Reset Suave In-Place e Card de Handoff

**Branch:** `feat/81-tokens-por-janela-e-reset-suave`
**Fase:** 81-02
**Tamanho:** `M`
**Depende de:** —
**Status:** `ready` (FLX-01 aprovado)

## Contexto & Motivação
Garantir que entre jobs o Builder e o Reviewer não acumulem contexto indefinidamente: envio de comando de reset in-place no próprio terminal e injeção do Card de Handoff Compacto (< 800 tokens) com o resumo da próxima tarefa sem reabrir janelas.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica `internal/pipeline/deliver.go` (`ARQ-01`).

## Decisões do projeto aplicadas
- `ADR-0004`: Pipeline Multi-Sessão.
- `ADR-0008`: Sessões com CLI aberta no tmux.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Despacho live | `internal/pipeline/deliver.go` | `DeliverJob` | estender | Injetar reset in-place e Card de Handoff estruturado. |
| Reset de contexto | `internal/pipeline/tmux.go` | `ClearContext` | reusar | Envio de `/clear` e `/new` para os adaptadores. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `COD-01`, `COD-03`, `TST-04`.

## Critérios de Aceitação Executáveis
1. O reset in-place não fecha nem reabre o processo do terminal tmux.
2. O Card de Handoff gerado possui menos de 800 tokens e contém apenas dados essenciais da task.
3. Testes unitários validam a entrega sem travamento.
