# 119 — Encerramento automático de entrega concluída

## Problema

Trabalho já entregue no mundo real (PR mesclado) não vira "concluído" no Forge. O `status.md` e a fila ficam desatualizados, a iniciativa segue ativa no `forge status` e ninguém sugere `/archive`. Relato do `hub-busize`: a iniciativa `3c23-onboarding-enxuto` teve o guia publicado (Planning #38), mas o `status.md` ainda lista T1, T4 e T7 como pendentes ("Não arquivar até a T7"); a `f52d` tem o mesmo problema.

## Para quem

Quem conduz iniciativas no Forge (Nathan e os projetos que usam o forge-sdd) e perde tempo com iniciativas fantasma no painel.

## Jornada

1. O PR da última tarefa é mesclado.
2. `forge archive --reconciliar` lista as tarefas cujo PR já foi mesclado.
3. Com `--aplicar`, fecha a fila e atualiza o `status.md`, e o Forge sugere `/archive`.

## Regras de negócio

- **Prova de entrega:** só PR mesclado. "Publicado" (Planning, wiki) não conta.
- **Modo:** reconciliar só com `--aplicar`; nada muda sozinho, nem no merge pelo Reviewer.
- **Iniciativa sem fila:** declara-se concluída pela linha `Estado: concluida` no `status.md`.
- Um critério único de "concluída" para `status`, `archive` e `doctor`.

## Escopo

**Dentro:** `forge archive --reconciliar [--aplicar]`; critério único de "concluída"; aviso do `doctor`; texto do CLOSE das skills propondo o passo manual.

**Fora:** heurística sobre texto livre do `status.md`; publicação como prova; corrigir os `status.md` do `hub-busize` (correção manual lá).

## Riscos

- Sem rede ou sem `gh`: degrada em silêncio para o comportamento atual.
- Marcar como concluída uma tarefa cujo PR foi mesclado por engano: por isso nada grava sem `--aplicar`.

Decisões do gate de 2026-11-23 e o detalhe técnico: `fixes/fix-119-01-encerramento-de-entrega-concluida.md`.
