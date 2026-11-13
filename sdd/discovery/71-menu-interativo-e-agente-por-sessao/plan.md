# Plano de Discovery 71 — Menu Navegável e Agente por Sessão (v4.2.0)

Branch única `feat/71-menu-interativo-e-agente-por-sessao` (`FLX-02`), em worktree `../forge-sdd-71`, criada a partir da `feat/70`. O PR mira `feat/70` até ela ser mesclada. Pasta de execução: `sdd/discovery/71-menu-interativo-e-agente-por-sessao/features/`. Status: `FLX-01` aprovado; fatiada em 5 tasks e enfileiradas.

## Ordem e dependências
`71-01 -> 71-02 -> 71-04`; `71-01 -> 71-03 -> 71-04`; `71-05` depende de `71-02`, `71-03` e `71-04`.
`71-02` e `71-03` podem andar em paralelo depois de `71-01`.
Ordem de fila sugerida: 71-01, 71-02, 71-03, 71-04, 71-05.

## Quebra preliminar (5 features)
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 71-01 Perfis por agente e `sessions` em `forge.json` | M | — | `internal/config`: perfis `claude`, `gemini`, `copilot`, `codex` (comando não interativo e interativo), campo `sessions` (agente por sessão), `CLIFor(sessão)`, validação e migração de `forge.json` da 4.1 sem `sessions`; `agy` continua aceito | `sdd-arquitetura`, `sdd-seguranca`; `ARQ-01`, `DUP-01`, `SEG-03`, `SEG-04`, `TST-02`, `TST-03` | — |
| 71-02 Launcher e worker por sessão | M | 71-01 | `internal/pipeline`: `Sessions` monta cada sessão com o agente dela; worker usa o perfil da sua sessão; `forge-sdd up --discovery/--builder/--reviewer`; `forge-up.sh` e `.ps1` repassam as flags | `sdd-codigo-e-testes`, `sdd-seguranca`; `SEG-01`, `SEG-03`, `COD-03`, `TST-04` | `novo-subcomando-cli-sdd` |
| 71-03 Checagem de agente antes de abrir | M | 71-01 | `internal/pipeline` ou `internal/doctor`: CLI no PATH, arquivos do agente no projeto, `forge.json` válido; resultado estruturado (OK, aviso, bloqueio) e oferta de `forge-sdd update --agent` | `sdd-seguranca`, `sdd-arquitetura`; `SEG-01`, `DUP-01`, `TST-03` | — |
| 71-04 Menu navegável e seletor | G | 71-02, 71-03 | `forge` sem argumentos: menu em dois caminhos com `huh`, seletor de agente por sessão que lembra a última escolha, ações que chamam as funções dos subcomandos, detecção de TTY e saída limpa sem terminal | `sdd-arquitetura`, `sdd-codigo-e-testes`; `ARQ-01`, `ARQ-02`, `DUP-01`, `COD-01`, `TST-04` | `novo-subcomando-cli-sdd` |
| 71-05 Docs, scaffold e release | M | 71-02, 71-03, 71-04 | `.forge/LEIA-ME.md`, README, `docs/pipeline.md`, ADR 0005, `concepts.md`, release notes `4.2.0`, goldens, E2E do menu com entrada simulada | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-02`, `TST-05` | — |

## Critério final
`go build ./... && go vet ./... && go test -race ./...` e `forge-sdd gate` sem falhas, mais um teste de ponta a ponta do menu com entrada simulada (escolher Pipeline, um agente por sessão, e conferir os três comandos gerados).

## Medição
Registrar no PR o número de passos para iniciar o pipeline antes e depois (hoje: `pipeline init`, editar `forge.json`, `up --cli`; depois: `forge`, escolher, confirmar).

## Pós-Portão
Após `FLX-01`: `/split-features` grava as specs em `features/` e enfileira. Camada B **por task**, só sobre o diff e os IDs citados, em paralelo entre tasks independentes (lição da 70: o lote de 7 tasks levou 17 minutos). Versão `4.2.0-beta`: bump no commit de release e tag `v4.2.0-beta` na branch (Regra 11) só com autorização.
