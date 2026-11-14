# feat-81-03 — Gatilho de Auto-Archive de Features

**Branch:** `feat/81-tokens-por-janela-e-reset-suave`
**Fase:** 81-03
**Tamanho:** `M`
**Depende de:** 81-02
**Status:** `ready` (FLX-01 aprovado)

## Contexto & Motivação
Disparar a rotina de auto-archive de features concluídas quando todas as suas tarefas forem aprovadas e não houver dependências pendentes, compactando a árvore ativa antes do reset de contexto.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica `internal/pipeline/` e `cmd/forge-sdd/archive.go` (`ARQ-01`).

## Decisões do projeto aplicadas
- `ADR-0004`: Pipeline Multi-Sessão.
- `FLX-06`: Auto-archive e governança de contexto.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Comando archive | `cmd/forge-sdd/archive.go` | `archiveCmd` | estender | Adicionar função programática de auto-archive por iniciativa. |
| Monitor da esteira | `internal/pipeline/` | `Watcher` | integrar | Acionar auto-archive após aprovação da última task. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-01`, `FLX-06`, `DUP-02`, `GOV-01`.

## Critérios de Aceitação Executáveis
1. Ao concluir uma iniciativa sem tarefas pendentes, o auto-archive move os artefatos finalizados para a pasta de histórico.
2. Nenhuma alteração manual de arquivos é exigida do usuário.
3. Testes unitários validam a detecção de término e arquivo.
