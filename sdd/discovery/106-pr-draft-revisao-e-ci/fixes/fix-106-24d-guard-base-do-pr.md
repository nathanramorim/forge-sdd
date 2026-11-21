# fix-106-24d — O guard recusa merge de PR com base fora da principal

**Branch:** `fix/106-24d-guard-base-nao-principal`
**Origem:** Iniciativa 106, PRs #187, #203 e #205
**Tamanho:** `S`
**Status:** `done`

## Sintoma

Três PRs foram mesclados com **base em outra branch** (`feat/106-24b-leque-por-worktree`), não na
`main`. Como a branch de baixo entrou por *squash*, o commit de cima ficou órfão: o arquivo existia
no Git, mas **sumiu da `main`**.

O caso concreto: o #203 (106-24c) foi mesclado na base da 24b; a 24b entrou por squash no #201; e
`internal/pipeline/onda.go` **nunca chegou à `main`** — a iniciativa foi declarada concluída com o
arquivo ausente.

## Feature Correlacionada

- **Arquivo de origem:** `scripts/pr_guard.py` — barreira de PR do Claude Code (`feat-106-14`)
- **Causa-raiz:** o guard conferia o parecer do HEAD e o `--draft`, mas **não a base do PR**. Um
  merge com base em branch de feature passava sem barreira.
- **Evidência:** `git log --all --oneline -- internal/pipeline/onda.go` → o commit existe em
  `0dcbccd`, e `test -f internal/pipeline/onda.go` na `main` → ausente.

## Modelo do módulo

**Enxuto** (`ARQ-11`): uma verificação nova no script existente, sem pacote Go novo e sem
dependência.

## Decisões do projeto aplicadas

- `DUP-01` — a barreira continua num só lugar (`pr_guard.py`), não num script novo.
- `SEG-01` — a base lida do `gh` é dado externo; valor ausente ou ilegível **não bloqueia**, porque
  o guard não pode inventar uma recusa sem prova.
- `RN-22` — um PR de feature tem a principal como base; nada tem base em outro PR.
- `TST-03` — o teste cobre o caso positivo (base na main passa) e o negativo (base em feature
  bloqueia), com um `gh` falso no PATH.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Verificação de base do PR | `scripts/pr_guard.py` | Ausente | **criar** `gh_base_do_merge`, `base_do_pr`, `principal_do_repo` |
| Leitura do número do PR | mesmo arquivo | `extrair_subcomando_gh_pr` devolve a ação e os args | **reusar** os args |
| Execução do `gh` | `internal/vcs` | `RunGitOrGh` | **não usar** aqui: o guard é Python e roda antes do binário |

## Dependências novas (`SEG-07`)

`Nenhuma`. `subprocess` e `json` da stdlib.

## Rules e IDs da Constituição Aplicáveis

`DUP-01`, `SEG-01`, `TST-03`, `FLX-02`, `FLX-03`.

## Critérios de Aceitação Executáveis

```bash
go test ./scripts/... -run 'TestPRGuardBloqueiaMergeComBaseForaDaPrincipal'
go test ./scripts/...
go vet ./... && go test ./...
```
