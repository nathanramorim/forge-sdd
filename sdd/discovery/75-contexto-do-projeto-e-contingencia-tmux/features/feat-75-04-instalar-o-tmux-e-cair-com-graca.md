# feat-75-04 — Instalar o tmux e cair com graça

**Branch:** `feat/75-contexto-do-projeto-e-contingencia-tmux`
**Fase:** 75-04
**Tamanho:** `G`
**Depende de:** —
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
Sem tmux, `auto` abre janelas soltas sem dizer por quê e `--terminal tmux` dá erro. A pessoa quer que o Forge ofereça instalar o tmux em qualquer sistema e, se não der, siga por uma contingência.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** escreve no sistema da pessoa (instalação de pacote), decide por SO e PATH, interage com o terminal e tem queda controlada; contratos e runner injetável (`ARQ-11`, `SEG-01`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Resolução de terminal | `internal/pipeline/tmux.go` (`ResolveTerminal`, `LaunchTmux`, `DefaultTmuxRunner`) | `auto` sem tmux vira nativo em silêncio; `tmux` explícito dá erro | estender | Aviso com motivo, oferta de instalação, queda do tmux em `auto`. |
| Runner de comandos | `tmux.go` (`TmuxRunner`), `runner.go` | Argv sem shell, injetável | reusar | O instalador usa o mesmo padrão e runner falso nos testes. |
| Confirmação | `internal/survey/select.go` (`ConfirmYesNo`) | Pergunta sim/não | reusar | A pergunta de instalar. |
| Terminal interativo | `cmd/forge-sdd/menu.go` (`isInteractive`) | Detecção de TTY | reusar | Sem terminal só imprime o comando. |
| `forge.json` | `internal/config/forge.go` (`Terminal`, `SaveForgeConfig`) | `auto`, `tmux`, `native` | reusar | Recusar e lembrar grava `terminal: native`; sem campo novo. |
| Checagem antes de abrir | `internal/pipeline/preflight.go` | CLIs e arquivos | estender | Item `tmux ausente` (aviso) com o comando do sistema. |

## Dependências novas
- Nenhuma. Gerenciadores de pacotes são programas externos opcionais (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-03`, `SEG-07`, `COD-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. `internal/pipeline/tmuxinstall.go`: `InstallPlan(env) (argv, desc, ok)` pela ordem `brew install tmux`, `apt-get install -y tmux`, `dnf install -y tmux`, `pacman -S --noconfirm tmux`, `apk add tmux`, `zypper --non-interactive install tmux`, com `sudo` quando não é root e o gerenciador pede; Windows nativo e sem gerenciador conhecido: `ok=false`. `Install` roda o argv herdando stdin, stdout e stderr (a senha do `sudo` é pedida no terminal da pessoa), sem shell.
2. Fluxo (`auto` sem tmux, terminal interativo): mostra o comando, pergunta; se instala e `tmux -V` responde, segue com tmux; se recusa, pergunta se lembra (`terminal: native`); sem instalador ou com falha, segue pela contingência com uma linha de motivo. Sem terminal interativo: imprime o comando e segue. `--terminal tmux` explícito: nunca instala; erro com o comando.
3. Queda do tmux: em `auto`, falha de `LaunchTmux` desfaz o que criou, avisa o motivo e abre pela contingência; `tmux` explícito mantém o erro.
4. `Preflight`: item `tmux ausente` como aviso, com o comando do sistema.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./cmd/forge-sdd/...
```
- `TestTmuxInstall`: tabela por SO e PATH (macOS, Debian, Fedora, Arch, Alpine, openSUSE; root e não root); Windows nativo sem opção; argv sem shell; runner falso confirma a ordem e que nada roda sem confirmação.
- `TestTerminalDecisao`: `auto` sem tmux, interativo e resposta sim instala e usa tmux; resposta não só grava `terminal: native` se a pessoa pedir; sem terminal só imprime; `--terminal tmux` não instala e erra com o comando.
- `TestTmuxFallback`: falha do `new-session` em `auto` desfaz, avisa e abre pela contingência; em `tmux` explícito retorna erro.
- `FORGE_TMUX_REAL=1 go test ./internal/pipeline -run TmuxReal` segue verde; `forge-sdd gate` sem falhas.
