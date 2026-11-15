# fix-84-01 — Tmux set-option não aceita prefixo '=' em status-right com tokens acumulados

**Branch:** `feat/84-painel-interativo-chat-e-fila-drilldown`
**Origem:** Discovery 81 / Iniciativa 84 (`internal/pipeline/tmux.go`)
**Tamanho:** `S`
**Status:** `done`

## Feature Correlacionada
- **Arquivos corrigidos:** `internal/pipeline/tmux.go`, `internal/pipeline/tmux_test.go`, `internal/pipeline/tmux_real_test.go`.
- **Causa-raiz:**
  1. Em `BuildTmuxPlan` (`internal/pipeline/tmux.go`), quando `totalTokens > 0`, os comandos `tmux set-option` para `status-right` e `status-right-length` passavam `-t exact` (`=forge-<proj>-<hash>`) em vez de `-t name` (`forge-<proj>-<hash>`), tanto no modo `live` (linhas 175-176) quanto no modo `headless` (linhas 199-200).
  2. O comando `tmux set-option -t` não aceita o prefixo `=` no tmux 3.8 e encerra com `exit status 1 (no such session: =forge-...)`.
  3. Ao receber erro em qualquer passo de `plan.Create`, `LaunchTmux` mata a sessão recém-criada e `LaunchTerminalTmux` aciona o `nativeFallback()`, abrindo 3 janelas nativas separadas no macOS (`O tmux não conseguiu abrir as sessões (tmux set-option: exit status 1). Abrindo em janelas do sistema.`).

## Escopo Aplicado
1. Substituído `exact` por `name` nas 4 chamadas de `tmux set-option` para `status-right` e `status-right-length` em `internal/pipeline/tmux.go`.
2. Estendido `TestTmux_BordaBadgeDeTokensEStatusRight` em `internal/pipeline/tmux_test.go` para cobrir os modos `live` e `headless` e validar que nenhum `set-option` usa o prefixo `=`.
3. Adicionado `TestTmuxReal_StatusRightComTokensNaoFalhaSetOption` em `internal/pipeline/tmux_real_test.go` validando contra um servidor `tmux` real com `totalTokens > 0`.
4. Bump de versão para `4.7.1-beta` (`cmd/forge-sdd/main.go`, `npm/package.json` e `sdd/releases/history.md`).
