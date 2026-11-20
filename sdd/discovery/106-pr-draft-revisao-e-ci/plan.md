# Plano — 106 (4 ondas, de baixo para cima; escopo VCS agnóstico)

Ordem: contrato e scripts primeiro; depois o fluxo; CI e atestado por último. Nada roda sem a reaprovação do escopo (`FLX-01`). Cada task traz teste negativo (`TST-03`) e passa em `forge gate --criterios`. Até 7 tasks por onda.

## Onda 0 — Adaptador de VCS (base; o VCS vem da Constituição: `github`, `azure-devops`, `nenhum`)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 0.1 | Verificar o que está "a verificar" na tabela do discovery: documentação (context7) do `az repos pr` e do MCP do Azure DevOps (flags de draft, marcar pronto, merge, autenticação interativa), build validation em draft, aprovação em Environment e se o autor aprova o próprio PR; medir tokens MCP versus CLI; escrever o resultado na spec | — | com 0.2 | `SEG-07`, `DUP-02` |
| 0.2 | `internal/vcs`: interface `Adaptador` (criar PR em draft, procurar PR aberto, marcar pronto, mesclar), leitura de `VCS / Work Item System` da Constituição, adaptadores `github` (move a lógica de `pipeline/pr.go` e `queue/pr.go`) e `nenhum` (branch pronta, sem comando); valor desconhecido é recusado | — | com 0.1 | `ARQ-01`, `ARQ-02`, `DUP-01`, `SEG-03`, `TST-03` |
| 0.3 | Adaptador `azure-devops` (CLI `az repos pr`, só no adaptador; o MCP é orientado pela skill); `AllowCommands` aceita `az` apenas nele; teste sem `gh` no PATH | 0.1, 0.2 | — | `SEG-03`, `SEG-07`, `TST-03` |
| 0.4 | `doctor` verifica o VCS declarado, a ferramenta correspondente (`gh`, `az` ou MCP) e o modo de atestado configurado, e avisa se faltar | 0.2, 0.5 | com 0.3 | `ARQ-05`, `TST-03` |
| 0.5 | Configuração do modo de atestado (`desligado`, `local`, `vcs`; padrão a confirmar, Decisão xiv; local da configuração, Decisão xv), leitura e validação no adaptador; valor inválido é recusado; o modo efetivo vai ao `status.md` e ao veredito | 0.2 | com 0.1 | `ARQ-10`, `SEG-01`, `GOV-01`, `TST-03` |
| 0.6 | Cadeia de fallback no adaptador (`vcs`, alternativa assinada onde houver, `local`; tempo limite; pergunta ao humano ou aviso no modo sem humano; nunca aprova em silêncio) e texto de aviso do que se perde por modo | 0.3, 0.5 | — | `FLX-01`, `SEG-05`, `TST-03` |

## Onda 1 — Mecanismo (contrato, scripts locais, hooks)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 1.1 | `harness.Parecer` ganha `head_sha`, `ponto` (`spec` ou `entrega`) e `modo` (`com-humano` ou `sem-humano`); `persistVerdict` os grava a partir do HEAD; veredito sem SHA é recusado | — | com 1.2 | `DUP-06`, `FLX-06`, `SEG-01`, `TST-03` |
| 1.2 | `scripts/licoes.py` (`uv`, `pathlib`) + `sdd/memory/licoes/` + teste; `forge lessons add` grava um arquivo por lição; migra as 2 entradas atuais | — | com 1.1 | `DUP-01`, `ARQ-03`, `GOV-03`, `TST-03` |
| 1.3 | `scripts/validar_entrega.py`: validação local determinística, sem LLM e sem rede (rules por ID, ADR citada e vigente, estrutura da spec, lições lidas, fila vazia, SHA); entrega JSON com `ok` e a lista `julgamento` (o que sobra para o LLM); reusa `forge gate --json` | 1.1 | — | `FLX-06`, `DUP-03`, `SEG-03`, `TST-03` |
| 1.4 | `scripts/pr_guard.py --check`: lê o VCS da Constituição e bloqueia criar PR sem draft, marcar pronto e mesclar sem veredito conforme do SHA atual, cobrindo `gh pr ...`, `az repos pr ...` e as ferramentas do MCP (se o hook alcançar: a verificar); consulta ao `uv` no context7 | 0.2, 1.1 | com 1.3 | `SEG-01`, `SEG-03`, `SEG-07`, `FLX-08` |
| 1.5 | Instalador do hook `PreToolUse` em `hooks.go` (hoje `mergeClaudeHook` não grava `matcher`; passa a aceitar matcher para `Bash` e, se a 0.1 confirmar, para as ferramentas de MCP) e no `.claude/settings.json`; `doctor` checa `uv` e o hook | 1.4 | — | `SEG-02`, `ARQ-05`, `TST-02` |
| 1.6 | Sinais `spec-para-revisao` e `pronto-para-revisao` no `status.md` (leitura e validação em `internal/`); linhas `Modo:` (com ou sem humano) e `Atestado:` (modo efetivo e fallback usado) dentro do teto de 1 KB (P-21) | 1.1 | com 1.4 | `FLX-06`, `ARQ-01`, `TST-03` |
| 1.7 | Modo sem humano: pergunta no início da sessão com aviso de riscos (nenhuma pessoa vê o código antes do PR sair do draft; erro do revisor passa; custo sem teto humano), padrão com humano; grava o modo | 1.1, 1.6 | — | `FLX-01`, `SEG-05`, `TST-03` |

