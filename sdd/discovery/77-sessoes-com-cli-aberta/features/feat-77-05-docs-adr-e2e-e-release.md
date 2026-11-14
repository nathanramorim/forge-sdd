# feat-77-05 — Docs, ADR, E2E e release

**Branch:** `feat/77-sessoes-com-cli-aberta`
**Fase:** 77-05
**Tamanho:** `M`
**Depende de:** 77-04
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
Finalizar a iniciativa: registrar ADR-0008, documentar o modo `live` nos guias e templates, atualizar conceitos e histórico de release, e validar de ponta a ponta com tmux real em servidor isolado.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Documentação, ADR, release notes e testes E2E (`ARQ-11`).

## Decisões do projeto aplicadas
- `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-05`.
- Orçamento de prompts e instruções (`P-24`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Documentação do pipeline | `README.md`, `docs/pipeline.md`, `.forge/LEIA-ME.md`, `site/public/guia-pipeline.html` | Descrevem modo headless | estender | Explicar modo `live`, entrega por despachante e limpeza de contexto. |
| ADRs | `sdd/spec/adr/` | ADR-0001 a ADR-0007 | criar | Criar `adr-0008-sessoes-com-cli-aberta-no-tmux.md`. |
| Conceitos | `sdd/spec/concepts.md` | Tabela de contratos | estender | Adicionar linha para despachante de entrega e modo live. |
| Teste com tmux real | `internal/pipeline/tmux_real_test.go` | Servidor isolado com `FORGE_TMUX_REAL` | estender | Adicionar teste E2E do despachante com agente script falso. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-05`.
- Rules: `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md`.

## Escopo
1. Criar `ADR-0008` e registrar novos contratos em `concepts.md`.
2. Atualizar release notes em `sdd/releases/history.md`.
3. Atualizar `README.md`, `docs/pipeline.md`, `LEIA-ME.md.tmpl` e `guia-pipeline.html`.
4. Implementar teste E2E `TmuxLiveReal` com servidor isolado de tmux.
5. Garantir gate Camada A 100% verde (`P-13`, `P-21`, `P-23`, `P-24`).

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./...
FORGE_TMUX_REAL=1 go test -v ./internal/pipeline -run TmuxLiveReal
go run ./cmd/forge-sdd gate
```
- E2E real com tmux isolado entrega prompt, detecta estado e reseta contexto.
- Portão de governança sem violações.
