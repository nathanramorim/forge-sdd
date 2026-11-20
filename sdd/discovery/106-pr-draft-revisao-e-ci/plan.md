# Plano — 106 (2 ondas, de baixo para cima)

Regra de ordem: contrato e scripts primeiro; depois quem os usa; passos dos comandos e goldens por último. Nenhuma task roda sem o Gate 1 (`FLX-01`). Cada task traz teste negativo (`TST-03`) e passa em `forge gate --criterios`.

## Onda 1 — Mecanismo (veredito, lições, hooks, ordem do PR)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 1.1 | `harness.Parecer` ganha `head_sha` e `ponto` (`spec` ou `entrega`); `persistVerdict` os grava a partir do HEAD; validação recusa veredito sem SHA | — | com 1.2 | `DUP-06`, `FLX-06`, `SEG-01`, `TST-03` |
| 1.2 | `scripts/licoes.py` (`uv`, `pathlib`) + `sdd/memory/licoes/` + teste; `forge lessons add` passa a gravar um arquivo por lição (migração conforme Decisão v) | — | com 1.1 | `DUP-01`, `ARQ-03`, `GOV-03`, `TST-03` |
| 1.3 | `scripts/pr_guard.py --check` (bloqueia `gh pr create` sem `--draft`; `gh pr ready` e `gh pr merge` sem veredito conforme do SHA atual); consulta ao `uv` no context7; testes `unittest` | 1.1 | — | `SEG-01`, `SEG-03`, `SEG-07`, `FLX-08` |
| 1.4 | Instalador do hook `PreToolUse` (matcher `Bash`) em `hooks.go` e no `.claude/settings.json` do repo; `doctor` checa `uv` e o hook | 1.3 | — | `SEG-02`, `ARQ-05`, `TST-02` |
| 1.5 | Sinais `spec-para-revisao` e `pronto-para-revisao` no `status.md` (validação e leitura em `internal/`); revisão só dispara no sinal; teto 2 rodadas por ponto (`max_rework_attempts` padrão 2) e pergunta ao humano | 1.1 | com 1.3 | `FLX-06`, `ARQ-01`, `TST-03` |
| 1.6 | `forge queue review` e `pipeline.OpenPR`: veredito conforme não cria PR; PR existente é o padrão; `ready` só com PR existente e SHA igual; achado vira lição (chama 1.2) | 1.1, 1.2, 1.5 | — | `ARQ-02`, `DUP-03`, `FLX-03` |
| 1.7 | Passos dos comandos e skills: `/discovery`, `/nova-feature`, `/novo-fix` criam branch, commitam, empurram e abrem draft; `/revisar` e `/implementar` seguem a ordem nova; menus "fechar o PR?" e "mesclar?"; leitura só das lições da rule; templates e goldens regenerados | 1.4, 1.6 | — | `DUP-10`, `TST-02`, `COD-05`, `FLX-01` |

## Onda 2 — CI, ADR e adoção
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 2.1 | ADR final (próxima livre) e ajuste de `GOV-03`; Constituição e `concepts.md` | onda 1 | com 2.2, 2.3 | `GOV-01`, `FLX-05` |
| 2.2 | Modelo de workflow: `if` de draft, `ready_for_review`, `concurrency` só em PR, matriz de 3 SO só em PR pronto e `main`, job "pronto?" (Decisão iii); aplicar ao `ci.yml` do forge-sdd | — | com 2.1, 2.3 | `ARQ-03`, `SEG-04` |
| 2.3 | `scripts/ci-politica.py`: lista repositórios, marca ok/fora, alerta de gasto e checagem mensal; sem token no git | — | com 2.1, 2.2 | `SEG-04`, `SEG-05` |
| 2.4 | README: um prompt curto "cole isto na sua IA" por ferramenta (hook, lições, CI, política); notas para Windows, Linux e macOS | 2.2, 2.3 | — | `COD-05`, `GOV-04` |
| 2.5 | `/upgrade-sdd` leva a política; `forge update` e goldens; nota de versão | 1.7, 2.1 | — | `TST-02`, `FLX-07` |
| 2.6 | Prova: PR de teste draft para ready (jobs `skipped`, depois uma execução) em cada repositório da primeira leva; confirma Decisão ii | 2.2 | — | `TST-01` |
| 2.7 | Aplicar nos repositórios da primeira leva (Decisão iv) via `/upgrade-sdd`, com aprovação do dono por repositório | 2.5, 2.6 | — | `FLX-01` |

Caminho crítico: 1.1, 1.3, 1.4, 1.7, 2.5, 2.7. A task 1.7 e a 2.5 tocam templates e goldens que a 105 também altera (onda 1 da 105): sequenciar com ela para não regravar goldens duas vezes.
