Leia `sdd/memory/progress.md` (e `sdd/.queue/queue.json`, se existir) assumindo o papel de **Líder** (`.agents/roles/lider.md`) e apresente o **Painel Executivo (Camada C)** para o humano:
1. **Concluído e Aprovado pela Governança (`approved` / `done`):** tasks/features entregues e validadas nas Camadas A e B.
2. **Em Andamento / Em Revisão (`building`, `review_pending`, `changes_requested`):** estado de cada etapa ativa na fila do Builder e eventuais achados de Governança.
3. **O que ainda não foi desenvolvido (`queued` / `todo`):** próximas tasks na fila ou features pendentes no roadmap.
4. **Exceções, Bloqueios e Auto-Skills:** bloqueios registrados, exceções temporárias (`constitution-exception`) e skills propostas (`forge-sdd skills list`).
5. **Sincronização remota (obrigatória antes do relatório):** consulte `sdd/memory/constitution.md` → "VCS / Work Item System".
   - Se for `nenhum`, pule esta etapa sem tentar nenhum comando de VCS (mesma regra já usada para PR).
   - Caso contrário, rode `git fetch` e compare a branch atual e `main` contra `origin` (ahead/behind). Se `github`, rode também `gh pr list` e cruze com `sdd/features/index.md` (heurística: prefixo de branch bate com a coluna `Branch` do índice).
   - Se `git`/`gh` falhar (sem rede, `gh` não instalado/autenticado), não trate como erro fatal — reporte um aviso curto ("sincronização remota indisponível: <motivo>") e prossiga com o restante do relatório baseado só no estado local.
6. **Seção "Divergência Remota" (só aparece se houver algo a reportar):** liste, com base na sincronização acima:
   - Branches remotas sem feature/fix correspondente em `sdd/features/index.md`.
   - PRs abertos não referenciados em `progress.md`/`index.md`.
   - Commits em `origin/main` ainda não incorporados à branch ativa.
7. **Comando sugerido (obrigatório):** encerre sempre com a linha `Próximo comando sugerido: <comando>`, calculada assim:
   - Divergência remota relevante encontrada (branch órfã, PR não referenciado) → sugerir investigar essa divergência antes de qualquer outro comando.
   - Nenhum discovery em `sdd/discovery/` e nenhuma feature registrada → `/discovery`
   - Discovery presente sem features correspondentes criadas em `sdd/features/` → `/split-features`
   - Existe feature com status `todo` ou item na fila → `/proxima-feature`
   - Todas as features estão `done` → `/archive` ou `/discovery` (novo ciclo)

Formato de saída: Painel Executivo em 4 seções + seção "Divergência Remota" (se houver) + lista de próximos passos + linha final "Próximo comando sugerido".
