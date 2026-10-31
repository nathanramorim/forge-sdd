# feat-59-01 — Motor de Filas por Estação (`sdd/.queue/stations/*.json`), Idempotência Estrita (`IdempotencyKey`) e Escrita Atômica

**Branch:** `feat/59-modulo-visual-estacoes-personagens`
**Fase:** 59-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `done`

## Objetivo

Estender o domínio `internal/queue` para gerenciar as **5 filas canônicas dedicadas por estação**:
1. `sdd/.queue/stations/lider.json` (`01. O Líder`)
2. `sdd/.queue/stations/pm-architect.json` (`02. PM & Arquiteto`)
3. `sdd/.queue/stations/planner.json` (`03. O Planejador`)
4. `sdd/.queue/stations/builder.json` (`04. Fila de Builders`)
5. `sdd/.queue/stations/governance.json` (`05. Analista de Governança`)

Implementar o motor de **Idempotência Estrita** com cálculo determinístico de `idempotency_key` via SHA-256 (`station + ":" + item_id + ":" + operation + ":" + artifact_path + ":" + attempt`), verificação de estado e **escrita atômica** (`writeFileAtomic` via arquivo temporário + `os.Rename`) com incremento monotônico de `version`. Garantir sincronia bidirecional transparente com `sdd/.queue/queue.json` para manter 100% de compatibilidade com a Feature 58, expondo os subcomandos na CLI `cmd/forge-sdd/queue.go` (`forge-sdd queue handoff`, `forge-sdd queue stations`).

## Modelo do módulo
- **Modelo:** `completo` dentro de `internal/queue` (gerencia as 5 filas de estação, idempotência e escrita atômica mantendo fonte única do conceito de fila `DUP-01`).
- **Justificativa:** Elimina concorrência e contenção de escrita entre agentes paralelos, garantindo que retries de LLM ou duplo disparo não dupliquem tarefas.

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Fila do Builder `queue.json` | `internal/queue/queue.go` | Estrutura `Queue`, `TaskItem`, `Enqueue`, `Next`, `Complete`, `Review` | `estender` | Adicionar suporte a `StationQueue`, `StationItem`, `idempotency_key`, atomic write e sincronia com `queue.json` (`DUP-01`) |
| CLI do subcomando queue | `cmd/forge-sdd/queue.go` | Comandos `enqueue`, `next`, `complete`, `review`, `list`, `panel` | `estender` | Adicionar `handoff` e `stations` preservando os comandos existentes |

## Dependências novas
Nenhuma (apenas biblioteca padrão Go `crypto/sha256`, `os`, `path/filepath`, `sync`, `time` — `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`
- **IDs:** `ARQ-01..04`, `DUP-01`, `DUP-02`, `DUP-05`, `SEG-04`, `TST-01..05`, `FLX-01..06`
- **Skills Reutilizadas:** `novo-subcomando-cli-sdd`

## Arquivos Tocados
- `internal/queue/queue.go`
- `internal/queue/stations.go`
- `internal/queue/atomic.go`
- `internal/queue/queue_test.go`
- `cmd/forge-sdd/queue.go`

## Critérios de Aceitação Executáveis
1. **CA-02 (Persistência das 5 Filas Canônicas e Escrita Atômica):**
   ```bash
   go test -v -run TestStationQueuesAtomic ./internal/queue/...
   ```
2. **CA-03 (Idempotência Estrita por Chave Determinística):**
   ```bash
   go test -v -run TestIdempotencyKey ./internal/queue/...
   ```
3. **CA-04 (Compatibilidade Retroativa e Subcomandos CLI):**
   ```bash
   go test -v -run TestQueueCLIStations ./cmd/forge-sdd/... && go vet ./...
   ```
