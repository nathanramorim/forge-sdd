# Comando: status

**Uso:** `/status` mostra o Painel Executivo, a fila e a divergência remota, sem implementar nada.

1. Rode `forge-sdd status` (e `--all-branches` para ver frentes paralelas).
2. Apresente o Painel Executivo de `references/lider.md`.
3. **Sincronização remota:** consulte `VCS / Work Item System` na Constitution; se for `nenhum`, pule. Senão, `git fetch` e compare a branch atual e `main` com `origin`; com `github`, rode `gh pr list` e cruze com as branches das iniciativas. Falha de rede ou de `gh` vira aviso curto, não erro.
4. **Divergência remota** (só se houver): branches remotas sem iniciativa, PRs abertos fora do estado local, commits de `origin/main` ausentes na branch.
5. Termine com `Próximo comando sugerido: <comando>`:
   - divergência relevante: investigue antes;
   - sem iniciativa: `/discovery`;
   - iniciativa sem features: `/split-features`;
   - tasks na fila: `/implementar`;
   - tudo aprovado: `/archive` ou novo `/discovery`.
