# feat-77-02 — Painéis ao vivo no tmux

**Branch:** `feat/77-sessoes-com-cli-aberta`
**Fase:** 77-02
**Tamanho:** `M`
**Depende de:** 77-01
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
No modo `live`, a janela principal do tmux roda as CLIs interativas dos agentes nos painéis 1 (Builder) e 2 (Reviewer), enquanto os despachantes (`worker --deliver`) rodam em segundo plano numa janela oculta `forge-ctl`.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Montagem de planos de tmux multi-janela, divisão de painéis, contingência suave para modo `headless` quando sem tmux (`ARQ-11`, `COD-01`, `SEG-01`).

## Decisões do projeto aplicadas
- `D-1 (ADR-0007)`: Dois processos por sessão no modo `live` (CLI interativa no painel visual, despachante na janela oculta `forge-ctl`).
- `D-7 (ADR-0007)`: Contingência: `session_mode: live` sem tmux (ou com `terminal: native`) reverte para `headless` com aviso claro.
- `ARQ-01`, `SEG-01`: Montagem segura de argv sem expansão de shell.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Plano tmux e execução | `internal/pipeline/tmux.go` | `BuildTmuxPlan`, `LaunchTmux` | estender | Suportar modo `live` criando janela oculta `forge-ctl` e comandos interativos para Builder e Reviewer. |
| Fallback e launcher | `internal/pipeline/launcher.go`, `up.go` | `LaunchPlanFor`, `upCmd` | estender | Exibir modo no `--dry-run` e acionar fallback para `headless` se terminal for nativo. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-03`, `COD-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.

## Escopo
1. Estender `BuildTmuxPlan` para aceitar `session_mode`. Em `live`:
   - Janela principal `forge`: painel 0 Discovery, painel 1 Builder (CLI interativa), painel 2 Reviewer (CLI interativa).
   - Janela oculta `forge-ctl`: `forge-sdd worker --role builder --deliver =<sessao>:forge.1` e `forge-sdd worker --role reviewer --deliver =<sessao>:forge.2`.
2. Em `headless`: manter comportamento idêntico ao atual.
3. Tratar contingência suave: se `terminal: native` ou tmux indisponível, executar fallback para `headless` e emitir aviso amigável.
4. Suporte a `forge-sdd up --dry-run` refletindo o modo `live` ou `headless`.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/pipeline -run TmuxLivePlan
go test -v ./internal/pipeline ./cmd/forge-sdd -run LiveFallback
```
- Plano `live` cria janela `forge-ctl` e painéis interativos.
- Plano `headless` permanece inalterado.
- Fallback para `headless` funciona corretamente em ambiente nativo.
