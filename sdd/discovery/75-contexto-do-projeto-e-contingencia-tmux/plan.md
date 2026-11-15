# Plano de Discovery 75 — Contexto do Projeto e Contingência do tmux

Branch única `feat/75-contexto-do-projeto-e-contingencia-tmux` (`FLX-02`), em worktree `../forge-sdd-74`, partindo da `feat/72` (PR empilhado sobre o #71). Pasta de execução: `sdd/discovery/75-contexto-do-projeto-e-contingencia-tmux/features/`.

## Ordem e dependências
`75-01 -> 75-02 -> 75-03`; `75-04 -> 75-05`; `75-01, 75-03, 75-05 -> 75-06`. As duas trilhas (contexto: 01-03; terminal: 04-05) não tocam os mesmos arquivos e podem andar em paralelo; 02 e 03 mexem nos mesmos textos de skill e rodam em sequência; 04 e 05 mexem no `launcher.go`/`tmux.go` e rodam em sequência.
Ordem de fila: 75-01, 75-04, 75-02, 75-05, 75-03, 75-06.

## Quebra preliminar (6 features)
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 75-01 `forge-sdd brief` | M | — | `internal/gate/brief.go` (seções, orçamento, ADRs vigentes, estado `rascunho/pronta`), `cmd/forge-sdd/brief.go` (`--check`, `--max-bytes`), testes e fixtures | `sdd-arquitetura`, `sdd-codigo-e-testes`; `ARQ-01`, `DUP-01`, `COD-01`, `TST-01`, `TST-04` | `novo-subcomando-cli-sdd` |
| 75-02 Protocolo de abertura da Discovery | M | 75-01 | `DiscoveryPrompt`; skill `sdd-concepcao` e `references/` (discovery, nova-feature, novo-fix, split-features); seção `Decisões do projeto aplicadas`; `CLAUDE.md`/`GEMINI.md`/`AGENTS.md` do scaffold; cópia em `.agents/`; goldens | `sdd-fluxo-e-portao`, `sdd-reuso-e-duplicacao`; `FLX-01`, `GOV-01`, `DUP-02`, `TST-02` | — |
| 75-03 Constituição primeiro | M | 75-01, 75-02 | `/constitution` (3 agentes) povoando módulos, fluxos, decisões, ADRs e segurança e fechando com `brief --check`; passo 1 no `init` e no menu; avisos no `Preflight` e no `doctor` | `sdd-fluxo-e-portao`, `sdd-codigo-e-testes`; `FLX-01`, `COD-01`, `TST-02`, `TST-04` | `novo-subcomando-cli-sdd` |
| 75-04 Instalar o tmux e cair com graça | G | — | `internal/pipeline/tmuxinstall.go` (tabela, `Install`), decisão de terminal (instalar, recusar e lembrar, sem TTY), queda do tmux em `auto`, aviso no `Preflight`, menu e `up` | `sdd-seguranca`, `sdd-codigo-e-testes`; `SEG-01`, `SEG-03`, `SEG-07`, `COD-03`, `TST-04` | — |
| 75-05 Windows Terminal com painéis | P | 75-04 | `terminalArgv`/`LaunchPlanFor` com `new-tab` + `split-pane`; sem `wt.exe`, caminho atual; testes do argv | `sdd-seguranca`, `sdd-codigo-e-testes`; `SEG-01`, `COD-03`, `TST-04` | — |
| 75-06 Docs, ADR e release | M | 75-01, 75-03, 75-05 | README, `docs/pipeline.md`, `.forge/LEIA-ME.md`, `guia-pipeline.html` (tmux, contingência, brief), ADR 0007, `concepts.md`, release notes, goldens, E2E de ponta a ponta | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-05` | — |

## Critério final
CA-01 a CA-11 de `criteria.md`; `go build ./... && go vet ./... && go test -race ./...` e `forge-sdd gate` verdes; E2E em projeto temporário recém-gerado: `brief --check` dá `rascunho`, depois de povoar dá `pronta`; E2E do terminal com instalador falso (instala, recusa, sem TTY, falha).

## Medição
Em 3 sessões Discovery de um mesmo projeto, contar quantas vezes a pessoa precisou repetir uma decisão já gravada (meta: 0) e anotar no PR.

## Pós-Portão
Após `FLX-01`: `/split-features` grava as specs e enfileira. Camada B por task, só sobre o diff. Release: bump a decidir (`4.2.1-beta` ou `4.3.0-beta`); tag na branch só com autorização (Regra 11).
