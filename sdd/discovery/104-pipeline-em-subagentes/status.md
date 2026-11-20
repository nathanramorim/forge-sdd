# Status — 104-pipeline-em-subagentes

**Estado:** ondas 1 a 3 em 20/11/2026: 104-01, 02, 03, 04, 05, 06, 07 e 11 conformes e mescladas; 12 specs passam no `gate --criterios`. Onda 4 pausada a pedido do dono.
**Próximo passo:** 104-12 (rastreio, depende da 06), 104-08, 104-09, 104-10. Aceites do dono no PR: `mcp serve`/ARQ-06, merge por lote `--no-ff`, desvios da 104-03. Fix à parte: `renderDir` segue symlink em `.claude`.
**Pendências fora do escopo:** `TestStatusLiveFlag` instável por tempo (falha também no `main`); 5 pareceres da 97 com `task_id` não decodificam.
