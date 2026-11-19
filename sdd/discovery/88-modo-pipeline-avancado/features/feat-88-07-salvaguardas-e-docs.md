# feat-88-07 — Salvaguardas do modo avançado e documentação

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** feat-88-02, feat-88-06

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** `internal/pipeline` (ARQ-01) e docs.

## Decisões do projeto aplicadas
- ADR-0009 (tokens e reset, proposto), ADR-0013.
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Detecção de limite de uso | `internal/pipeline/usagelimit.go` | Detecta limite da CLI | reusar | Pausa em vez de falhar. |
| Timeout do job | `builder.job_timeout_minutes` | Já existe | reusar | Soma tetos de tokens e de jobs/dia. |
| Doctor | `cmd/forge-sdd/doctor.go` | Check-up | estender | Avisa `advanced` sem ambiente isolado. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-05`, `COD-01`, `COD-03`, `TST-01`
- Rules: `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. Tetos de tokens e tempo por job e de jobs por dia; ao estourar, para o despacho com motivo no inbox.
2. `.forge/runtime/STOP` e `forge down` param a esteira de forma limpa.
3. `forge doctor` avisa; LEIA-ME e `sdd/FLOW.md` documentam o modo (experimental).

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/pipeline -run TestSalvaguardas && go test ./...
```
- `TestSalvaguardas_TetoPorJob`, `TestSalvaguardas_TetoDiario`, `TestSalvaguardas_Stop`, `go test ./...` verde.
- `go vet ./...` limpo.
