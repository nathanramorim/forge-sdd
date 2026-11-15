# feat-59-04 — Integração de Handoff nas 5 Estações (`.agents/`), Catálogo de Conceitos (`concepts.md`) e Documentação (`FLOW.md` / `README.md` / `history.md`)

**Branch:** `feat/59-modulo-visual-estacoes-personagens`
**Fase:** 59-04
**Tamanho:** `P`
**Depende de:** `feat-59-01-filas-por-estacao-idempotencia-escrita-atomica`, `feat-59-03-interface-visual-embarcada-estacoes-e-governanca`
**Status:** `done`

## Objetivo

Finalizar a entrega do **Discovery 59**:
1. Atualizar `sdd/spec/concepts.md` (`DUP-01`) registrando a extensão de `internal/queue` (Filas por Estação com Idempotência) e o novo conceito `Módulo Visual Plugável (Forge Studio)` em `internal/studio`.
2. Atualizar as instruções de handoff em `.agents/roles/` e `.agents/commands/` (e seus templates em `internal/scaffold/templates/`) documentando as filas por estação e o subcomando `forge-sdd studio`.
3. Atualizar `sdd/FLOW.md`, `README.md` e `sdd/releases/history.md` com os novos fluxos e comandos.
4. Executar bateria de validação determinística (`go test ./...`, `go vet ./...`, `forge-sdd gate --branch feat/59-modulo-visual-estacoes-personagens --mode bloqueio` e `forge-sdd doctor`).

## Modelo do módulo
- **Modelo:** `enxuto` (documentação de conceitos, templates de agentes e fechamento de release).
- **Justificativa:** Consolida as novas capacidades em documentação viva e garante conformidade com a Constituição.

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Catálogo de Conceitos | `sdd/spec/concepts.md` | Tabela canônica de donos de conceitos | `estender` | Adicionar `internal/studio` e atualizar `internal/queue` (`DUP-01`) |
| Documentação do Fluxo | `sdd/FLOW.md` e `README.md` | Instruções operacionais | `estender` | Documentar o uso de `forge-sdd studio` e filas de estações |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-fluxo-e-portao.md`
- **IDs:** `ARQ-01..04`, `DUP-01`, `DUP-02`, `FLX-01..07`, `GOV-01..04`
- **Skills Reutilizadas:** `novo-subcomando-cli-sdd`

## Arquivos Tocados
- `sdd/spec/concepts.md`
- `sdd/FLOW.md`
- `README.md`
- `sdd/releases/history.md`
- `.agents/roles/*.md`
- `internal/scaffold/templates/`

## Critérios de Aceitação Executáveis
1. **CA-01 a CA-10 (Bateria Global e Portão Determinístico):**
   ```bash
   go test -race ./... && go vet ./... && go run ./cmd/forge-sdd gate --branch feat/59-modulo-visual-estacoes-personagens --mode bloqueio
   ```
