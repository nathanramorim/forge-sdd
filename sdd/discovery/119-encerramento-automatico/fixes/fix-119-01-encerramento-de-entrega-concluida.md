# fix-119-01 — Entrega concluída no mundo real não encerra tarefas nem iniciativa

**Branch:** `fix/119-encerramento-automatico` (só a spec; a correção vai em branch por feature, ADR-0022)
**Origem:** relato do `hub-busize` (iniciativas `3c23-onboarding-enxuto` e `f52d`): guia publicado (Planning #38), ajustes em PR, mas o `status.md` ainda lista T1, T4 e T7 como pendentes ("Não arquivar até a T7") e o `forge status` as mostra como ativas.
**Tamanho:** `M`
**Status:** `spec`

## Sintoma

Trabalho entregue (PR mesclado, publicado) não vira "concluído" sozinho. O `status.md` e a fila ficam desatualizados, a iniciativa segue ativa no `forge status` e ninguém sugere `/archive`.

## Feature Correlacionada

- **Arquivo de origem:** `sdd/discovery/105-experiencia-cli-e-harness/features/feat-105-05-status-esconde-concluida.md`; `internal/pipeline/archive.go` (`ListArchiveCandidates`, `pendingReason`); `internal/layout/scan.go` (`buildInitiative`).
- **Causa-raiz (a confirmar na implementação):** "concluída" depende só de `queue.json` com todos os itens fechados. Nada reconcilia a fila com o mundo real (PR mesclado). Iniciativa sem fila nunca é candidata (`len(q.Items)==0`). O `status.md` é prosa livre que ninguém cruza com a fila. Além disso, a 105-05 esconde itens `completed`/`approved`, mas `pendingReason` só aceita `approved`: os dois critérios divergem.
- **Evidência:** caso do `hub-busize` acima; leitura de `ListArchiveCandidates`.

## Modelo do módulo

**Completo** (`ARQ-11`): regra nova em `internal/pipeline` (função pura, testável), adaptador em `cmd/forge-sdd`, texto nas skills e no template do `status.md`.

## Dependências novas

Nenhuma (`SEG-07`). A consulta de PR mesclado usa o adaptador `internal/vcs`.

## Rules e IDs da Constituição Aplicáveis

`DUP-01`, `DUP-02`, `ARQ-01`, `ARQ-05`, `ARQ-10`, `COD-01`, `TST-02`, `TST-03`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Decisões do projeto aplicadas

- Reconciliar é **proposta por padrão**; só grava com `--aplicar`. Nunca mescla, nunca arquiva sozinho.
- Sem rede ou sem `gh`: degrada em silêncio para o comportamento atual.
- Um critério único de "concluída" (`DUP-01`), usado por `status`, `archive` e `doctor`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Esconder concluída | feat-105-05, `layout.buildInitiative` | só olha a fila (`completed`/`approved`) | **estender**: mesma função, critério unificado |
| Arquivar concluídas | `forge archive --concluidas`, `pipeline.ListArchiveCandidates` | exige fila não vazia e itens `approved` | **estender**: aceitar `completed` e iniciativa sem fila com `status.md` marcado como concluída |
| Mudar status de tarefa | `queue.SetTaskStatuses` | já grava de forma atômica | **reusar** para fechar as tarefas |
| PR por branch | `internal/vcs` (`ProcurarPRAberto`) | acha só PR aberto | **estender** o adaptador com `ProcurarPRMesclado(branch)` |
| Higiene no doctor | `internal/doctor` | sem checagem de iniciativa velha | **estender** com aviso `iniciativa-entregue-nao-encerrada` |
| Fechamento no CLOSE | skills `sdd-build`, `sdd-governanca`, `sdd-operacao` | mandam atualizar `status.md`, sem fechar a fila | **estender** o texto do CLOSE |

Nada novo duplica conceito: tudo estende o que já existe.

## Comportamento esperado

1. `forge status` sinaliza "entregue? rode `forge archive --reconciliar`" quando há PR mesclado para as branches da fila, mas itens ainda abertos.
2. `forge archive --reconciliar [--aplicar]` lista as tarefas cujo PR está mesclado; com `--aplicar` fecha na fila, atualiza `Status` e `Próximo passo` do `status.md` (≤ 1 KB) e sugere `/archive`.
3. Iniciativa sem fila, com a linha `Estado: concluida` no `status.md`, entra em `--concluidas`. Se o `status.md` ainda diz "Não arquivar até a T…" e a fila (ou o PR) já está fechada, o `forge archive` avisa a divergência.
4. `ListArchiveCandidates` e `status` usam a mesma função de "concluída".
5. O CLOSE das skills de Builder e Reviewer passa a propor, como passo manual após o merge, rodar `forge archive --reconciliar`; só `--aplicar` grava.
6. `forge doctor` avisa (não falha) sobre iniciativa entregue e não encerrada.

## Decisões do gate (Nathan, 2026-11-23)

1. **Escopo:** um fix só, dentro da iniciativa 119 (escopo M); o Builder divide em tasks na fila.
2. **Prova de entrega:** só PR mesclado. "Publicado" (Planning, wiki) não conta.
3. **Modo:** reconciliar só com `--aplicar`; nada muda sozinho, nem no merge pelo Reviewer.
4. **Iniciativa sem fila:** declara-se concluída pela linha `Estado: concluida` no `status.md`.
5. **hub-busize (f52d e 3c23):** correção manual lá, fora deste repositório.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/pipeline ./internal/queue ./internal/vcs ./internal/layout ./internal/doctor -count=1
go test ./cmd/forge-sdd -run 'Archive|Status|Reconcil' -count=1
go test ./... -count=1
```

Cada comportamento acima ganha ao menos um teste nomeado (`TestReconciliar*`, `TestConcluidaCriterioUnico`).

## Fora do escopo

- Heurística sobre texto livre do `status.md` para achar "pendências" (só o marcador explícito de concluída).
- Publicação (Planning, wiki) como prova de entrega: só PR mesclado conta.
- Corrigir os `status.md` do `hub-busize` (fora deste repositório).
