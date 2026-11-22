# ADR-0020: Deploy da LP por Cloudflare Workers

- **Data:** 2026-11-20
- **Status:** `aceito` (20/11/2026, aprovado pelo dono no chat)
- **Regras afetadas (IDs da constitution):** `ARQ-04` (o `site/` segue isolado do binário), `SEG-04` (token só em segredo do CI), `GOV-01` (esta ADR). Nenhum ID é criado ou removido.
- **Autor / Responsável:** Concepção; decisão registrada após o commit `24c1b51`

## Relação com ADRs vigentes
Nenhuma ADR anterior; a hospedagem estava só em `stack.md`.

## Contexto
`forge-sdd.org` precisa servir a LP estática de `docs/` sem Vercel e sem repositório separado.

## Opções Consideradas
1. **A. Worker com assets estáticos**, deploy por CI. 2. **B. Vercel.** 3. **C. GitHub Pages.**

## Decisão
- **D-1:** adotar **A**: `site/wrangler.jsonc` (Worker `lp-forgesdd`) serve `../docs` como assets, com rotas `forge-sdd.org/*` e `www.forge-sdd.org/*`.
- **D-2 (original, substituída em 22/11/2026):** `.github/workflows/deploy-lp.yml` publicava em push no `main` que toque `docs/**`, `site/wrangler.jsonc` ou o workflow, por `cloudflare/wrangler-action@v3` (wrangler fixo `4.165.0`), com `CLOUDFLARE_API_TOKEN` em segredo (`SEG-04`).

- **D-2 (emenda de 22/11/2026, fix-106-11):** o gatilho passa a ser `workflow_dispatch` mais o script local `scripts/lp-deploy.sh` (regera, `--check`, `npx wrangler@4.165.0 deploy` em `site/`; `--status` avisa LP desatualizada). Motivo: todo merge em `docs/` (toda nota de release) disparava deploy, 100% dos runs falhavam por permissão do token e o volume de Actions era alto; a regra é computação local primeiro. O token segue só em `secrets` (`SEG-04`).

## Consequências
- **Positivas:** deploy automático e sem custo de servidor. **Negativas:** dependência da Cloudflare e da versão fixa do wrangler.
- **Ações de acompanhamento:** não verificado: estado atual do deploy em produção.
