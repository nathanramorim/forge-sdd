# feat-68-06b — Seleção Interativa de Agentes (init/update) e Instalação com Confirmação

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-06b
**Tamanho:** `M`
**Depende de:** `feat-68-06`, `feat-68-01`
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Parte fatiada da 68-06. Em `init` e `update`, pergunta quais agentes o usuário vai usar (multisseleção entre copilot, claude, gemini, codex, agy e outros de `acp.adapters`), grava `acp.enabledAgents` e, para os requisitos faltantes apontados por `internal/preflight`, mostra o comando exato do preset e pede confirmação (default "não"). `--yes` executa sem perguntar (CI); não interativo sem `--yes` e `--dry-run` apenas reportam.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** prompt em `internal/survey` e instalador em `internal/preflight` com executor e prompt injetáveis; `cmd/` só adapta flags (`ARQ-01`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Prompts | `internal/survey/survey.go` | `Run`, `RunUpdate`, `RunConfirm` (huh) | estender | Multisseleção de agentes ACP; reusa `RunConfirm` |
| Detecção | `internal/preflight` (68-06) | `Check`, `Plan` | reusar | Sem segunda implementação (`DUP-03`) |
| Config | `internal/config` (68-01) | `acp.enabledAgents`, `install` do preset | reusar | Gravação via API de config |
| Fluxo init/update | `cmd/forge-sdd/main.go` | `--yes`, `--dry-run` | estender | Acionar preflight após seleção de agentes de scaffold |
| Execução externa | `internal/gate` | argv explícito | reusar padrão | `SEG-03` |

## Dependências novas
- Nenhuma (`huh` já presente) (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-03`, `SEG-05`, `SEG-02`, `ARQ-01`, `ARQ-05`, `DUP-03`, `COD-03`, `COD-05`, `TST-03`, `TST-04`. Instalação só via comando do preset/config, mostrado antes e confirmado. Rules: `sdd-seguranca.md`, `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`. Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-02`, `TST-05`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race -run "TestPreflightInstall|TestSurveyAgents|TestInitUpdateAgents" ./internal/preflight/... ./internal/survey/... ./cmd/... && go test -race ./...
```
- Resposta vazia (default) nunca chama o executor; "sim" chama com o argv exato do preset.
- `--yes` executa sem prompt; não interativo sem `--yes` e `--dry-run` nunca executam.
- Seleção grava `acp.enabledAgents`; executor com falha vira erro enriquecido.
- `install.command` com metacaracteres de shell é rejeitado.
