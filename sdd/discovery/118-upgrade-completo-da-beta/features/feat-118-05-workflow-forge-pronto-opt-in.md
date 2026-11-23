# feat-118-05 — Workflow próprio `forge-pronto.yml` (opt-in `--ci`) e texto do `/upgrade-sdd` honesto

**Depende de:** feat-118-01 e feat-118-02 (onda 2); onda 3, em paralelo com 118-03 e 118-04
**Achado:** 4. **Tamanho:** `M`. **Status:** `spec`

## Objetivo

Quem quiser o job `pronto?` roda `forge update --ci` (ou `forge init --ci`) e recebe o arquivo **novo** `.github/workflows/forge-pronto.yml`. O `ci.yml` do usuário nunca é lido para escrita. O `/upgrade-sdd` deixa de prometer o que nenhum código entrega (RN-3): acréscimo no `ci.yml`, `ci-politica.py`, `atestado.py` e campos de modo (estes últimos passam a ser entregues pela 118-03).

## Entradas

- Job `pronto?` já existente em `.github/workflows/ci.yml:31-47` (recusa PR em draft; usa `gh pr view` e `GITHUB_TOKEN`).
- Texto `templates/.agents/skills/sdd-operacao/references/upgrade-sdd.md.tmpl` (linhas 11-40) e o golden dele.

## Saídas

- Template em `templates/ci/forge-pronto.yml.tmpl` (**fora** de `templates/.github`, que `scaffold.go:81-83` põe em `globalRoots` com o Copilot e o `renderDir` renderizaria sem `--ci`), destino `.github/workflows/forge-pronto.yml`, com destino mapeado explicitamente (o `renderDir` cairia em `ci/`). Só o job `pronto?` com `if: github.event_name == 'pull_request'`, `permissions: pull-requests: read` e `on: pull_request` com `types: [ready_for_review]`. Renderizado **apenas** com `--ci`. Conferir que a diretiva `//go:embed` cobre `templates/ci`.
- Flag booleana `--ci` em `init` e `update` (flag, não subcomando: `ARQ-06` não é emendada). Sem `--ci` nada é criado; com `--ci` e arquivo já existente sem o marcador de gerado, preserva e avisa (`SEG-02`).
- `upgrade-sdd.md.tmpl` e golden reescritos: o forge entrega o hook e o script (14a/14b), campos de modo (118-03) e, com `--ci`, o `forge-pronto.yml`; o usuário decide o que mudar no `ci.yml` dele; `ci-politica.py` e `atestado.py` deixam de ser prometidos; aviso de que edição nas skills `sdd-*` é sobrescrita.

## Modelo do módulo

**Completo** (`ARQ-11`): flag no CLI, template, `shouldPreserve` e texto da skill. Justificativa: toca comando, scaffold e a skill embutida.

## Decisões do projeto aplicadas

- Decisão do dono (22/11/2026): workflow próprio e opt-in; `ci.yml` do usuário intacto; ADR-0022/0023 (PR e CI em dois pontos).
- `ARQ-06`: flag sobre `init`/`update` existentes, sem comando novo e sem emenda.
- `ARQ-03`: o workflow vem por `embed.FS`.
- `SEG-02`: nunca sobrescreve workflow do usuário; marcador `# forge-sdd:generated` (mesma decisão do 118-02).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Job `pronto?` | `.github/workflows/ci.yml:31-47` | funciona só no repo do forge-sdd | **reusar** o passo como conteúdo do template | `DUP-02` |
| Templates `.github` | `scaffold.go:81-83` (`globalRoots` inclui `templates/.github` com o Copilot) | tudo sob `templates/.github` é renderizado com o Copilot | **criar** `templates/ci/` fora dele, com entrada opt-in | não vazar o workflow sem `--ci` |
| Flags de `update` | `cmd/forge-sdd/main.go:445-451` | sem `--ci` | **estender** | flag opt-in |
| Texto do `/upgrade-sdd` | `upgrade-sdd.md.tmpl:18-27` | promete `ci.yml`, scripts e campos | **estender** (corrigir) | RN-3 |
| Dono do arquivo gerado | feat-118-02 | marcador único | **reusar** | `DUP-06` |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-02`, `ARQ-06` (flag, não subcomando), `ARQ-03`, `ARQ-11`, `DUP-02`, `TST-02` (template, texto e golden no mesmo PR), `TST-03` (sem `--ci` **incluindo o agente Copilot**, com `--ci`, workflow existente do usuário).

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestSemCINaoCriaWorkflow|TestComCICriaForgePronto|TestWorkflowDeclaraReadyForReview|TestCIPreservaWorkflowDoUsuario|TestCINuncaTocaNoCiYml|TestUpgradeSddNaoPrometeOQueNaoEntrega)$' -count=1
go test ./cmd/forge-sdd -run '^(TestUpdateFlagCIEOptIn|TestInitFlagCIEOptIn)$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestSemCINaoCriaWorkflow|TestComCICriaForgePronto|TestWorkflowDeclaraReadyForReview|TestCIPreservaWorkflowDoUsuario|TestCINuncaTocaNoCiYml|TestUpgradeSddNaoPrometeOQueNaoEntrega|TestUpdateFlagCIEOptIn|TestInitFlagCIEOptIn)$' | grep -c '^Test')" -eq 8 ]
go test ./internal/scaffold ./cmd/forge-sdd -count=1
```

## Fora do escopo

- Entregar `scripts/ci-politica.py` e `scripts/atestado.py`.
- Reescrever ou mesclar o `ci.yml` do usuário.
