# fix/54-npm-publish-latest-tag-404

**Branch:** `fix/54-npm-publish-latest-tag-404`
**Fase:** 54
**Depende de:** —
**Status:** `todo`

## Objetivo

`npm-publish.yml` publica a tag `beta` sem problema, mas falha ao publicar
a tag `latest` — descoberto ao promover a `v2.3.0` a estável nesta sessão.
`npm view @nathanramorim/forge-sdd dist-tags` continua mostrando
`latest: 1.9.4` mesmo após a tag `v2.3.0` ter sido publicada e o
GitHub Release correspondente ter sido criado com sucesso (não
pre-release).

## Situação atual (levantada nesta sessão)

- Workflow "Publish npm" rodou para a tag `v2.3.0` e terminou com status
  `success` no GitHub Actions, mas o step "Publish to npm" tem
  `continue-on-error: true` (`.github/workflows/npm-publish.yml:46`) — o
  erro real fica mascarado no resumo do Actions.
- Log do step (`gh run view <id> --log`) mostra:
  ```
  npm notice Publishing to https://registry.npmjs.org/ with tag latest and public access
  npm error code E404
  npm error 404 Not Found - PUT https://registry.npmjs.org/@nathanramorim%2fforge-sdd - Not found
  npm error 404 The requested resource '@nathanramorim/forge-sdd@2.3.0' could not be found or you do not have permission to access it.
  ```
- A mesma pipeline, com o mesmo `NPM_TOKEN`, publicou `2.3.0-beta` com
  `--tag beta` com sucesso em 2026-08-26 (run 32978841454). A falha é
  específica de publicar com `--tag latest`.
- Hipótese mais provável: `NPM_TOKEN` é um *granular access token* do
  npm restrito a publicar apenas versões pre-release (padrão comum de
  segurança para bloquear publicação acidental de `latest` via CI) — o
  registry retorna 404 em vez de 403 nesse cenário. Não confirmado; requer
  checagem manual em npmjs.com por quem tem acesso à conta/token (fora do
  alcance desta sessão — nenhuma ferramenta disponível acessa o painel do
  npmjs.com).

## Critérios de Aceitação Executáveis

1. Causa raiz confirmada (escopo do token, permissão de conta, ou outra)
   documentada aqui no Handoff.
2. `NPM_TOKEN` ajustado (novo token gerado com escopo correto, ou
   permissão da conta revista) para permitir publicação com `--tag
   latest`.
3. Publicação manual ou via novo push de tag confirma `npm view
   @nathanramorim/forge-sdd dist-tags` retornando `latest: 2.3.0` (ou a
   versão estável vigente no momento do fix).
4. (Opcional, recomendado) Remover `continue-on-error: true` do step
   "Publish to npm" em `.github/workflows/npm-publish.yml`, ou substituí-lo
   por uma checagem explícita pós-publish (`npm view ... dist-tags`) que
   falhe o workflow visivelmente — para que essa falha não volte a passar
   despercebida como "success".

## Handoff

Aberto nesta sessão a partir do achado durante a promoção da v2.3.0 a
estável (ver `sdd/releases/history.md`). Ainda não implementado — bloqueado
em acesso à conta npmjs.com, que nenhuma ferramenta desta sessão alcança.
