# feat-69-01 — Remoção do Forge Studio

**Branch:** `feat/69-otimizacao-estrutura-tokens-e-paralelismo`
**Fase:** 69-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
O Studio (`internal/studio`, 5.262 linhas, `three.min.js`) não será usado e custa leitura, manutenção e contexto. Esta task apaga o pacote, o subcomando e toda citação, e preserva o ACP (`internal/acp`) como núcleo sem interface. O único chamador de `StartHeartbeat` era o `server.go` do Studio, então o heartbeat passa a ser iniciado pelo servidor `forge-sdd acp`.

## Modelo do módulo
- **Modelo:** `enxuto` (remoção, sem módulo novo).
- **Justificativa:** o único importador de `internal/studio` é `cmd/forge-sdd/studio.go`; remover os dois não deixa símbolo órfão (`ARQ-02`, `ARQ-11`, `COD-04`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Pacote e comando do Studio | `internal/studio/**`, `cmd/forge-sdd/studio.go` | 9 arquivos Go e assets; único importador é `studio.go` | remover | Decisão do usuário (discovery 69); `COD-04` |
| Chamadas ao pool | `internal/studio/server.go` | `pool.Warmup`, `StartHeartbeat`, `SanitizeOrphanActivities`, `ReconcileStationActivities` | estender | `forge-sdd acp` passa a iniciar heartbeat e a sanear atividades órfãs |
| Warmup fora do Studio | `cmd/forge-sdd/attach.go:52`, `internal/acp/dispatcher.go:169` | Já chamam `Warmup` | reusar | Permanecem como estão |
| Citações ao Studio | `internal/scaffold/templates/**`, goldens, `README.md`, `sdd/FLOW.md` §8, `docs/manual-forge-studio.md`, `sdd/spec/**`, `.agents/**` | Mencionam o comando e a interface | remover | Editar templates e regenerar goldens (`TST-02`) |
| Catálogo de conceitos | `sdd/spec/concepts.md` | Linha "Módulo Visual Plugável" e `acp.HTTPTransport` inexistente em código não-teste | estender | Apagar a linha do Studio e corrigir a do ACP (`DUP-01`) |
| Rule ACP | `.agents/rules/sdd-acp.md` | Seções sobre Studio e transporte HTTP/SSE | estender | Remover só o que cita o Studio; manter o restante |
| Feat 68 | `sdd/discovery/68-*/` | 68-04 e 68-05 já canceladas | reusar | Sem novo ajuste na 68 |

## Dependências novas
- Nenhuma. Remove `three.min.js` embutido; `go.mod` não ganha entradas (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-02`, `ARQ-06`, `COD-04`, `COD-05`, `DUP-01`, `TST-02`, `TST-05`, `FLX-05`.
- Rules: `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`, `sdd-acp.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `GOV-01` (ADR no mesmo PR).

## Escopo
1. Apagar `internal/studio/` e `cmd/forge-sdd/studio.go`; remover o registro do comando em `main.go`.
2. Em `cmd/forge-sdd/acp.go`, iniciar `pool.StartHeartbeat` e a sanitização de atividades órfãs no ciclo de vida do servidor, encerrando com o contexto.
3. Apagar `docs/manual-forge-studio.md`; limpar `README.md`, `sdd/FLOW.md` (§8 e menções), templates, goldens, `sdd/spec/**`, `.agents/**` e adaptadores `.claude/`, `.gemini/`, `.github/`.
4. Criar `sdd/spec/adr/adr-0001-remocao-do-studio.md` a partir de `adr-0000-modelo.md`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./...
test ! -d internal/studio && test ! -f cmd/forge-sdd/studio.go
! forge-sdd studio >/dev/null 2>&1
! grep -rliE 'forge-sdd studio|forge studio' internal/scaffold sdd/FLOW.md README.md docs .agents .claude .gemini .github sdd/spec
```
- Teste novo em `cmd/forge-sdd`: iniciar o servidor `acp` com contexto cancelável inicia o heartbeat e o encerra sem goroutine remanescente.
- `sdd/spec/concepts.md` não cita `internal/studio` nem `acp.HTTPTransport`.
- ADR 0001 existe e referencia `GOV-01`.
