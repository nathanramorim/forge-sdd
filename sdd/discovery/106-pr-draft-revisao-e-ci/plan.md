# Plano — 106 (3 ondas, de baixo para cima)

Ordem: contrato e scripts primeiro; depois o fluxo; CI e atestado por último. Nada roda sem a reaprovação do escopo (`FLX-01`). Cada task traz teste negativo (`TST-03`) e passa em `forge gate --criterios`. Até 7 tasks por onda.

## Onda 1 — Mecanismo (contrato, scripts locais, hooks)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 1.1 | `harness.Parecer` ganha `head_sha`, `ponto` (`spec` ou `entrega`) e `modo` (`com-humano` ou `sem-humano`); `persistVerdict` os grava a partir do HEAD; veredito sem SHA é recusado | — | com 1.2 | `DUP-06`, `FLX-06`, `SEG-01`, `TST-03` |
| 1.2 | `scripts/licoes.py` (`uv`, `pathlib`) + `sdd/memory/licoes/` + teste; `forge lessons add` grava um arquivo por lição; migra as 2 entradas atuais | — | com 1.1 | `DUP-01`, `ARQ-03`, `GOV-03`, `TST-03` |
| 1.3 | `scripts/validar_entrega.py`: validação local determinística, sem LLM e sem rede (rules por ID, ADR citada e vigente, estrutura da spec, lições lidas, fila vazia, SHA); entrega JSON com `ok` e a lista `julgamento` (o que sobra para o LLM); reusa `forge gate --json` | 1.1 | — | `FLX-06`, `DUP-03`, `SEG-03`, `TST-03` |
| 1.4 | `scripts/pr_guard.py --check`: bloqueia `gh pr create` sem `--draft`, `gh pr ready` e `gh pr merge` sem veredito conforme do SHA atual; consulta ao `uv` no context7 | 1.1 | com 1.3 | `SEG-01`, `SEG-03`, `SEG-07`, `FLX-08` |
| 1.5 | Instalador do hook `PreToolUse` (matcher `Bash`) em `hooks.go` e no `.claude/settings.json`; `doctor` checa `uv` e o hook | 1.4 | — | `SEG-02`, `ARQ-05`, `TST-02` |
| 1.6 | Sinais `spec-para-revisao` e `pronto-para-revisao` no `status.md` (leitura e validação em `internal/`); linha `Modo:` (com ou sem humano) dentro do teto de 1 KB (P-21) | 1.1 | com 1.4 | `FLX-06`, `ARQ-01`, `TST-03` |
| 1.7 | Modo sem humano: pergunta no início da sessão com aviso de riscos (nenhuma pessoa vê o código antes do PR sair do draft; erro do revisor passa; custo sem teto humano), padrão com humano; grava o modo | 1.1, 1.6 | — | `FLX-01`, `SEG-05`, `TST-03` |

## Onda 2 — Fluxo (split pelo Builder, fix com prioridade, perguntas, ordem do PR)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 2.1 | `forge queue review` e `pipeline.OpenPR`: veredito conforme não cria PR; PR existente é o padrão; `ready` só com SHA igual e [Continuar] | 1.1, 1.6 | — | `ARQ-02`, `DUP-03`, `FLX-03` |
| 2.2 | Revisor abre novo-fix por desvio (usa 1.3 e 1.2: mapeia o erro e grava a lição), devolve ao Builder, prioridade sobre feature nova; teto: padrão `max_rework_attempts` 3 (2 devoluções) na mesma conta do `Attempt`; ao estourar, pergunta ao dono | 1.2, 1.3, 2.1 | — | `FLX-06`, `DUP-01`, `TST-03` |
| 2.3 | Split pelo Builder sem nova aprovação: `dispatch` aceita job sem tasks; passo do Builder cria `features/`; checagem determinística de que toda task cita o `plan.md` e nada sai dele | 1.3 | com 2.2 | `FLX-01`, `FLX-02`, `TST-03` |
| 2.4 | Paralelismo do Builder para features e fixes: `PlanWaves` com `Builder.MaxParallelTasks` e o leque (`PrepararLeque`/`Juntar`); verificar a ligação do leque ao laço do goal | 2.3 | com 2.2 | `DUP-03`, `ARQ-02`, `TST-04` |
| 2.5 | Passos dos comandos e skills: `/discovery`, `/nova-feature`, `/novo-fix` abrem draft; [Seguir]/[Ajustar] e [Continuar]/[Ajustar]; `/split-features` vira passo do Builder; leitura só das lições da rule; templates e goldens | 1.5, 1.7, 2.1, 2.3 | — | `DUP-10`, `TST-02`, `COD-05`, `FLX-01` |
| 2.6 | ADR final (próxima livre) e redação nova do `FLX-01` e do `GOV-03` (`GOV-01`); `concepts.md` | 2.5 | — | `GOV-01`, `FLX-05` |
| 2.7 | `/upgrade-sdd` leva a política; `forge update` e goldens; nota de versão | 2.5, 2.6 | — | `TST-02`, `FLX-07` |

## Onda 3 — CI, atestado e adoção
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 3.1 | Workflow modelo: `if` de draft, `ready_for_review`, `concurrency` só em PR, matriz de 3 SO só em PR pronto e `main`, job "pronto?" só no check obrigatório da `main`; aplicar ao `ci.yml` do forge-sdd | — | com 3.3 | `ARQ-03`, `SEG-04` |
| 3.2 | Job de verificação do atestado (ubuntu, segundos): `scripts/atestado.py --verificar` confere veredito conforme do SHA do HEAD e a prova escolhida (Decisão vi); `validar_entrega.py` fora da Action | 1.1, Decisão vi | com 3.1 | `SEG-04`, `SEG-05`, `FLX-06` |
| 3.3 | `scripts/ci-politica.py`: lista repositórios, marca ok/fora, alerta de gasto e checagem mensal; sem token no git | — | com 3.1 | `SEG-04`, `SEG-05` |
| 3.4 | README: um prompt curto "cole isto na sua IA" por ferramenta; notas para Windows, Linux e macOS | 3.1, 3.2, 3.3 | — | `COD-05`, `GOV-04` |
| 3.5 | Prova: PR de teste draft para ready (jobs `skipped`, depois uma execução) e atestado inválido bloqueando, em cada repositório da primeira leva | 3.1, 3.2 | — | `TST-01` |
| 3.6 | Aplicar nos repositórios da primeira leva via `/upgrade-sdd`, com aprovação do dono por repositório | 2.7, 3.5 | — | `FLX-01` |

Caminho crítico: 1.1, 1.3, 2.1, 2.3, 2.5, 2.7, 3.6. As tasks 2.5 e 2.7 tocam templates e goldens que a 105 também altera; 2.5 também toca `split-features.md`. Sequenciar com a 105 para não regravar goldens duas vezes.
