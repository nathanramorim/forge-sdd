# Plano de Discovery 72 — Visibilidade, tmux e Entrega ao Builder (v4.2.0)

Branch única `feat/72-visibilidade-tmux-e-entrega-ao-builder` (`FLX-02`), em worktree `../forge-sdd-72`, empilhada na `feat/71`. O PR mira `feat/71`; a `4.2.0-beta` sai com as duas. Pasta de execução: `sdd/discovery/72-visibilidade-tmux-e-entrega-ao-builder/features/`. Status: `FLX-01` pendente.

## Ordem e dependências
`72-01` e `72-04` são independentes e podem andar em paralelo; `72-02 -> 72-03 -> 72-05`; `72-04 -> 72-05`; `72-01 -> 72-05`.
Ordem de fila sugerida: 72-01, 72-02, 72-03, 72-04, 72-05.

## Quebra preliminar (5 features)
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 72-01 Entrega ao Builder e despacho sem card | M | — | `dispatch` sem evidência quando `require_existing_card: false`; guarda de modo em `queue next`/`complete` (`FORGE_WORKER` ou `--local`); skills `split-features` (passo de `dispatch`) e `implementar`/discovery (a Discovery não executa); templates e goldens | `sdd-fluxo-e-portao`, `sdd-seguranca`; `FLX-01`, `FLX-08`, `SEG-01`, `TST-03` | `novo-subcomando-cli-sdd` |
| 72-02 Visão limpa do worker | G | 72-01 | `LiveView` (`io.Writer`): log bruto completo, eventos estruturados de claude, gemini e codex, máscara de segredos com os padrões do gate, filtro de ruído, `--raw`; acompanhamento por `git status` e batimento; flags de saída estruturada nos perfis | `sdd-seguranca`, `sdd-codigo-e-testes`; `SEG-05`, `DUP-01`, `COD-01`, `COD-03`, `TST-04` | — |
| 72-03 Falha legível e pausa por limite de uso | M | 72-02 | motivo com a última linha útil; classificação de limite de uso por CLI; `bus.Requeue`; pausa e retomada com teto; aviso `info` no inbox | `sdd-codigo-e-testes`, `sdd-seguranca`; `SEG-01`, `SEG-05`, `COD-03`, `TST-03`, `TST-04` | — |
| 72-04 Abrir pelo tmux | M | — | `internal/pipeline/tmux.go`: plano de comandos, três painéis nomeados, attach ou switch-client, sessão existente; `up --terminal`; `terminal` em `forge.json`; menu e checagem | `sdd-seguranca`, `sdd-codigo-e-testes`; `SEG-01`, `SEG-03`, `COD-03`, `TST-04` | — |
| 72-05 Docs, ADR e release | M | 72-01, 72-03, 72-04 | `.forge/LEIA-ME.md`, README, `docs/pipeline.md` (tmux, visão limpa, limite de uso), ADR 0006, `concepts.md`, atualização da entrada `4.2.0` das release notes, goldens, E2E do worker com agente falso | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-02`, `TST-05` | — |

## Critério final
`go build ./... && go vet ./... && go test -race ./...` e `forge-sdd gate` sem falhas, mais um E2E em que o agente falso emite eventos de cada CLI e o worker mostra as linhas esperadas, e um E2E de limite de uso (job volta à fila, pausa, retoma).

## Medição
Registrar no PR o número de linhas que o worker mostra por job frente ao log bruto, em bytes, e o tempo até a primeira linha útil na tela (hoje: nenhuma até o fim).

## Pós-Portão
Após `FLX-01`: `/split-features` grava as specs e enfileira. Camada B **por task**, só sobre o diff, em paralelo entre tasks independentes. Release: `4.2.0-beta` reunindo os discoveries 71 e 72; tag na branch só com autorização (Regra 11).
