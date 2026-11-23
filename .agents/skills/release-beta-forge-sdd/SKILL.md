---
name: release-beta-forge-sdd
description: Publicar uma versão -beta do forge-sdd (bump, notas, LP, tag, npm) sem esquecer arquivo nem conta de deploy
status: proposed
origin: auto-governance
---

# Skill: release-beta-forge-sdd

1. Branch release/<versão> a partir de origin/main (git fetch antes).
2. Bump em cmd/forge-sdd/main.go, npm/package.json, README.md, npm/README.md (cabeçalho e linha da tabela), docs/index.html e nos testes com a versão fixa (lp_*_test.go, upgrade_de50_test.go x2).
3. Notas em sdd/releases/history.md, seção nova no topo, título em negrito = ganho da pessoa, só produto forge-sdd; rode python3 scripts/novidades_lp.py.
4. go test ./... -count=1 (TestStatusLiveFlag é instável sob carga: rode isolado).
5. PR em draft, Reviewer, aprovação do dono, CI verde, merge.
6. Tag v<versão> na main: o release.yml publica no npm; confira npm view dist-tags.
7. scripts/lp-deploy.sh com a conta explícita (--conta ou .forge/config/lp-deploy.json); confira no ar.
