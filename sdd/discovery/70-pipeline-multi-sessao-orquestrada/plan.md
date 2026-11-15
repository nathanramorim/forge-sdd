# Plano de Discovery 70 — Pipeline Multi-Sessão Orquestrada (v4.1.0)

Branch única `feat/70-pipeline-multi-sessao-orquestrada` (`FLX-02`), em worktree `../forge-sdd-70`, criada a partir da `feat/69` (que contém a 4.0.0-beta). O PR mira `feat/69` até ela ser mesclada. Pasta de execução: `sdd/discovery/70-pipeline-multi-sessao-orquestrada/features/`. Status: `FLX-01` aprovado; fatiada em 7 tasks e enfileiradas.

## Ordem e dependências
`70-01 -> 70-02 -> 70-03 -> 70-04 -> 70-05 -> 70-07`; `70-01 -> 70-06 -> 70-07`; `70-06` pode andar em paralelo a `70-02` e `70-03`.
Ordem de fila sugerida: 70-01, 70-02, 70-06, 70-03, 70-04, 70-05, 70-07.

## Quebra preliminar (Rule of 7: 7 features)
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 70-01 Layout e configuração do Forge | M | — | `.forge/` em `internal/layout`; `ForgeConfig` e `NotificationsConfig` em `internal/config` com defaults e validação; templates `forge.json`, `notifications.json` e `.forge/runtime/.gitignore`; recusa de segredo literal | `sdd-arquitetura`, `sdd-seguranca`; `ARQ-01`, `ARQ-02`, `DUP-01`, `DUP-04`, `SEG-01`, `SEG-04`, `TST-02`, `TST-03` | — |
| 70-02 Barramento de jobs | G | 70-01 | `internal/bus`: contrato do job, estados por diretório, `enqueue`, `claim` por `rename` com lease, prioridade e FIFO, `complete`, `fail`, `recover`; `forge-sdd bus` | `sdd-arquitetura`, `sdd-seguranca`, `sdd-codigo-e-testes`; `ARQ-01`, `SEG-01`, `SEG-02`, `COD-03`, `TST-03`, `TST-04` | `novo-subcomando-cli-sdd` |
| 70-03 Gate pré-build e `/implementar` por modo | M | 70-01, 70-02 | `forge-sdd dispatch`: gate determinístico (spec, quebra, refs de ADR e rules, `forge.json`), evidência de card e job; modo por `forge.json` em `/implementar` e `/implement` | `sdd-fluxo-e-portao`, `sdd-seguranca`; `FLX-01`, `FLX-08`, `DUP-02`, `SEG-01`, `TST-03` | `novo-subcomando-cli-sdd` |
| 70-04 Worker e Builder paralelo | G | 70-02, 70-03 | `forge-sdd worker --role builder`: laço, runner por perfil de CLI (`claude`, `agy`), planejador de ondas por `**Depende de:**`, `git worktree`, handoff para o Reviewer | `sdd-codigo-e-testes`, `sdd-seguranca`; `SEG-03`, `SEG-05`, `COD-01`, `COD-03`, `TST-04` | `novo-subcomando-cli-sdd` |
| 70-05 Reviewer e anti-loop | M | 70-02, 70-04 | agente `reviewer` e skill; `forge-sdd bus review` (conforme, `pr_rework` com `attempt`, escalonamento); feedback estruturado; `worker --role reviewer` | `sdd-fluxo-e-portao`; `FLX-06`, `FLX-08`, `SEG-06`, `TST-03` | — |
| 70-06 Notificações e inbox | M | 70-01 | `internal/notify` (inbox, toast por SO, e-mail SMTP, webhook); `forge-sdd inbox`; hook de leitura e comando `/inbox` nas 3 ferramentas | `sdd-seguranca`, `sdd-arquitetura`; `SEG-02`, `SEG-04`, `SEG-05`, `ARQ-10`, `TST-04` | — |
| 70-07 Launcher, scaffold, docs e release | G | 70-03, 70-04, 70-05, 70-06 | `forge-up.sh` e `forge-up.ps1`; skills (`concepcao` carrega ADRs e rules; `build`; `governanca`; `operacao`); templates, goldens, ADR 0004, `concepts.md`, release notes `4.1.0`, E2E simulado | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-02`, `TST-05` | — |

Rule of 7: 7 tasks em três ondas lógicas (base, motor, entrega). A Governança pode exigir dividir `70-02`, `70-04` ou `70-07` por serem `G`.

## Critério final
`go build ./... && go vet ./... && go test -race ./...` e `forge-sdd gate` sem falhas, mais `go test ./internal/pipeline -run TestE2E_PipelineCompleto`.

## Decisões FLX-01 incorporadas (2026-11-13)
Decisões 1 a 5 do `discovery.md` aceitas como propostas. Acréscimo do fatiamento: o launcher fica em `forge-sdd up` (Go) com `forge-up.sh` e `forge-up.ps1` como envoltórios finos (D-8, `criteria.md`), e o scaffold entrega só arquivos `.example`, de modo que o modo pipeline nunca é ativado sem `forge-sdd pipeline init`. `duvida` do Reviewer escala para a pessoa, sem rework automático.

## Medição
Registrar no PR o conjunto de leitura por job (spec, IDs, rules) frente ao fluxo de sessão única, em bytes, como foi feito no ADR 0003.

## Pós-Portão
Após `FLX-01`: `/split-features` grava as specs em `features/` e enfileira em ordem de dependência. Camada B por subagente `governance`, uma rodada por bloco de tasks. Versão `4.1.0`: bump no commit de release, tag `v4.1.0` (estável, sem `-beta`) só após a mescla da `feat/69`; antes disso, `v4.1.0-beta` na branch (Regra 11).
