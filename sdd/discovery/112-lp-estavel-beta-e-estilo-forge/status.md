# Status — 112-lp-estavel-beta-e-estilo-forge

**Gate 1:** aprovado (D-1 a D-5) e `feat-112-03` aprovada em 21/11/2026 — JSON no deploy + fallback em runtime, seção antes do rodapé, abas com as duas versões do topo, uma branch por iniciativa.
**Status:** 01 e 02 na `main` (#173). 03 implementada: `scripts/novidades_lp.py`, `docs/data/releases.json`, seção `#novidades`, passos em `deploy-lp.yml` e `release.yml`, `lp_novidades_test.go` (4 verdes), Instagram trocado por LinkedIn.
**Publicação:** feita em 21/11/2026 por deploy manual (version `0b7f05f7`), após 14 falhas seguidas da Action. O deploy automático segue quebrado: o secret `CLOUDFLARE_API_TOKEN` sem `Workers Scripts: Edit` + `Workers Routes: Edit`.
**Próximo:** Camada B, PR em draft e merge da 03; regenerar o token da Cloudflare.