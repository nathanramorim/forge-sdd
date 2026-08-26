Leia `sdd/memory/progress.md` e reporte:
1. Features concluídas na última sessão
2. Feature ativa (status `in-progress`) e tasks pendentes
3. Próxima feature (status `todo`) e suas tasks
4. Bloqueios registrados
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
   - Existe feature com status `todo` → `/proxima-feature`
   - Todas as features estão `done` → `/archive` ou `/discovery` (novo ciclo)

Formato de saída: tabela markdown + seção "Divergência Remota" (se houver) + lista de próximos passos + linha final "Próximo comando sugerido".
