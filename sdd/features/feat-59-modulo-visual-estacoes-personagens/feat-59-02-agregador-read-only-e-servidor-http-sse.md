# feat-59-02 — Agregador Read-Only de Estado (`internal/studio`) e Servidor HTTP/SSE (`forge-sdd studio`)

**Branch:** `feat/59-modulo-visual-estacoes-personagens`
**Fase:** 59-02
**Tamanho:** `M`
**Depende de:** `feat-59-01-filas-por-estacao-idempotencia-escrita-atomica`
**Status:** `done`

## Objetivo

Criar o backend desacoplado no pacote `internal/studio` que atua como um observador passivo (estritamente Read-Only, `SEG-02`/`SEG-06`) sobre o estado do projeto:
1. Agrega as 5 filas por estação (`sdd/.queue/stations/*.json`), `sdd/.queue/queue.json`, `sdd/memory/progress.md`, `sdd/discovery/*` e `.agents/skills/*` no DTO determinístico `StudioSnapshot`.
2. Expõe a API HTTP local:
   - `GET /api/state`: snapshot JSON completo das estações, contratos de handoff, governança e métricas.
   - `GET /api/events`: stream Server-Sent Events (SSE) que monitora o vetor de versões e notifica o frontend apenas quando há mutação real.
3. Bloqueia métodos de mutação (`POST`, `PUT`, `DELETE` retornam `405 Method Not Allowed`).
4. Cria o comando CLI `cmd/forge-sdd/studio.go` com flags `--port`, `--host` (default `127.0.0.1`), `--dir`, `--open`.

## Modelo do módulo
- **Modelo:** `enxuto` dentro de `internal/studio` e `cmd/forge-sdd/studio.go`.
- **Justificativa:** Mantém separação estrita de domínios. O studio é um leitor passivo que não interfere na execução nem bloqueia comandos de terminal.

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Filas por estação | `internal/queue` | Motor de 5 filas `sdd/.queue/stations/*.json` criado em `feat-59-01` | `reusar` | Consumir os leitores públicos de fila sem alterar sua lógica (`DUP-01`) |
| Subcomandos CLI | `cmd/forge-sdd/*.go` | Padrão Cobra de subcomandos com flags | `criar` | Criar `cmd/forge-sdd/studio.go` seguindo a skill `novo-subcomando-cli-sdd` (`DUP-05`) |

## Dependências novas
Nenhuma (apenas `net/http`, `context`, `sync`, `time`, `encoding/json` da stdlib — `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-seguranca.md`
- **IDs:** `ARQ-01..04`, `SEG-01..06`, `TST-01..05`, `FLX-01..06`
- **Skills Reutilizadas:** `novo-subcomando-cli-sdd`

## Arquivos Tocados
- `internal/studio/server.go`
- `internal/studio/state.go`
- `internal/studio/sse.go`
- `internal/studio/studio_test.go`
- `cmd/forge-sdd/studio.go`

## Critérios de Aceitação Executáveis
1. **CA-01 & CA-07 (Subcomando CLI e Endpoints HTTP/SSE):**
   ```bash
   go test -v -run TestStudioServer ./internal/studio/...
   ```
2. **CA-09 (Segurança Read-Only e Rejeição de Mutações):**
   ```bash
   go test -v -run TestStudioReadOnlySecurity ./internal/studio/... && go vet ./...
   ```
