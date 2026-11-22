# fix-106-10 — Validar na máquina primeiro; o GitHub Actions roda uma vez, no fim

**Branch:** `fix/106-10-ci-local-primeiro`
**Origem:** Iniciativa 106, relato do dono em 22/11/2026: "desde o dia 20 subiu o volume de actions disparadas; a prioridade é computação local e só no final, depois de tudo validado, rodar o CI do GitHub".
**Tamanho:** `M`
**Status:** `spec` (decisão do dono: política "local primeiro")

## Sintoma

Execuções do workflow `CI` por dia, medidas com `gh api .../actions/runs`: 20/11 ≈ 48, 21/11 ≈ 166, 22/11 ≈ 121. Em 22/11, 91 vieram de `pull_request` (51 falharam) e 30 de `push` na `main`. Em 10 runs de PR com falha inspecionados, o job que falha é sempre `pronto?`.

## Feature Correlacionada

- **Arquivo de origem:** `.github/workflows/ci.yml` (feat-106 do CI em dois pontos, #181).
- **Causa-raiz (três, medidas):**
  1. `pronto?` roda em **todo** push de PR em draft e **falha de propósito**: cada commit de trabalho gera um run vermelho (51 em 22/11).
  2. `on.push.branches: [main]` repete test+lint a cada merge, sobre código que o PR já validou.
  3. Cada commit em PR pronto (`synchronize`) refaz test+lint inteiros, mesmo quando a máquina local já os rodou.
- **Evidência:** `/tmp/runs.tsv` desta investigação (414 runs de 17 a 22/11); `ci.yml` linhas 15-20, 34-37, 54, 73.

## Modelo do módulo

**Enxuto.** Sem subcomando novo no Go: um script `scripts/ci-local.sh` (mesmos passos do `ci.yml`), mudanças em `ci.yml` e na regra do fluxo.

## Dependências novas

Nenhuma (`SEG-07`): bash e Python já usados pelo repositório.

## Rules e IDs da Constituição Aplicáveis

`GOV-01`, `DUP-01`, `SEG-01`, `SEG-07`, `FLX-07`; rules `sdd-fluxo-e-portao` e a de PR em draft (ADR-0022).

## Arquivos que mudam

`.github/workflows/ci.yml` (gatilhos, chamada do script, comentário de `pronto?`), `scripts/ci-local.sh` (novo), `scripts/pr_guard.py`, `scripts/ci-politica.py`, `.agents/rules/sdd-fluxo-e-portao.md` (+ template e golden), spec 106 (`discovery.md` RN-9/RN-10, `criteria.md`, `plan.md`) e `sdd/releases/history.md`. A ADR-0022 não afirma a matriz; a decisão fica na spec 106, sem ADR nova. `docs/data/releases.json` é regenerado só se `history.md` mudar.

## Decisões do projeto aplicadas

- `GOV-01` — a mudança de política de CI toca o que a ADR/spec 106 declara (RN-9/RN-10): **atualizar a spec 106 no mesmo PR**; sem ADR nova, pois a decisão fica na própria spec.
- `DUP-01` — `scripts/ci-local.sh` e `ci.yml` rodam **a mesma lista de passos**; o workflow chama o script (um só lugar).
- `SEG-01` — o script não lê nem grava fora do repositório; sem rede.
- `FLX-07` — a mudança entra na beta 5.2.0.
- Regra do dono: decisões por componente de ask.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Passos do CI | `ci.yml` jobs `test`, `lint` e cross-compile | build, vet, test, golangci-lint, `GOOS=windows/darwin` | **extrair** para `scripts/ci-local.sh` e o job chama o script |
| Selo de validação local | `scripts/atestado.py`, modo `local` | confere o parecer do Reviewer | **reusar** o padrão: grava `.forge/runtime/ci-local.json` com `head_sha` |
| Trava do `gh pr ready` | `scripts/pr_guard.py` | exige parecer `conforme` com `head_sha` do HEAD | **estender**: exige também o selo local com o mesmo `head_sha` |

## Mapa do fluxo

Diagrama dinâmico com os agentes e onde o CI entra: [`../diagramas/fluxo-ci-local-primeiro.md`](../diagramas/fluxo-ci-local-primeiro.md).

## Comportamento esperado

1. **O gatilho é só `ready_for_review`.** `on.pull_request.types` passa de `[opened, synchronize, reopened, ready_for_review]` para `[ready_for_review]`. Push em draft, em PR pronto ou na abertura **não cria run nenhum** (hoje cria, e é daí que vem o volume). Nova rodada = voltar o PR a draft e marcá-lo pronto de novo, ou `workflow_dispatch` manual.
   **Sentido do gatilho (alinhamento do dono):** o PR fica em draft por toda a esteira (Builder, selo local, Reviewer). `ready_for_review` é o sinal de "a esteira terminou e o Reviewer liberou: rode o CI final". Nos textos e na ajuda o evento se chama **liberado para o CI** (o GitHub não aceita nome de evento próprio; o evento real continua `ready_for_review`).
2. **Sem CI no push da `main`.** Remover `on.push` de `ci.yml`. O PR é a barreira; o merge squash não revalida.
3. **`scripts/ci-local.sh`** roda build, vet, testes (`-count=1`), lint se instalado e o cross-compile; ao passar, grava o selo `ci-local.json` com `head_sha` e hora. Falha não grava selo.
4. **`gh pr ready` exige o selo** do HEAD atual (como já exige o parecer). Sem selo ou com SHA velho, recusa e manda rodar `scripts/ci-local.sh`.
5. **O GitHub Actions roda uma vez**, ao sair do draft. Os jobs `pronto?`, `test` e `lint` rodam nesse único evento; `concurrency` segue cancelando o run anterior da mesma PR. Dentro do job, `pronto?` continua recusando draft (rede de segurança para o `workflow_dispatch`).
6. **Medição:** `scripts/ci-politica.py` ganha a linha "runs por dia por evento" para o dono ver a queda.

## Critérios de Aceitação Executáveis

- Um push em PR (draft ou pronto) **não cria run**; só `ready_for_review` e `workflow_dispatch` criam (verificável em `gh run list`).
- Um merge na `main` não dispara o workflow `CI`.
- `scripts/ci-local.sh` falha com teste quebrado (exit ≠ 0, sem selo) e passa limpo (exit 0, selo com o HEAD).
- `gh pr ready` sem selo do HEAD é recusado com mensagem acionável; com selo e parecer, passa.
- `go test ./scripts/... ./internal/scaffold -list 'CILocal|PRGuardExigeSelo'` lista ao menos 3 testes (`-ge 3`); hoje listam 0, pois só existem 6 `TestPRGuard*` anteriores.
- `scripts/ci-politica.py` imprime a linha "runs por dia por evento" (teste `TestCIPoliticaRunsPorEvento` existe via `-list`).
- A spec 106 (RN-9/RN-10) descreve a política nova e **não** menciona mais a matriz de três sistemas.

## Selo local: onde é lido

- **Quem grava:** `scripts/ci-local.sh`, sempre contra o HEAD atual; qualquer commit novo invalida o selo e o Builder roda o script de novo (o Reviewer também o reexecuta).
- **Quem lê:** `pr_guard.py`, junto com o parecer, ao interceptar `gh pr ready` (e o equivalente `az repos pr update` quando o VCS for Azure). Fica em `.forge/runtime/` (dentro do repo, não versionado), então vale só na máquina que trabalha; não é prova para terceiros.

## Riscos aceitos (declarados)

1. **Botão "Ready" da web** e `gh` fora do hook não passam pelo guard: o CI roda sem selo. Mitigação futura: `pronto?` conferir o parecer no job (spec própria).
2. **PR aberto já como pronto** não dispara CI (`opened` sai). A esteira sempre abre em draft (`gh pr create --draft`); PR de contribuidor externo e `reopened` ficam sem CI até um `workflow_dispatch`.
3. **Commit depois do `ready`** não roda CI. O fluxo manda voltar a draft; quebrar isso é desvio do fluxo.
4. **Sem check obrigatório** na proteção da `main` (`required_status_checks` vazio); o comentário "job obrigatório" do `ci.yml` está desatualizado e é corrigido. O merge `--admin` já ignorava os checks.

## Fora do escopo

- O `Deploy LP` que falha em `main` (workflow separado, mesmo sintoma de custo; vira fix próprio).
- Trocar o mecanismo de branch protection do GitHub.
