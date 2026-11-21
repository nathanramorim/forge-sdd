# Plano — 106 (ondas 0 a 3 e a onda P da Constituição; spec volta à principal, branch e PR em draft por feature)

Ordem: base (adaptador, modo, estado) primeiro; depois o mecanismo; depois o fluxo; CI, atestado e adoção por último. Nada roda sem a reaprovação do escopo (`FLX-01`). Cada task traz teste negativo (`TST-03`) e passa em `forge gate --criterios`. Até 7 tasks por onda. Marca: **[IA]**, **[HUMANO]**, **[AUTO]** indica quem executa o passo que a task constrói. **Escopo do fecho aprovado em 2026-11-20**: a 108 (prioridade 0) exige parte das tasks; veja `## Fecho da 108 (prioridade 0)` ao fim. A edição da Constituição está **toda** na task 0.8 (PR único); nenhuma outra task a edita.

## Onda 0 — Base (adaptador de VCS, modo do atestado, estado por feature)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 0.1 | Verificar o que está "a verificar": documentação (context7) do `az repos pr` e do MCP do Azure DevOps (draft, pronto, merge, autenticação interativa), build validation em draft, Environment com aprovadores, autor aprovando o próprio PR; medir tokens MCP versus CLI; escrever o resultado na spec | — | com 0.2 | `SEG-07`, `DUP-02` |
| 0.2 | `internal/vcs`: interface `Adaptador` (criar PR em draft, procurar PR aberto, marcar pronto, mesclar), leitura de `VCS / Work Item System`, adaptadores `github` (move a lógica de `pipeline/pr.go` e `queue/pr.go`) e `nenhum` (branch pronta, merge local); valor desconhecido recusado | — | com 0.1 | `ARQ-01`, `ARQ-02`, `DUP-01`, `SEG-03`, `TST-03` |
| 0.3 | **[ADIADA: fora do fecho da 108; depende da 0.1 e de prova com conta real]** Adaptador `azure-devops` (`az repos pr` só no adaptador; o MCP é orientado pela skill); `AllowCommands` aceita `az` apenas nele; teste sem `gh` no PATH | 0.1, 0.2 | — | `SEG-03`, `SEG-07`, `TST-03` |
| 0.4 | `doctor` verifica o VCS declarado, a ferramenta (`gh`, `az` ou MCP), os campos de modo da Constituição criados na 0.8 (presença e valor válido), informa que `azure-devops` está "não verificado" e que o modo `vcs` do atestado está adiado, e informa **qual agente tem hook** (só o Claude Code; Gemini e Copilot dependem dos passos dos comandos e da pipeline) e avisa se faltar | 0.2, 0.5, 0.8 | com 0.3 | `ARQ-05`, `TST-03` |
| 0.5 | Leitor único dos campos de modo da tabela da Constituição, criados na 0.8: `Modo do atestado` (`desligado`, `local` padrão, `vcs`; **`vcs` é aceito como valor e recusado com a mensagem "modo adiado"**) e `Modo de revisão humana` (`com-humano` padrão, `sem-humano`), **lidos da branch base** (Decisões xiv e xv, aprovadas); a branch da feature não os sobrepõe; valor inválido recusado; o modo efetivo vai ao `status.md` e ao veredito; o `WIP humano` (107-12) e os campos do quadro (108) usam o mesmo leitor | 0.2, 0.8 | com 0.1 | `ARQ-10`, `SEG-01`, `GOV-01`, `TST-03` |
| 0.6 | **Cadeia de fallback do atestado só para os VCS `github` e `nenhum`** (decisão do dono, 2026-11-20): lê o modo (0.5); o `local` é o fim da cadeia, com tempo limite; sem resposta, pergunta ao humano (no modo sem humano, avisa e grava); nunca aprova em silêncio; `vcs` é recusado com "modo adiado" (a perna da prova do VCS e a alternativa assinada ficam na 3.5, adiada); o adaptador `azure-devops` é reconhecido e recusa a operação com a mensagem "não verificado" (a perna depende da 0.3 e de prova com conta real, fora do fecho; o `forge doctor` informa na 0.4); textos do que se perde por modo | 0.2, 0.5 | — | `FLX-01`, `SEG-05`, `TST-03` |
| 0.7 | Estado de execução sem conflito (Decisão xvi, **aprovada: B**): arquivo de estado por feature na branch da feature (`sdd/discovery/<id>/state/<sid>.json`); claim pela existência da branch; fila calculada na leitura a partir da principal (`plan.md`, `features/` e estados); `queue.json` vira plano; cobre `layout.Current` (maior prefixo existente; alimenta `ActiveInitiative`, `QueuePath`, `StatusPath`), `goalSpecApproved` (marcador `Gate 1: aprovado` no `status.md`), `archive` (`branchGuard` e leitura a partir da principal) e os mutadores (`SetTaskStatuses`, `BlockTasks`, `Review`, consumo da fila); teste que reproduz o conflito de `updated_at` | 0.2 | com 0.3 | `DUP-04`, `DUP-01`, `ARQ-02`, `TST-03` |

