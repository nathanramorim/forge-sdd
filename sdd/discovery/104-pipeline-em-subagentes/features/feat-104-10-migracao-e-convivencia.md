# feat-104-10 — Migração e convivência com o modo de 3 sessões

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-09, feat-104-12, feat-104-11

## Problema
O modo atual (tmux, workers) tem usuários e testes; trocar de uma vez é arriscado.

## Comportamento
1. `pipeline_mode: sessions|chat` em `forge.json`; padrão `sessions` (sem mudança para quem não opta).
2. `forge update` preserva `sessions`; `forge pipeline mode chat` troca e mostra o diagnóstico.
3. `docs/pipeline.md` e `.forge/LEIA-ME.md` descrevem os dois modos e quando usar cada um; o relatório da avaliação (feat-104-11) decide a recomendação.
4. Remoção do modo `sessions` fica fora (nova iniciativa, ADR própria).

## Modelo do módulo
**Enxuto**: estende `internal/config`, `internal/scaffold` e docs.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0004, ADR-0008, ADR-0013.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Config de modo | `internal/config` (`ForgeConfig`) | `session_mode`, `autonomy` | **Estender** | Sem chave nova duplicada |
| Docs | `docs/pipeline.md`, `.forge/LEIA-ME.md` | 3 sessões | **Atualizar** | — |
| Testes do modo atual | `internal/pipeline/e2e_test.go` | Cobrem `sessions` | **Reusar** como guarda | Regressão |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `COD-04`, `DUP-06`, `TST-01`. Rules: `sdd-arquitetura`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestPipelineModeDefaultSessions|TestUpdatePreservaModo' ./internal/config | grep -q 'TestPipelineModeDefaultSessions'
go test ./internal/config -run 'TestPipelineModeDefaultSessions|TestUpdatePreservaModo'
go vet ./... && go test ./...
```
- `go test ./internal/pipeline` (e2e do modo atual) segue verde.
- Teste negativo: valor de modo inválido é recusado.

## A confirmar no Gate 1
Quando o modo `chat` vira padrão? Padrão conservador: só por nova ADR após a avaliação.
