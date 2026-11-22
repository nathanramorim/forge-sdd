# fix-106-11 — O deploy da LP deixa de rodar a cada push e passa a ser acionado por quem decide

**Branch:** `fix/106-11-deploy-lp-manual`
**Origem:** Iniciativa 106, relato do dono em 22/11/2026: falhas na `main` e volume de Actions; prioridade é computação local e GitHub Actions só no fim.
**Tamanho:** `M`
**Status:** `spec`

## Sintoma

Duas coisas distintas, medidas com `gh run list --workflow "Deploy LP"`:

1. **Volume:** o workflow dispara em todo push na `main` que toque `docs/**`, `site/wrangler.jsonc` ou o próprio YAML: 3 runs em 19/11, 5 em 20/11, 5 em 21/11 e 5 em 22/11. Toda nota de release mexe em `docs/data/releases.json`, então toda release dispara deploy.
2. **Falha:** **100% dos runs falham desde 19/11, inclusive os 3 `workflow_dispatch`** (20/11 ×2, 21/11 ×1). O erro final é a API da Cloudflare recusando `.../assets-upload-session` ("No access to the specified resource"): é **permissão do token** do CI, não o gatilho. O aviso `npx canceled ... wrangler@4.166.0` do log é inofensivo (a checagem `--no-install` da action; a instalação do 4.165.0 segue depois) e **não** é a causa.

## Feature Correlacionada

- **Arquivo de origem:** `.github/workflows/deploy-lp.yml` (commit `24c1b51`, 19/11) e a ADR-0020 (D-2: "publica em push no main que toque `docs/**`").
- **Causa-raiz do volume:** o deploy foi ligado a todo merge em `docs/`, e a nota de release sempre toca `docs/`.
- **Causa-raiz da falha:** token `CLOUDFLARE_API_TOKEN` sem permissão de Workers Assets. Tirar o `push` **não** conserta o deploy; só tira o ruído. **A LP continua sem publicar pelo CI até o dono resolver o token** (fora do repositório).
- **Evidência:** log do run `45053581186` (22/11, 18:10Z) e a lista de runs acima.

## Modelo do módulo

**Enxuto.** Só o workflow e a documentação de como publicar; sem Go novo.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`GOV-01` (muda a D-2 da ADR-0020: emenda no mesmo PR), `ARQ-04`, `SEG-04` (o token segue só em `secrets`, nunca em log), `DUP-01` (regerar e `--check` ficam no script, um lugar só).

## Decisões do projeto aplicadas

- **ADR-0020, D-2:** substituída por emenda no mesmo PR: o gatilho passa a ser `workflow_dispatch` e o script local; `sdd/spec/stack.md` (linhas 28-29) é atualizado junto.
- Regra do dono: computação local primeiro; GitHub Actions só no fim, validado (fix-106-10).
- Publicar a LP é pelo Wrangler a partir de `site/` ou pelo workflow, nunca pela Vercel (memória do projeto).

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Checagem do JSON contra o histórico | `scripts/novidades_lp.py --check` e `TestNovidadesGerador` (`internal/scaffold/lp_novidades_test.go:145`) | o teste valida, mas **não regera**; o deploy atual regera e depois valida | **reusar**: o script local **regera e então roda o `--check`** (senão sobe JSON atrasado) |
| Deploy local | `site/wrangler.jsonc`, Wrangler 4.165 | funciona na máquina do dono | **documentar** como caminho principal |
| Acionamento manual | `workflow_dispatch` já existe no YAML | presente | **manter como único gatilho** |

## Comportamento esperado

1. `deploy-lp.yml` perde `on.push`; fica só `workflow_dispatch`. Nenhum merge dispara deploy.
2. `scripts/lp-deploy.sh` (**novo**, ainda não existe) roda `python3 scripts/novidades_lp.py`, depois `--check`, e `npx wrangler@4.165.0 deploy` em `site/`; imprime a URL e lembra de conferir no ar os arquivos que a página carrega (vídeo e imagens). O conjunto de assets não muda (`docs/.assetsignore` vale igual no local e no CI).
3. O workflow manual chama o mesmo script (DUP-01).
4. **Lembrete contra LP velha:** o script `scripts/lp-deploy.sh --status` compara o último commit que tocou `docs/` com o último deploy registrado em `.forge/runtime/lp-deploy.json` e avisa "LP desatualizada"; o item de release (checklist de publicação) manda rodar o deploy. Sem isso a LP envelheceria sem sinal.
5. **Token:** a recusa da Cloudflare vai para `pendencias-humanas` (permissão de Workers Assets do `CLOUDFLARE_API_TOKEN`); rotacionar é decisão do dono.
6. **Consumidores atualizados:** ADR-0020 (emenda da D-2), `sdd/spec/stack.md`, e as menções de deploy por push em `feat-112-03`, `101-01`, `101-03` e no `plan.md`/`status.md` da 112.

## Critérios de Aceitação Executáveis

Os de 2 e 3 ficam **vermelhos até o Builder criar o script** (esperado).

- `[ "$(grep -cE '^\s*push:' .github/workflows/deploy-lp.yml)" -eq 0 ]` e `[ "$(grep -c workflow_dispatch .github/workflows/deploy-lp.yml)" -ge 1 ]`.
- `bash scripts/lp-deploy.sh --dry-run` sai com 0, imprime os comandos (regerar, `--check`, `wrangler deploy`) e não chama a rede.
- `[ "$(go test ./scripts/ -list 'LPDeploy' | grep -c '^Test')" -ge 3 ]`: dry-run sem rede, recusa quando o `--check` falha, `--status` avisa LP desatualizada.
- `grep -rn "push no main" sdd/spec/adr/adr-0020-deploy-da-lp-por-cloudflare-workers.md` não acha mais a D-2 antiga (emenda registrada) e `sdd/spec/stack.md` não afirma deploy automático.
- **Pós-merge (informativo):** um merge em `docs/` não cria run do `Deploy LP` (`gh run list`).

## Fora do escopo

- Rotacionar ou recriar o token da Cloudflare.
- Trocar o provedor de hospedagem.
