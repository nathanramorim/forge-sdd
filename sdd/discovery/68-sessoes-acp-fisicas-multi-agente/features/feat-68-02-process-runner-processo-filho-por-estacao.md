# feat-68-02 — ProcessRunner: Processo Filho Real por Estação

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-02-process-runner-processo-filho-por-estacao
**Tamanho:** `M`
**Depende de:** feat-68-01
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Cria o componente que inicia o agente como processo filho (`os/exec`, argv explícito), registra PID, liga stdin/stdout ao protocolo de linhas, executa health check (`initialize` JSON-RPC com timeout) e encerra com segurança (sinal, `Wait`, sem zumbis). Captura stderr/stdout para o log da estação.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** componente stdlib em `internal/acp`, sem daemon externo nem Docker (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Moldura stdio JSON-RPC | `internal/acp/transport_stdio.go` | `StdioTransport` (Reader/Writer) | reusar | Mesmo enquadramento de linhas ao falar com o filho |
| Log da estação | `internal/acp/attach.go` | `AttachBridge.BroadcastLog` | reusar | Saída do filho vira `SessionLogEntry` |
| Tipos | `internal/acp/types.go` | `Request`, `Response` | reusar | Health check usa `initialize` existente |
| Execução de processos | `internal/acp`, `internal/gate` | Não há runner de agente | criar | `internal/acp/process.go` + `process_test.go` (`COD-01`) |
| Workspace | `internal/acp/session.go` | `ValidateWorkspace` | reusar | `cmd.Dir` confinado (`SEG-01`) |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-04`, `SEG-01`, `SEG-03` (argv explícito, sem `sh -c`), `SEG-05` (env/args fora de logs), `COD-01`, `COD-03`, `TST-03`, `TST-04` (processo falso via `os.Args[0]`, sem rede).
- Rules: `sdd-acp.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05` (pt-BR, sem emojis), `TST-05`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race -run "TestProcessRunner" ./internal/acp/...
```
- Inicia filho falso, expõe PID > 0 e responde ao health check.
- Filho que não responde dentro do timeout resulta em erro de health check.
- Comando inexistente retorna erro enriquecido sem panic.
- `Stop` encerra o filho e `Wait` coleta (sem processo remanescente).

### Decisões FLX-01 incorporadas
- Dois modos de health check: `initialize` (JSON-RPC, valida resposta) e `liveness` (processo não saiu e escrita em stdin não falha). Modo vem de `healthCheck.method`; timeout e intervalo configuráveis com default herdado do warmup atual.
- `ProcessRunner` expõe canal de saída (`Exited() <-chan ExitInfo` com exit code e causa) para o restart do pool (68-03).
- Testes extras: `TestProcessRunnerLivenessMode`, `TestProcessRunnerExitInfo` (exit code não zero e morte por sinal).
