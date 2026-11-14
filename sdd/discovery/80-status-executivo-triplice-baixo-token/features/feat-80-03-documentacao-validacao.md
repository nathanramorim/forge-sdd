# feat-80-03 — Documentação, LEIA-ME e Validação Final

**Branch:** `feat/80-status-executivo-triplice-baixo-token`
**Fase:** 80-03
**Tamanho:** `P`
**Depende de:** 80-01, 80-02
**Status:** `ready` (FLX-01 aprovado)

## Contexto & Motivação
Registrar a documentação do subcomando `status --live` e o padrão tríplice no LEIA-ME e comandos, garantindo 100% de conformidade com os portões de governança.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Documentação e validação de suite (`DUP-10`).

## Decisões do projeto aplicadas
- `DUP-10`: Orçamento de contexto e governança.
- `COD-05`: Clareza e concisão na documentação.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Documentação comandos | `docs/comandos.md` | Guia de comandos CLI | atualizar | Incluir seção para `forge-sdd status --live`. |
| Template LEIA-ME | `.forge/LEIA-ME.md.tmpl` | Resumo de operação do pipeline | atualizar | Documentar a consulta rápida de status de baixo token. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `FLX-05`, `GOV-01`, `TST-05`: Suíte completa 100% verde e 0 violações no portão.

## Critérios de Aceitação Executáveis
1. Documentar `forge-sdd status --live` em `docs/comandos.md` e `.forge/LEIA-ME.md.tmpl`.
2. Execução de `go test ./...` e `forge-sdd gate` sem nenhuma regressão.
