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

## Comportamento esperado

1. **Draft não dispara nada.** O job `pronto?` passa a rodar só quando o PR **não** é draft (`if: ... draft == false`) e no evento `ready_for_review`. Push em draft gera runs com jobs pulados, sem minuto cobrado e sem vermelho.
2. **Sem CI no push da `main`.** Remover `on.push` de `ci.yml`. O PR é a barreira; o merge squash não revalida.
3. **`scripts/ci-local.sh`** roda build, vet, testes (`-count=1`), lint se instalado e o cross-compile; ao passar, grava o selo `ci-local.json` com `head_sha` e hora. Falha não grava selo.
4. **`gh pr ready` exige o selo** do HEAD atual (como já exige o parecer). Sem selo ou com SHA velho, recusa e manda rodar `scripts/ci-local.sh`.
5. **O GitHub Actions roda uma vez**, ao sair do draft (`ready_for_review`). Novo commit em PR pronto continua cancelando o run anterior (`concurrency`), mas o fluxo normal não empurra commit depois do `ready`.
6. **Medição:** `scripts/ci-politica.py` ganha a linha "runs por dia por evento" para o dono ver a queda.

## Critérios de aceitação

- Num PR em draft, um push não cria job com falha nem minuto cobrado (verificável em `gh run list`).
- Um merge na `main` não dispara o workflow `CI`.
- `scripts/ci-local.sh` falha com teste quebrado (exit ≠ 0, sem selo) e passa limpo (exit 0, selo com o HEAD).
- `gh pr ready` sem selo do HEAD é recusado com mensagem acionável; com selo e parecer, passa.
- `go test ./scripts/... ./internal/scaffold -list 'CILocal|PRGuard'` lista ao menos 3 testes novos (`-ge 3`).
- A spec 106 (RN-9/RN-10) descreve a política nova e **não** menciona mais a matriz de três sistemas.

## Fora do escopo

- O `Deploy LP` que falha em `main` (workflow separado, mesmo sintoma de custo; vira fix próprio).
- Trocar o mecanismo de branch protection do GitHub.
