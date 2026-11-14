# feat-80-02 — Subcomando nativo forge-sdd status --live

**Branch:** `feat/80-status-executivo-triplice-baixo-token`
**Fase:** 80-02
**Tamanho:** `M`
**Depende de:** —
**Status:** `done` (implementado e validado)

## Contexto & Motivação
Disponibilizar o comando `forge-sdd status --live` para inspecionar programaticamente a tríplice (painel tmux ou log de runtime, fila de tarefas e commits recentes git) em uma única invocação rápida (< 100ms).

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Adiciona flag e execução em `cmd/forge-sdd/status.go` e `internal/pipeline` (`ARQ-01`).

## Decisões do projeto aplicadas
- `ARQ-01`: Arquitetura Go e separação cmd vs internal.
- `ADR-0006`: Visibilidade tmux e leitura de buffers.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Comando status | `cmd/forge-sdd/status.go` | `statusCmd` | estender | Adicionar flag `--live` / `-l`. |
| Leitura painel tmux | `internal/pipeline/tmux.go` | `TmuxRunner` | reusar | Extrair últimas linhas da sessão. |
| Painel da fila | `cmd/forge-sdd/queue.go` | `queue panel` | reusar | Inspecionar status da fila em tempo real. |
| Histórico git | stdlib `os/exec` | `git log` | reusar | Obter commits recentes. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `COD-01`, `COD-03`, `TST-04`: Testes unitários com mock/dry-run para execução com e sem tmux ativo.

## Critérios de Aceitação Executáveis
1. [x] `forge-sdd status --live` aceita a flag `--live` (ou `-l`).
2. [x] Quando executado com `--live`, imprime o resumo canônico formatado contendo painel ativo, fila e git log.
3. [x] `TestStatusLiveFlag` em `cmd/forge-sdd/status_test.go` valida a saída com 100% de sucesso.
