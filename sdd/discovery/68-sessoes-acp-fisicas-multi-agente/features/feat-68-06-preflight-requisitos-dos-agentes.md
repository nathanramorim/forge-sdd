# feat-68-06 — Detecção de Requisitos dos Agentes (`internal/preflight`) e Reuso em doctor/Studio

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-06
**Tamanho:** `M`
**Depende de:** `feat-68-01` (config `acp.*` e presets). Alimenta `feat-68-03`, `feat-68-04` e `feat-68-06b`.
**Status:** `approved` (FLX-01 aprovado)

> Fatiamento: esta task entrega apenas a detecção reutilizável (`Check`/`Plan`, relatório, doctor e aviso do Studio). Seleção interativa, instalação com confirmação e `--yes` estão em `feat-68-06b`.

## Contexto & Motivação
Os adaptadores ACP exigem CLIs e runtimes (por exemplo `npx` para `claude-agent-acp` e `codex-acp`). Hoje nada verifica isso, e uma estação falharia em runtime sem explicação. O preflight pergunta quais agentes o usuário vai usar (sem instalar todos), detecta o que falta, mostra o comando exato de instalação e só executa com confirmação explícita.

## Comportamento
1. **Seleção (init e update):** multisseleção entre os adaptadores de `acp.adapters` (presets: copilot, claude, gemini, codex, agy, e outros definidos pelo usuário). Os escolhidos são gravados em `acp.enabledAgents` via a API de config da 68-01. Só eles têm sessões no pool.
2. **Detecção:** para cada agente escolhido, `LookPath` do `command` do adaptador, versão (quando o preset define `versionArgs`) e requisitos declarados no preset (`requires`, ex.: `node`, `npx`).
3. **Faltantes:** exibe o comando exato de instalação definido no preset (`install.command` e `install.args`) e pede confirmação explícita (default "não").
   - `--yes`: executa as instalações definidas no preset sem perguntar (uso em CI).
   - Não interativo sem `--yes`: apenas reporta o que falta e como instalar; nunca executa.
   - `--dry-run`: nunca executa (`ARQ-05`).
4. **Reuso:** a mesma função de checagem é usada por `forge-sdd doctor` (relatório) e pelo boot do Studio (aviso claro de requisito ausente; nunca bloqueia o boot; modo leitura).
5. **Chip:** `N` do chip "x de N conectadas" passa a ser o número de estações com agente habilitado/configurado (ajuste em 68-03/68-04). Estação sem agente habilitado aparece como "não configurada", não como erro.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** um único pacote de checagem em `internal/` com dependências injetáveis (`LookPath`, executor, prompt); `cmd/` só adapta flags e UI (`ARQ-01`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Prompts interativos | `internal/survey/survey.go` | `Run`, `RunUpdate`, `RunConfirm` (huh, dependência já existente) | estender | Nova função de multisseleção de agentes e reuso de `RunConfirm`; sem nova dependência |
| Fluxo init/update e `--yes` | `cmd/forge-sdd/main.go` | `--yes`, `--dry-run`, seleção de `--agent` (agentes de scaffold) | estender | Encaixar o preflight após a seleção; não confundir `agents` do scaffold com agentes ACP |
| Doctor | `cmd/forge-sdd/doctor.go`, `internal/*` | Relatório de saúde do projeto | estender | Nova seção "Requisitos dos agentes" chamando a função única |
| Config e presets | `internal/config` (68-01) | Bloco `acp.*`, presets embutidos | estender | Acrescentar `enabledAgents`, `requires`, `versionArgs`, `install` nos presets |
| Boot do Studio | `internal/studio/server.go` | Warmup sem diagnóstico | estender | Chamar a checagem e expor aviso (68-04) |
| Checagem de requisitos | `internal/` | Inexistente | criar | Pacote `internal/preflight` com `Check`, `Plan`, `Install` (`DUP-03`: um único dono) |
| Execução de comandos externos | `internal/gate` (git/go) | Padrão de argv explícito | reusar padrão | `exec.Command` sem shell (`SEG-03`) |

## Dependências novas
- Nenhuma. Stdlib (`os/exec`, `context`) e `huh` já presente (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-03` (argv explícito, nunca `sh -c`), `SEG-01` (validar `install.command`/`args` do preset e do usuário na borda; rejeitar metacaracteres de shell e comando vazio), `SEG-04`/`SEG-05` (sem segredos nem env em logs ou telemetria), `SEG-02` (grava só em `sdd/.sddrc`), `ARQ-01`, `ARQ-05`, `ARQ-03` (presets embutidos), `DUP-01`/`DUP-03` (função única), `COD-01`, `COD-03`, `COD-05` (pt-BR, sem emojis), `TST-03`, `TST-04`.
- Regra de segurança adicional: a instalação só roda comando vindo do preset/config do usuário, com confirmação explícita ou `--yes`; nada hardcoded fora do preset. Instalar executa código de terceiros: o comando exato é sempre mostrado antes.
- Rules: `sdd-seguranca.md`, `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-02` (golden/help se mudar), `TST-05`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race -run "TestPreflight" ./internal/preflight/... ./internal/survey/... ./cmd/... && go test -race ./...
```
Com `LookPath`, executor e prompt injetáveis (fakes, sem rede, `TST-04`):
- Agente com CLI e requisitos presentes: relatório "ok", nada a instalar.
- CLI ausente: relatório lista o comando exato do preset; resposta padrão (vazia) = não instala.
- Confirmação "sim": executor fake recebe argv exato do preset; falha do executor vira erro enriquecido, sem panic.
- `--yes`: executa sem prompt; modo não interativo sem `--yes` e `--dry-run`: nunca chama o executor.
- `npx` ausente para claude/codex: reporta Node/npx como requisito faltante.
- Preset com `install.command` contendo metacaracteres de shell é rejeitado.
- Seleção grava `acp.enabledAgents`; `Status` do pool e o chip usam N = estações habilitadas.
- `doctor` e boot do Studio chamam a mesma função (teste de que não há segunda implementação) e o Studio sobe com requisito ausente.
