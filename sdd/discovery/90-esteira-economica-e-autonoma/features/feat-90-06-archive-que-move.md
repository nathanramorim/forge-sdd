# feat-90-06 — Archive que move, no fechamento

**Iniciativa:** 90-esteira-economica-e-autonoma
**Branch:** feat/90-esteira-economica-e-autonoma (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** feat-90-01, feat-90-02

## Problema
`AutoArchiveInitiative` copia em vez de mover, repete a linha do progress-log a cada aprovação, grava telemetria zerada e roda em qualquer branch. Resultado: 84, 88 e 89 ficaram em `sdd/discovery` e em `sdd/archive`.

## Comportamento
1. Move `sdd/discovery/<id>` para `sdd/archive/discovery/<id>` (a origem deixa de existir), uma linha no progress-log, telemetria com os totais reais (feat-90-02). Idempotente. Só na branch da iniciativa e sem task ou fix pendente; senão recusa com o motivo.
2. Ao aprovar o último job (`review.go`, `CheckAndAutoArchive`): arquiva sozinho só se a fila tem `origem: toca-ai` (gravada pelo motor da 89, `goal.go`); nas demais, apenas avisa "pronta para arquivar" no inbox.
3. `forge archive --concluidas --dry-run` lista as candidatas sem mover.

## Modelo do módulo
**Enxuto** (ARQ-11): corrige `internal/pipeline/archive.go`; flag em `cmd/forge-sdd/archive.go`; campo `origem` em `internal/queue`.

## Decisões do projeto aplicadas
ADR-0009 (D-4 ajustado), ADR-0012, ADR-0014; promessa de archive do `/toca-ai` (feat-89-01).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Archive | `pipeline.AutoArchiveInitiative`, `CheckAndAutoArchive` | Copia e repete | corrigir na mesma função | Uma implementação |
| Motor do `/toca-ai` | `pipeline/goal.go` | Enfileira as etapas | gravar `origem` | Mantém a promessa da 89 |
| Telemetria | `monitor.RecordSessionMetrics` | Zerada no archive | usar totais reais | feat-90-02 |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`DUP-01`, `SEG-02`, `FLX-04`, `COD-04`, `TST-03`. Rules: `sdd-reuso-e-duplicacao`, `sdd-seguranca`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
go test ./internal/pipeline ./cmd/forge-sdd -run 'TestArchive_MoveIdempotente|TestArchive_AoAprovar|TestArchiveCmd_Concluidas'
go vet ./... && go test ./...
```
- Rodar duas vezes não muda nada; fora da branch ou com pendência, recusa.
- Iniciativa `origem: toca-ai` arquiva ao aprovar o último job; as demais só avisam.
