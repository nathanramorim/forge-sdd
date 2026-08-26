# feat/03-ergonomia-de-comandos-e-sincronizacao (03-03)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-03
**Depende de:** — (independente das demais features deste pacote)
**Status:** `todo`

## Objetivo

`.agents/commands/status.md` passa a rodar uma etapa de sincronização remota (`git fetch`, comparação ahead/behind, `gh pr list`) antes de montar o relatório, evitando que `progress.md`/`index.md` sejam interpretados/editados com base em estado desatualizado em relação ao GitHub (branch órfã, PR aberto não referenciado).

## Critérios de Aceitação Executáveis

1. `.agents/commands/status.md` ganha uma etapa executada antes do relatório: `git fetch`, comparação ahead/behind da branch atual e de `main` contra `origin`.
2. Se "VCS / Work Item System" (`sdd/memory/constitution.md`) for `github`, roda `gh pr list` e cruza com `sdd/features/index.md` (heurística: prefixo de branch bate com a coluna `Branch` do índice).
3. Se VCS for `nenhum`, ou `gh` estiver indisponível/sem rede, a etapa é pulada com um aviso curto no relatório — nunca erro fatal, consistente com a regra já existente de não tentar comando de VCS quando `nenhum`.
4. Relatório de `/status` ganha uma seção "Divergência Remota" (só aparece quando há algo a reportar) listando: branches remotas sem feature/fix correspondente no índice, PRs abertos não referenciados em `progress.md`/`index.md`, e commits em `origin/main` ainda não incorporados à branch ativa.
5. O cálculo de "Próximo comando sugerido" (regra já existente) passa a considerar a divergência remota encontrada.
6. Sem regressão no formato de saída já existente (tabela markdown + lista de próximos passos + linha final) quando não há nenhuma divergência a reportar.

## Handoff

Independente de 03-01/03-02/03-04/03-05 — pode ser feita em paralelo.
