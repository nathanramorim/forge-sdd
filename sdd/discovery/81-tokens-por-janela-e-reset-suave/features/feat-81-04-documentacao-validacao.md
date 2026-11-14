# feat-81-04 — Documentação, LEIA-ME e Validação Final

**Branch:** `feat/81-tokens-por-janela-e-reset-suave`
**Fase:** 81-04
**Tamanho:** `P`
**Depende de:** 81-01, 81-03
**Status:** `ready` (FLX-01 aprovado)

## Contexto & Motivação
Atualizar a documentação técnica, manuais e LEIA-ME para refletir as novidades de monitor de tokens por janela e auto-reset in-place, garantindo validação 100% verde nos portões.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Documentação e validação de suite (`DUP-10`).

## Decisões do projeto aplicadas
- `DUP-10`: Orçamento de contexto e governança.
- `COD-05`: Clareza e concisão na documentação.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Documentação comandos | `docs/comandos.md` | Guia de comandos | atualizar | Incluir novas flags de tokens e auto-reset. |
| Template LEIA-ME | `.forge/LEIA-ME.md.tmpl` | Resumo de operação do pipeline | atualizar | Documentar a mecânica de reset suave in-place. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `FLX-05`, `GOV-01`, `TST-05`.

## Critérios de Aceitação Executáveis
1. Documentar o comportamento de tokens e auto-reset in-place em `docs/comandos.md` e `.forge/LEIA-ME.md.tmpl`.
2. Execução de `go test ./...` e `forge-sdd gate` sem nenhuma regressão.
