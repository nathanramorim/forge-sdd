# feat-72-04 — Abrir pelo tmux

**Branch:** `feat/72-visibilidade-tmux-e-entrega-ao-builder`
**Fase:** 72-04
**Tamanho:** `M`
**Depende de:** —
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Três janelas de terminal soltas dificultam ver o conjunto. Com `tmux` instalado, `forge-sdd up` passa a abrir uma tela com os três painéis nomeados.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** novo destino no launcher existente, com argv explícito e testes por geração de comandos (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Launcher | `internal/pipeline/launcher.go` (`SessionsFor`, `LaunchPlanFor`, `quoteArg`) | Plano de abertura por SO com o comando de cada sessão já quotado | estender | o tmux é mais um destino do mesmo plano; o comando de cada painel é o mesmo `Inner` |
| Comando `up` | `cmd/forge-sdd/up.go` | `--discovery`, `--builder`, `--reviewer`, `--dry-run` | estender | `--terminal tmux|native|auto` |
| Configuração | `internal/config/forge.go` | Dono de `forge.json` | estender | campo `terminal` (padrão `auto`), validado por lista fechada |
| Checagem e menu | `pipeline.Preflight`, `cmd/forge-sdd/menu.go` | Itens ok, aviso e bloqueio | estender | o menu mostra se vai abrir pelo tmux e deixa trocar |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-03`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. `internal/pipeline/tmux.go`: `TmuxPlan(env, sessions)` devolve os comandos (`new-session -d`, `split-window -h`, `split-window -v`, `select-pane -T`, `attach-session` ou `switch-client` quando `TMUX` está definido) como argv explícito; Discovery grande à esquerda, Builder e Reviewer empilhados à direita, títulos `forge-discovery`, `forge-builder`, `forge-reviewer`.
2. Nome da sessão `forge-<projeto>` saneado (`[A-Za-z0-9_-]`, até 40 caracteres); o comando de cada painel é o `Inner` já quotado; nada vindo de `forge.json` vira argumento sem validação (`SEG-03`).
3. Sessão existente (`tmux has-session`): só conecta, sem duplicar painéis.
4. `forge-sdd up --terminal tmux|native|auto` e `terminal` em `forge.json` (padrão `auto`: tmux quando está no PATH, senão o terminal do sistema como hoje); `--terminal tmux` sem tmux é erro claro com a instalação sugerida.
5. `--dry-run` imprime os comandos do tmux exatamente como serão executados; o menu mostra "abre no tmux" ou "abre em janelas" na confirmação.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./internal/config/... ./cmd/forge-sdd/...
```
- `TestTmux`: plano exato (painéis, títulos, attach ou switch-client), nome de sessão saneado, sessão existente só conecta, injeção em nome de projeto recusada, `--terminal native` e ausência de tmux mantêm o terminal do sistema, `--terminal tmux` sem tmux dá erro claro.
- `TestUp_Terminal`: precedência (flag > `forge.json` > auto) e valor inválido recusado.
- `forge-sdd up --dry-run` com `--terminal tmux` imprime os comandos esperados; testes da 71 passam sem edição.
