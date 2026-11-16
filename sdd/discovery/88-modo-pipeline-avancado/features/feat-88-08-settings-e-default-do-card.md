# feat-88-08 — Menu Settings e default do card

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** nenhuma

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** `internal/config` e `cmd/forge-sdd/menu.go` (ARQ-01, ARQ-06).

## Decisões do projeto aplicadas
- ADR-0005 (menu = mesmo código do subcomando).
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Chave de card | `menu.go` (`configMenu`, caso `card`), `PMConfig.RequireExistingCard` | Confirm solto; default `true` | mover | Uma casa para chaves liga/desliga (DUP-01). |
| Init do pipeline | `runPipelineInit` | Grava defaults sem perguntar | estender | Pergunta se usa gerenciador de projeto; sem terminal grava `false`. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-06`, `DUP-01`, `COD-03`, `TST-01`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. `DefaultForgeConfig().PMMCP.RequireExistingCard = false`; `forge.json` existente mantém o valor.
2. `forge pipeline init` pergunta no terminal e grava a resposta.
3. Menu principal "Settings" com chaves `ativado`/`desativado` (por ora só o card), gravadas por `SaveForgeConfig`; remove o item de `configMenu`; atualiza `docs/pipeline.md` e golden.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/config ./cmd/forge-sdd -run 'TestDefaultCard|TestPipelineInit_PerguntaCard|TestMenu_Settings'
```
- `TestDefaultCard_Falso`, `TestPipelineInit_PerguntaCard`, `TestMenu_Settings_AlternaCard`.
- `go vet ./...` limpo.
