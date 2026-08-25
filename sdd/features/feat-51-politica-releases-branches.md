# feat/51-politica-releases-branches

**Branch:** `feat/51-politica-releases-branches`
**Fase:** 51
**Depende de:** —
**Status:** `done`

## Objetivo

Formalizar e automatizar a política de retenção de Releases/tags no GitHub
(hoje sem limite — `gh release list` mostra 49+ entradas, incluindo dezenas
de rascunhos duplicados de tentativas antigas de `v1.1.0`, `v1.6.0-beta.0`,
`v1.6.1-beta.0/1` etc.) e reforçar que toda versão beta publicada tenha
destaque explícito do que foi melhorado — hoje isso já ocorre de forma
manual e inconsistente em `sdd/releases/history.md`. Solicitado pelo
usuário: "revise as branches e releases, no git pode deixar somente as 3
ultimas releases, as versoes beta devem ter destaque falando do que foi
melhorado".

## Situação atual (levantada nesta sessão)

- `gh release list` retorna 49 releases, com múltiplos rascunhos (`Draft`)
  duplicados para a mesma tag (ex: 5x `v1.1.0`, 5x `v1.6.1-beta.0`) —
  sobras de execuções antigas do workflow de release.
- `.goreleaser.yaml` (`release:`) não define `prerelease: auto` — releases
  de tags `-beta` não são marcadas com o badge "Pre-release" do GitHub,
  ficando visualmente idênticas às estáveis.
- `.github/workflows/release.yml` ("Extract Release Notes") já busca a
  seção correspondente em `sdd/releases/history.md` por versão, mas cai em
  uma mensagem genérica (`"Release notes para a versão X. Veja
  sdd/releases/history.md"`) quando não encontra — sem checagem que force
  destaque real para tags beta.
- Branches remotas obsoletas identificadas (candidatas a limpeza, a
  confirmar merge status com o usuário antes de apagar):
  `chore/complete-feat-24`, `chore/release-1.9.2`, `chore/release-1.9.3`,
  `chore/remove-feat-25`, `docs/sync-npm-readme-news`,
  `feat/add-new-model-preserve`, `feat/cli-autopilot-autonomy`,
  `feat/dev-flow-guide`, `feat/npm-release-and-deploy`,
  `feat/open-source-readme`, `feat/release-44-fluxo-sdd`,
  `feat/remove-homebrew`, `feat/upgrade-node-ci`, `release/1.9.1`.
- Não há, hoje, nenhum mecanismo (workflow ou script) que apague
  releases/tags antigos — a retenção é 100% manual/inexistente.

## Critérios de Aceitação Executáveis

1. `.goreleaser.yaml` ganha `prerelease: auto` em `release:`, para que toda
   tag contendo `-beta`/`-rc`/etc. seja automaticamente marcada como
   "Pre-release" no GitHub, distinguindo visualmente betas de estáveis sem
   depender de passo manual.
2. Novo step (ou job) no `.github/workflows/release.yml`, executado **após**
   o `goreleaser-action`, que:
   - Lista releases via `gh release list` ordenados por data.
   - Mantém as **3 releases mais recentes** (independente de
     beta/estável) e apaga as demais com `gh release delete <tag> --yes`
     seguido de `git push origin :refs/tags/<tag>` para remover a tag
     correspondente.
   - Roda de forma idempotente (não falha se já houver ≤ 3 releases).
3. O mesmo step (ou um step anterior) falha explicitamente o workflow se a
   tag publicada for beta (`-beta` no nome) e a seção extraída de
   `sdd/releases/history.md` for igual à mensagem de fallback genérica —
   forçando que toda beta tenha, de fato, uma lista de destaques (`*
   **Destaque:** ...`) antes de ser publicada.
4. Limpeza pontual (única, não recorrente) do estado atual do repositório,
   executada manualmente por este agente **com confirmação explícita do
   usuário antes de qualquer comando destrutivo**:
   - Releases/tags além das 3 mais recentes apagados no GitHub, aplicando a
     mesma regra do item 2 ao histórico existente.
   - Branches remotas já mescladas em `main` (validar com
     `git branch -r --merged origin/main`) listadas no item "Situação
     atual" apagadas; branches não mescladas (ex: `feat/cli-autopilot-autonomy`,
     que segue em teste conforme `sdd/memory/progress.md`) preservadas e
     reportadas ao usuário para decisão manual.
5. `sdd/memory/constitution.md` ganha uma regra curta documentando a
   política de retenção (3 releases mais recentes no GitHub; NPM continua
   retendo todas as versões publicadas, pois pacotes são imutáveis — ver
   Fase 49 em `progress.md`) e a exigência de destaque obrigatório em
   releases beta.

## Handoff

Ambiguidade resolvida com o usuário no Build: "3 últimas releases" = as 3
releases **estáveis** mais recentes, mais qualquer beta publicada depois
da mais antiga dessas 3 (betas ativas à frente da última estável retida).
Implementado em `.goreleaser.yaml` (`prerelease: auto`) e em dois novos
steps de `.github/workflows/release.yml` ("Enforce Beta Release
Highlights" e "Enforce Release Retention Policy"), documentado na Regra 12
da Constituição. Limpeza pontual executada com confirmação explícita do
usuário: releases/tags/rascunhos fora da política e 13 branches remotas
com PR já mergeado foram apagados; 3 branches sem PR merged e com commits
fora de `main` (`feat/npm-release-and-deploy`, `feat/open-source-readme`,
`feat/release-44-fluxo-sdd`) foram preservadas por decisão do usuário —
seguem como possível trabalho a resgatar ou descartar em sessão futura.