## Onda 2 — Fluxo (split pelo Builder, fix com prioridade, perguntas, ordem do PR)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 2.1 | `forge queue review` e `pipeline.OpenPR` delegam ao adaptador de VCS: veredito conforme não cria PR; PR existente é o padrão; "marcar pronto" só com SHA igual e [Continuar] (com `nenhum`, só libera a branch) | 0.2, 1.1, 1.6 | — | `ARQ-02`, `DUP-03`, `FLX-03` |
| 2.2 | Revisor abre novo-fix por desvio (usa 1.3 e 1.2: mapeia o erro e grava a lição), devolve ao Builder, prioridade sobre feature nova; teto: padrão `max_rework_attempts` 3 (2 devoluções) na mesma conta do `Attempt`; ao estourar, pergunta ao dono | 1.2, 1.3, 2.1 | — | `FLX-06`, `DUP-01`, `TST-03` |
| 2.3 | Split pelo Builder sem nova aprovação: `dispatch` aceita job sem tasks; passo do Builder cria `features/`; checagem determinística de que toda task cita o `plan.md` e nada sai dele | 1.3 | com 2.2 | `FLX-01`, `FLX-02`, `TST-03` |
| 2.4 | Paralelismo do Builder para features e fixes: `PlanWaves` com `Builder.MaxParallelTasks` e o leque (`PrepararLeque`/`Juntar`); verificar a ligação do leque ao laço do goal | 2.3 | com 2.2 | `DUP-03`, `ARQ-02`, `TST-04` |
| 2.5 | Passos dos comandos e skills: `/discovery`, `/nova-feature`, `/novo-fix` abrem draft; [Seguir]/[Ajustar] e [Continuar]/[Ajustar]; `/split-features` vira passo do Builder; leitura só das lições da rule; templates e goldens | 1.5, 1.7, 2.1, 2.3 | — | `DUP-10`, `TST-02`, `COD-05`, `FLX-01` |
| 2.6 | ADR final (próxima livre; inclui a decisão do adaptador de VCS), redação nova do `FLX-01` e do `GOV-03` e da nota da linha 30 de `VCS / Work Item System` (`GOV-01`); `concepts.md` (dono de PR: `internal/vcs`) | 2.5 | — | `GOV-01`, `FLX-05` |
| 2.7 | `/upgrade-sdd` leva a política; `forge update` e goldens; nota de versão | 2.5, 2.6 | — | `TST-02`, `FLX-07` |

## Onda 3 — CI, atestado e adoção
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 3.1 | Workflow modelo por VCS: no `github`, `if` de draft, `ready_for_review`, `concurrency` só em PR, matriz de 3 SO só em PR pronto e `main`, job "pronto?" só no check obrigatório da `main`; aplicar ao `ci.yml` do forge-sdd. No `azure-devops`, modelo de Azure Pipelines e política de branch (build validation; comportamento em draft: a verificar na 0.1); no `nenhum`, nada | 0.1 | com 3.3 | `ARQ-03`, `SEG-04` |
| 3.2 | Verificação do atestado (job ubuntu, segundos): `scripts/atestado.py --verificar` confere veredito conforme do SHA do HEAD e a prova conforme o modo (0.5): `desligado` não confere; `local` confere veredito e SHA; `vcs` confere a prova (Decisões vi e xi: GitHub, Environment com revisor; Azure DevOps, Environment do Azure Pipelines com aprovadores; `nenhum`, só local), com a queda de fallback da 0.6 sem aprovar em silêncio; `validar_entrega.py` fora da pipeline | 0.1, 1.1, Decisão xi | com 3.1 | `SEG-04`, `SEG-05`, `FLX-06` |
| 3.3 | `scripts/ci-politica.py`: lista repositórios, marca ok/fora, alerta de gasto e checagem mensal; sem token no git | — | com 3.1 | `SEG-04`, `SEG-05` |
| 3.4 | README: um prompt curto "cole isto na sua IA" por ferramenta; notas para Windows, Linux e macOS | 3.1, 3.2, 3.3 | — | `COD-05`, `GOV-04` |
| 3.5 | Prova em **cada VCS suportado** (`github`, `azure-devops`, `nenhum`; o teste de `azure-devops` e `nenhum` roda sem GitHub): PR de teste draft para ready, mais a queda de fallback com a prova do VCS ausente (cai no local, avisa e pergunta) (jobs `skipped`, depois uma execução), atestado inválido bloqueando e **Environment com revisor obrigatório em repositório pessoal privado** (se o plano não permitir, fallback commit/tag assinado), em cada repositório da primeira leva | 3.1, 3.2 | — | `TST-01` |
| 3.6 | Aplicar nos repositórios da primeira leva via `/upgrade-sdd`, com aprovação do dono por repositório | 2.7, 3.5 | — | `FLX-01` |

Caminho crítico: 0.2, 1.1, 1.3, 2.1, 2.3, 2.5, 2.7, 3.6. As tasks 2.5 e 2.7 tocam templates e goldens que a 105 também altera; 2.5 também toca `split-features.md`. Sequenciar com a 105 para não regravar goldens duas vezes.