## Onda P — PR único da Constituição e da rule de modos (primeira do fecho; só a task 0.8)
Fica fora da tabela da Onda 0 porque ela já tem 7 tasks (teto por onda). Número 0.8 por ordem inicial; no fecho é a primeira da onda 1.
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 0.8 | **PR único da Constituição e da rule de modos** (ADR-0022 e `GOV-01`): todas as edições de `sdd/memory/constitution.md` num só PR, mais a rule `.agents/rules/sdd-modos-de-projeto.md` e o ajuste de `sdd-fluxo-e-portao.md`; conta de bytes em `discovery.md`, seção "PR único da Constituição" (alvo 9.182 de 9.216 bytes, medidos com `wc -c`) | — | com 0.2 e 1.1 | `GOV-01`, `GOV-04`, `FLX-05`, `TST-03` |

## Onda 1 — Mecanismo (contrato, scripts locais, hooks)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 1.1 | `harness.Parecer` ganha `head_sha`, `ponto` (`spec` ou `entrega`), `modo` e `feature`; `persistVerdict` os grava a partir do HEAD da branch; veredito sem SHA recusado | — | com 1.2 | `DUP-06`, `FLX-06`, `SEG-01`, `TST-03` |
| 1.2 | `scripts/licoes.py` (`uv`, `pathlib`) + `sdd/memory/licoes/` + teste; `forge lessons add` grava um arquivo por lição; migra as 2 entradas | — | com 1.1 | `DUP-01`, `ARQ-03`, `GOV-03`, `TST-03` |
| 1.3 | **[AUTO]** `scripts/validar_entrega.py --ponto spec\|entrega [--feature]`: local, sem LLM e sem rede (rules por ID, ADR citada e vigente, estrutura da spec, lições lidas, fila da feature vazia, SHA); JSON com `ok` e `julgamento` (sobra para o LLM); reusa `forge gate --json` | 1.1 | — | `FLX-06`, `DUP-03`, `SEG-03`, `TST-03` |
| 1.4 | **[AUTO]** `scripts/pr_guard.py --check`: bloqueia criar PR sem draft, marcar pronto e mesclar sem veredito conforme do SHA daquele PR; cobre `gh pr`, `az repos pr` e MCP (a verificar); vale só no Claude Code, e o texto de aviso diz que nos outros agentes valem os passos dos comandos e o job "pronto?" da pipeline | 0.2, 1.1 | com 1.3 | `SEG-01`, `SEG-03`, `SEG-07`, `FLX-08` |
| 1.5 | **[AUTO]** Instalador do hook `PreToolUse` em `hooks.go` (hoje sem `matcher`); `doctor` checa `uv` e o hook e imprime o aviso de cobertura por agente | 1.4 | — | `SEG-02`, `ARQ-05`, `TST-02` |
| 1.6 | Sinais `spec-para-revisao` (iniciativa) e `pronto-para-revisao` (por feature); linhas `Modo:` e `Atestado:` no `status.md` (1 KB); teto por feature (`Attempt`, padrão 3) | 1.1, 0.7 | com 1.4 | `FLX-06`, `ARQ-01`, `TST-03` |
| 1.7 | **[HUMANO]** Modo sem humano: pergunta no início com aviso de riscos; padrão com humano; grava o modo | 1.1, 1.6 | — | `FLX-01`, `SEG-05`, `TST-03` |

