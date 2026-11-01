# feat-68-04 — Studio: Chip "x de 5 conectadas", Estação em Erro e Modo Leitura

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-04-studio-chip-conectadas-e-modo-leitura
**Tamanho:** `M`
**Depende de:** feat-68-03, feat-68-06
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Substitui o chip fixo "5 SESSÕES PRONTAS" por contagem real vinda de `pool/status`: "x de 5 conectadas" (pt-BR). Estações sem conexão aparecem em erro com motivo. O boot do Studio nunca bloqueia: warmup com timeout, falha parcial ou total mantém o Studio em modo leitura (telemetria, fila e specs navegáveis; ações que exigem agente ficam desabilitadas com aviso).

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** ajustes em `internal/studio` e assets embutidos; sem framework novo (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Boot e warmup | `internal/studio/server.go` (l.70-105, 300-320) | Warmup ignora erro | estender | Resultado parcial e modo leitura |
| Chip | `internal/studio/assets/index.html:29`, `studio.js` | `sessionPoolCount` fixo | estender | Valor dinâmico; texto "x de 5 conectadas" |
| Estado via SSE | `server.go` `AddStateListener` | Propaga estado das estações | reusar | Reflete mudança para erro em tempo real |
| Testes do Studio | `internal/studio/*_test.go` | `server_test.go`, `studio_test.go` | estender | Casos de falha e modo leitura |
| E2E do pool | `internal/acp/e2e_pool_test.go` | Pool e2e | estender | Cenário com processos falsos |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-10` (sem lógica de domínio no JS), `ARQ-03` (assets via `embed.FS`), `FLX-04`, `TST-02`, `TST-03`, `TST-04`, `COD-05` (UI pt-BR, sem emojis).
- Rules: `sdd-acp.md`, `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05` (pt-BR, sem emojis), `TST-05`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/studio/... ./internal/acp/... && go test ./...
```
- API de estado do Studio devolve `connectedCount` e `total`; HTML/JS não contém mais o texto "SESSÕES PRONTAS" fixo.
- Boot com todos os adaptadores falhando: servidor responde HTTP 200 e estações em `error`.
- Boot em menos que o timeout de warmup mesmo com filho travado.

### Decisões FLX-01 incorporadas
- Estação em `restarting` aparece como "reiniciando" e não conta como conectada no chip.
- Timeout de warmup vem de `acp.warmupTimeoutMs` (default herdado do atual); estações em modo `liveness` contam como conectadas.

### Ajuste por 68-06
- Chip "x de N conectadas" com N = estações habilitadas. O boot chama o preflight (`internal/preflight`) e exibe aviso pt-BR de requisito ausente, sem bloquear o boot (modo leitura). Teste: `TestStudioPreflightWarningNonBlocking`.
