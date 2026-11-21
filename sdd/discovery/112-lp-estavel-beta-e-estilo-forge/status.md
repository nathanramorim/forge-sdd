# Status — 112-lp-estavel-beta-e-estilo-forge

**Gate 1:** aprovado pelo dono em 21/11/2026 (D-1 a D-5 com as recomendações). `feat-112-03` aguarda Gate 1, aprovada pelo dono no pedido de 21/11/2026 (JSON gerado no deploy + fallback em runtime, seção própria antes do rodapé, abas com as duas versões do topo, uma branch por iniciativa).
**Status:** `feat-112-01` e `feat-112-02` na `main` (#173), aprovadas nas Camadas A e B. `feat-112-03` (seletor de novidades da versão na LP) tem spec em `features/` e ainda não foi implementada.
**Publicação:** **concluída em 21/11/2026** por deploy manual (`npx wrangler@4.165.0 deploy` em `site/`, version `0b7f05f7`), depois de 14 execuções seguidas da Action `Deploy LP` falharem. O site publicado bate byte a byte com `docs/index.html` na `main`. O deploy automático segue quebrado: o secret `CLOUDFLARE_API_TOKEN` não tem `Workers Scripts: Edit` + `Workers Routes: Edit`.
**Próximo passo:** Gate 1 da `feat-112-03` e `/implementar`.