## Onda 2 — Fluxo (spec na principal, feature por branch, novo-fix)
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 2.1 | `queue review` e `pipeline.OpenPR` delegam ao adaptador: veredito conforme não cria PR; "marcar pronto" só com SHA igual e [Continuar]; sem VCS, só libera a branch | 0.2, 1.1, 1.6 | — | `ARQ-02`, `DUP-03`, `FLX-03` |
| 2.2 | **[IA/AUTO/HUMANO]** Spec volta à principal: PR da spec em draft, Conferência 1, [Seguir], merge da spec pelo adaptador após o [Seguir] (Decisão xviii, aprovada; automático no modo sem humano; merge local no `nenhum`); marcador `Gate 1: aprovado` no `status.md` (substitui `queue.json` em `goalSpecApproved`); `archive` roda a partir da principal | 0.7, 1.3, 2.1 | — | `FLX-01`, `FLX-02`, `TST-03` |
| 2.3 | **[IA]** Split pelo Builder sem nova aprovação: `dispatch` aceita job sem tasks; confere que toda feature cita o `plan.md` e nada sai dele | 1.3, 2.2 | com 2.4 | `FLX-01`, `FLX-02`, `TST-03` |
| 2.4 | **[IA/AUTO]** Uma branch e um PR em draft por feature: base sempre `base_branch`; `ResolveBase` e `OpenPR` pelo adaptador; leque com worktree por task; `Juntar` vira o merge local do `nenhum` e dá lugar a `Publicar` nos demais VCS (Decisão xxii, aprovada: `Base` = principal; branch da task = `feat/<id>-<nome>-<sid>`; `Juntar` só no merge local do `nenhum`; `Publicar` nos demais; menu de conflito mantido; testes de `leque_test.go` migram; sem chamador fora de `leque.go`); feature dependente espera a predecessora ser mesclada (Decisão xx, aprovada); [Continuar] agrupado por onda (Decisão xix, aprovada) | 0.7, 2.1, 2.2 | com 2.3 | `DUP-03`, `ARQ-02`, `TST-04` |
| 2.5 | **[IA]** Novo-fix aberto pelo reviewer por desvio (usa 1.3 e 1.2), prioridade sobre feature nova, paralelismo de features e fixes por `PlanWaves` e `MaxParallelTasks`; teto por feature; estourou, pergunta ao dono | 1.2, 1.3, 2.4 | — | `FLX-06`, `DUP-01`, `TST-03` |
| 2.6 | Passos dos comandos e skills: `/discovery`, `/nova-feature`, `/novo-fix` só da spec; Builder por feature; perguntas [Seguir]/[Ajustar] e [Continuar]/[Ajustar]; leitura só das lições da rule; templates e goldens (sequenciar com a 105 e a 107) | 1.5, 1.7, 2.2, 2.3, 2.4 | — | `DUP-10`, `TST-02`, `COD-05`, `FLX-01` |
| 2.7 | ADR final dos fluxos (próxima livre, numerada no merge; **não edita a Constituição**: a ADR-0022 e todas as edições da Constituição saem na 0.8); `concepts.md` (dono de PR: `internal/vcs`; dono do estado: o arquivo por feature); textos dos comandos e templates que citam os modos apontam para a rule `sdd-modos-de-projeto` (criada na 0.8), sem repetir o conteúdo | 2.6, 0.8 | — | `GOV-01`, `FLX-05` |

