---
description: Onde moram a LP e a wiki do forge-sdd e como publicá-las (regra deste repositório, não vai para os projetos dos usuários).
globs: ["docs/**", "site/**", "scripts/lp-deploy.sh", ".github/workflows/release.yml"]
---

# LP e wiki do forge-sdd

- **A LP e a wiki vivem neste repositório.** A LP é `docs/index.html` (com `site/` para o Astro e o Wrangler); a wiki é `docs/wiki.html` (rota `/wiki`). Não existe mais o repositório `lp-forge-sdd`: nada é copiado nem sincronizado para fora.
- **Documentação nova** (guia, passo a passo, página de ajuda) entra como seção em `docs/wiki.html`, com a entrada correspondente no menu lateral, e não como `.md` solto em `docs/` esperando um sync.
- **Publicar** é `scripts/lp-deploy.sh` (local, com `--dry-run` e `--status`). O workflow `Deploy LP` só roda por `workflow_dispatch`.
- **Novidades da LP** vêm de `sdd/releases/history.md` via `python3 scripts/novidades_lp.py`, que regera `docs/data/releases.json`. As notas falam só do produto forge-sdd (nada de LP, site ou deploy).
- O `release.yml` não sincroniza wiki: se alguém reintroduzir esse passo, é regressão.
