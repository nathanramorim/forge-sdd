# fix-106-11 — O deploy da LP deixa de rodar a cada push e passa a ser acionado por quem decide

**Branch:** `fix/106-11-deploy-lp-manual`
**Origem:** Iniciativa 106, relato do dono em 22/11/2026: falhas na `main` e volume de Actions; prioridade é computação local e GitHub Actions só no fim.
**Tamanho:** `S`
**Status:** `spec`

## Sintoma

O workflow `Deploy LP` dispara em todo push na `main` que toque `docs/**` (e em `site/wrangler.jsonc` e no próprio YAML). Em 22/11 foram 4 execuções, **todas com falha** na etapa `Deploy Worker (LP)`: a API da Cloudflare recusa `assets-upload-session` e o `npx` reclama de `wrangler@4.166.0` ausente. Cada merge de PR com `docs/data/releases.json` (que está em `docs/`) gera um run vermelho.

## Feature Correlacionada

- **Arquivo de origem:** `.github/workflows/deploy-lp.yml` (commit `24c1b51`, 19/11, "migrar deploy para Cloudflare Workers + CI automático").
- **Causa-raiz:** o deploy foi ligado ao evento errado. Publicar a LP é uma decisão (depois da nota de release e da conferência), não um efeito colateral de todo merge em `docs/`. Além disso, qualquer nota de release mexe em `docs/data/releases.json`, então **toda release** dispara o deploy, com ou sem mudança visível.
- **Evidência:** `gh run list --workflow "Deploy LP"`: 4 falhas em 22/11, 5 por dia em 20 e 21/11; log do run de 18:10Z em 22/11 (`assets-upload-session ... failed`).

## Modelo do módulo

**Enxuto.** Só o workflow e a documentação de como publicar; sem Go novo.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-01` (o token `CLOUDFLARE_API_TOKEN` segue só em `secrets`, nunca em log), `DUP-01` (a checagem do `releases.json` continua em `scripts/novidades_lp.py --check`, um lugar só), `FLX-07`.

## Decisões do projeto aplicadas

- Regra do dono: computação local primeiro; GitHub Actions só no fim, validado (fix-106-10).
- Publicar a LP é pelo Wrangler a partir de `site/` ou pelo workflow, nunca pela Vercel (memória do projeto).

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Checagem do JSON contra o histórico | `scripts/novidades_lp.py --check` e `TestNovidadesGerador` | já roda no CI de PR (`go test`) | **reusar**: o deploy não precisa regerar para validar |
| Deploy local | `site/wrangler.jsonc`, Wrangler 4.165 | funciona na máquina do dono | **documentar** como caminho principal |
| Acionamento manual | `workflow_dispatch` já existe no YAML | presente | **manter como único gatilho** |

## Comportamento esperado

1. `deploy-lp.yml` perde `on.push`; fica só `workflow_dispatch`. Nenhum merge dispara deploy.
2. O deploy normal é local: `scripts/lp-deploy.sh` (novo) roda `python3 scripts/novidades_lp.py --check` e `npx wrangler@4.165.0 deploy` em `site/`, e imprime a URL e a lembrança de conferir no ar os arquivos que a página carrega (vídeo e imagens).
3. O workflow manual continua disponível para quem não tem o token local e usa o mesmo script (DUP-01).
4. A causa da recusa da Cloudflare (`assets-upload-session`) é investigada **no deploy local**: se o token do CI perdeu permissão de Workers Assets, o item vai para `pendencias-humanas` (rotacionar o token é decisão do dono, fora do repositório).
5. A doc de publicação (`sdd/memory/` ou README da LP, onde já houver) descreve os dois caminhos.

## Critérios de Aceitação Executáveis

- `grep -c "push:" .github/workflows/deploy-lp.yml` devolve `0` e `grep -c "workflow_dispatch" .github/workflows/deploy-lp.yml` devolve `1`.
- `bash scripts/lp-deploy.sh --dry-run` sai com 0, imprime os dois comandos e **não** chama a rede.
- `go test ./scripts/... -list 'LPDeploy'` lista ao menos 2 testes (`-ge 2`): dry-run sem rede e recusa quando o `--check` falha.
- Um merge na `main` que mude `docs/` não cria run do `Deploy LP` (`gh run list`).

## Fora do escopo

- Rotacionar ou recriar o token da Cloudflare.
- Trocar o provedor de hospedagem.