## Onda 3 — CI, atestado e adoção
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 3.1 | Workflow modelo por VCS: `github` com `if` de draft, `ready_for_review`, `concurrency` só em PR, matriz de 3 SO só em PR pronto e principal, job "pronto?" só no check obrigatório da principal; Azure Pipelines e build validation (a verificar na 0.1); `nenhum` nada; aplicar ao `ci.yml` do forge-sdd | 0.1 | com 3.3 | `ARQ-03`, `SEG-04` |
| 3.2 | **[AUTO]** Verificação do atestado **por PR de feature, só no modo `local`** (job ubuntu, segundos): `scripts/atestado.py --verificar` nos modos `desligado` e `local`, conforme o modo (0.5) e a cadeia (0.6); `--modo vcs` sai com código diferente de zero e a mensagem "modo adiado" (a prova do Environment com aprovadores fica adiada, junto da 3.5); PR da spec **sempre** em `local` (Decisão xxi, aprovada: o campo da Constituição vale só para PRs de feature); `validar_entrega.py` fora da pipeline | 0.5, 0.6, 1.1 | com 3.1 | `SEG-04`, `SEG-05`, `FLX-06` |
| 3.3 | **[AUTO]** `scripts/ci-politica.py` (`uv`, `pathlib`), **somente leitura** por padrão, saída em CSV e em texto simples, token só por variável de ambiente (nunca no git; nenhum arquivo gravado sem pedido). **Sub-etapa inicial:** repetir o que o hub testou em 20/11/2026 e registrar o resultado na spec: uso de cobrança da organização com o token existente; da conta pessoal com escopo `user` em token clássico; `actions/runs/{id}/jobs` para o tempo (o `timing` veio zerado). Hoje é "testado pelo hub em 20/11/2026, a repetir pela 3.3". **Entrega:** relatório com **as duas fontes (cobrança e cálculo por jobs) e a diferença**, nunca uma só (ausência rotulada); lista repositórios e marca ok/fora; alerta de gasto e checagem mensal; minutos equivalentes de Linux por job (duração arredondada ao minuto vezes 1x/2x/10x); `cancelled` e `skipped` à parte; métrica normalizada por PR mesclado e por push; `--congelar-linha-de-base` grava o arquivo versionado com data, repositórios e método; método na seção "Medição da economia" do discovery | 0.1, 3.1 | com 3.2 | `SEG-04`, `SEG-05`, `TST-03` |
| 3.4 | README: um prompt curto "cole isto na sua IA" por ferramenta; notas para Windows, Linux e macOS | 3.1, 3.2, 3.3 | — | `COD-05`, `GOV-04` |
| 3.5 | **[ADIADA: fora do fecho da 108; é a prova por VCS e do modo `vcs`]** Prova em **cada VCS** (`github`, `azure-devops`, `nenhum`; os dois últimos sem GitHub): PR da spec em draft até o merge; PR de teste de feature draft para ready, jobs `skipped` e depois uma execução; atestado inválido bloqueando; Environment em repositório pessoal privado (fallback assinado); queda de fallback; coleta da linha de base e da etapa 1 (`concurrency`) com o `ci-politica.py` | 3.1, 3.2, 3.3 | — | `TST-01` |
| 3.6 | `/upgrade-sdd` leva a política; `forge update` e goldens; nota de versão | 2.7 | com 3.5 | `TST-02`, `FLX-07` |
| 3.7 | Aplicar nos repositórios da primeira leva via `/upgrade-sdd`, com aprovação do dono por repositório | 3.5, 3.6 | — | `FLX-01` |

Caminho crítico: 0.2, 0.7, 1.1, 2.1, 2.2, 2.4, 2.6, 2.7, 3.6, 3.7. As tasks 2.6, 2.7 e 3.6 tocam templates, goldens, `split-features.md`, `hooks.go` e `doctor`, que a **105** também altera (onda 1 e 2 dela), e a tabela de roteamento e o verificador da **107** (ela depende do script da 1.3 e do fluxo da 2.2 a 2.5): sequenciar 2.6 depois da onda 1 da 105 e a onda 2 da 107 depois da 1.3 e da 2.4.

## Fecho da 108 (prioridade 0)
Escopo do fecho aprovado em 2026-11-20 pelo dono (fecho completo; três correções de escopo). Ondas do analista, em ordem topológica; até 7 tasks por onda. Cada task só depende de tasks do fecho de ondas anteriores (conferido por script).

| Onda do fecho | Tasks da 106 | Depende de (dentro do fecho) |
|---|---|---|
| 1 | **0.8** (PR único da Constituição, primeira), 0.2, 1.1 | — |
| 2 | 0.5, 0.7, 1.3 | 0.2, 0.8, 1.1 |
| 3 | 1.6, 0.6 (só `github` e `nenhum`) | 1.1, 0.7, 0.2, 0.5 |
| 4 | 2.1, 3.2 (só `local`) | 0.2, 1.1, 1.6, 0.5, 0.6 |
| 5 | 2.2 | 0.7, 1.3, 2.1 |
| 6 | 2.3, 2.4 | 1.3, 2.2, 0.7, 2.1 |

**Ficam FORA do fecho:** 0.1, 0.3 e a perna Azure, 0.4, 1.2, 1.4, 1.5, 1.7, 2.5, 2.6, 2.7 (reduzida: ADR final dos fluxos e `concepts.md`), 3.1, 3.3, 3.4, 3.5 (adiada: prova por VCS e modo `vcs`), 3.6, 3.7. Da 107, a 108 exige 107-01, 107-02 e 107-12; a 107-12 depende da 0.5 e da 0.8.
