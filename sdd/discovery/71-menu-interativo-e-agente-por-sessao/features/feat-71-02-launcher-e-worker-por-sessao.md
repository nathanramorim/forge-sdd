# feat-71-02 — Launcher e Worker por Sessão

**Branch:** `feat/71-menu-interativo-e-agente-por-sessao`
**Fase:** 71-02
**Tamanho:** `M`
**Depende de:** 71-01
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
O launcher abre as três sessões com uma CLI só, e o worker usa um perfil único. Esta task faz cada sessão usar o agente escolhido para ela, sem mudar a abertura por sistema operacional.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão de código existente em `internal/pipeline` e `cmd/forge-sdd`, sem pacote novo; a concorrência e o launcher por SO já existem e têm teste (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Launcher por SO | `internal/pipeline/launcher.go` (`Sessions`, `LaunchPlan`, `terminalArgv`) | Uma CLI para as três; comando da Discovery fixo por CLI | estender | `Sessions` recebe o agente de cada sessão e usa `interactive_args` do perfil; a abertura por SO não muda |
| Worker | `internal/pipeline/worker.go`, `cmd/forge-sdd/worker.go` | Usa o perfil de `cfg.CLI` | estender | usa `cfg.CLIFor(papel)` e o nível de permissão |
| Comando `up` | `cmd/forge-sdd/up.go` | `--cli` e `--dry-run` | estender | `--discovery`, `--builder`, `--reviewer`; `--cli` continua e vale para as três |
| Scripts de abertura | `.forge/forge-up.sh`, `.forge/forge-up.ps1` (templates) | Envoltórios finos que chamam `forge-sdd up` | estender | repassam as novas flags sem lógica própria (`DUP-06`) |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-03`, `SEG-05`, `COD-01`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.
- Skill: `novo-subcomando-cli-sdd` (flags novas em subcomando existente).

## Escopo
1. `Sessions(env, sessions)` monta `forge-discovery` com o comando interativo do agente escolhido (prompt inicial fixo que manda seguir `.agents/skills/sdd-concepcao`; `claude` e `copilot` ainda recebem `--agent`/`-n`) e `forge-builder`/`forge-reviewer` com `forge-sdd worker --role <papel> --name <sessão>`. Título ANSI mantido para todas as CLIs.
2. `forge-sdd worker` resolve o perfil por `cfg.CLIFor(papel)` e expande `{permissions}` conforme `cfg.Permissions`; o log do worker informa o agente e o nível de permissão ao iniciar (sem imprimir o prompt nem o ambiente).
3. `forge-sdd up [--cli <c>] [--discovery <c>] [--builder <c>] [--reviewer <c>] [--dry-run]`: cada flag sobrescreve `sessions`; `--cli` preenche as três; sem nenhuma, vale `sessions` ou `cli`. Valor fora da lista fechada é erro com a lista aceita.
4. `forge-up.sh` e `forge-up.ps1` repassam os argumentos recebidos.
5. Sem regressão: `forge-sdd up --cli claude` e `--cli agy` da 4.1 geram os mesmos comandos de antes.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./cmd/forge-sdd/...
```
- `TestSessions_AgentePorSessao`: `discovery=claude`, `builder=codex`, `reviewer=gemini` geram os três comandos corretos por SO (macOS, Linux, Windows), sem shell, com título ANSI.
- `TestWorker_PerfilDaSessao`: runner falso confere o argv do `builder` e do `reviewer` com agentes diferentes e o nível de permissão.
- `TestUp_FlagsPorSessao`: precedência (`flag` > `sessions` > `cli`), valor inválido recusado, `--dry-run` não abre nada.
- Compatibilidade: os testes de launcher da 4.1 continuam passando sem edição.
- `bash -n .forge/forge-up.sh` passa; o `.ps1` é validado por teste de conteúdo.
