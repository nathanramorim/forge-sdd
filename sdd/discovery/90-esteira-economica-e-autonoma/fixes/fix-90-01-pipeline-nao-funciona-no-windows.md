# fix-90-01 — Pipeline não funciona no Windows: worker/CLI trava

**Branch:** `fix/90-01-pipeline-nao-funciona-no-windows`
**Origem:** Iniciativa 90, `internal/pipeline/runner.go`, `cmd/forge-sdd/worker.go`, `internal/survey/select.go`, `internal/pipeline/launcher.go`
**Tamanho:** `M`
**Status:** `done`

## Sintoma
No Windows, após `forge pipeline up` abrir as janelas corretamente (wt.exe ou cmd.exe), o `forge worker` (Builder/Reviewer) trava ou o agente Claude não inicia. As janelas aparecem, mas não avançam. Em alguns cenários o seletor interativo do `huh` não renderiza no cmd.exe nativo (exige digitação em vez de setas).

## Feature Correlacionada
- **Arquivo de origem:** `90-05` (`cmd/forge-sdd/worker.go`) e `feat-75-05` (`internal/pipeline/launcher.go`) — onde o suporte Windows foi introduzido.
- **Causa-raiz (hipóteses):**
  1. `cmd/forge-sdd/worker.go:45` usa `syscall.SIGTERM` em `signal.NotifyContext`. No Windows, o pacote `os/signal` não suporta `SIGTERM` via `signal.Notify`; o contexto pode nunca capturar o sinal de parada, travando o shutdown do worker.
  2. `internal/pipeline/runner.go:142` cria `exec.CommandContext` sem definir `cmd.Stdin`. Em Go, `Stdin == nil` redireciona para `os.DevNull`; mas no Windows, o Claude CLI pode tentar abrir `CONIN$`/`CONOUT$` diretamente para detectar TTY, causando conflito com a janela cmd.exe já aberta e travando o subprocesso.
  3. `internal/survey/select.go:40` chama `huh.NewForm(...).Run()` sem verificar suporte VT no terminal. No `cmd.exe` clássico (sem `ENABLE_VIRTUAL_TERMINAL_PROCESSING`), o bubbletea/huh falha ao entrar em modo raw, podendo produzir saída garbled ou travar aguardando input de teclado.
- **Evidência:**
  - `cmd/forge-sdd/worker.go:45`: `signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)` — `SIGTERM` ineficaz no Windows.
  - `internal/pipeline/runner.go:142–145`: `cmd.Stdin` não atribuído; `cmd.Env` herda `os.Environ()` sem supressão de console.
  - `internal/survey/select.go:40`: `huh.NewForm(...).Run()` sem `WithAccessible`.

## Modelo do módulo
**Enxuto** (ARQ-11): ajustes em `cmd/forge-sdd/worker.go`, `internal/pipeline/runner.go` e `internal/survey/select.go`. Sem módulo novo; sem dependências externas adicionais.

## Decisões do projeto aplicadas
- D1 (Go), D3 (huh v0.3), ADR-0004 (Pipeline Multi-Sessão), ADR-0008 (Sessões com CLI aberta no Tmux).
- Rules: `sdd-codigo-e-testes` (COD-01, COD-03), `sdd-arquitetura` (ARQ-01, ARQ-04), `sdd-fluxo-e-portao` (FLX-01).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Suporte Windows no worker (SIGTERM) | `cmd/forge-sdd/worker.go:45` | `syscall.SIGTERM` em `signal.NotifyContext` | **remover** `syscall.SIGTERM`; manter só `os.Interrupt` | Go/Windows não entrega `SIGTERM` via `signal.Notify` |
| Stdin do subprocesso no runner | `internal/pipeline/runner.go:142` | `cmd.Stdin` não atribuído | **definir** `cmd.Stdin = io.NopCloser(strings.NewReader(""))` para runs não interativas | Evita que o Claude CLI herde o console e tente abrir `CONIN$` |
| Modo acessível no huh | `internal/survey/select.go:40` e `gesture.go` | Nenhuma detecção de TTY | **adicionar** detecção de VT/TTY; passar `WithAccessible(true)` quando não suportado | huh v0.3 suporta modo acessível (texto plano); preserva UX no Windows Terminal |
| Detecção VT no Windows | `internal/survey/` | Nenhuma | **usar** `term.IsTerminal(int(os.Stdout.Fd()))` (já disponível via `golang.org/x/term` ou verificação manual no Windows) | Decisão: `golang.org/x/term` já indiretamente usado pelo bubbletea; verificar se já é transitivo antes de adicionar `go get` |

## Dependências novas
- `golang.org/x/term` para `term.IsTerminal` — verificar se já é dependência transitiva de `github.com/charmbracelet/bubbletea` antes de declarar (Regra 5: consultar context7 na versão exata); se já transitivo, não adicionar diretamente.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-04`, `COD-01`, `COD-03`, `SEG-01`, `TST-01`, `TST-05`.

## Premissas (a confirmar no Gate 1)
- O `forge worker` está travando no run do agente, não no poll da fila.
- O Windows Terminal (wt.exe) suporta VT; apenas o cmd.exe clássico não suporta → o modo acessível só ativa quando `term.IsTerminal` retorna false.
- Remover `syscall.SIGTERM` não afeta o comportamento em Linux/macOS (que já usam `os.Interrupt` como sinal primário de parada interativa).

## Escopo
1. `cmd/forge-sdd/worker.go:45`: remover `syscall.SIGTERM`; manter `os.Interrupt` apenas.
2. `internal/pipeline/runner.go`: definir `cmd.Stdin = io.NopCloser(strings.NewReader(""))` antes de `cmd.Run()` em execuções não interativas.
3. `internal/survey/select.go` e `gesture.go`: extrair helper `isTTY() bool`; passar `.WithAccessible(!isTTY())` em todos os `huh.NewForm(...).Run()` e `tea.NewProgram(...)`.
4. Testes unitários para `isTTY()` e para o runner com `Stdin` definido.
5. Teste manual documentado no Windows Terminal + cmd.exe nativo.

## Fora do escopo
- Suporte a wt.exe com múltiplos painéis (já cobre launcher.go).
- Suporte a WSL (já testado em 90).
- Mudar a lógica de polling do worker.

## Critérios de Aceitação Executáveis
```bash
go vet ./...
go test -v -race ./internal/survey/... -run 'TestSelectOne|TestIsTTY'
go test -v -race ./internal/pipeline/... -run 'TestExecRunner|TestWorker'
go test ./...
# Manual no Windows: forge pipeline up → janelas abrem → forge worker aguarda jobs
# → job enfileirado → agente Claude inicia e conclui o job sem travar
```
- Worker para com Ctrl+C (apenas `os.Interrupt`) no cmd.exe do Windows.
- Runner define `Stdin` vazio para o subprocesso Claude.
- Menu `forge pipeline up` renderiza o seletor de setas no Windows Terminal e cai para modo texto no cmd.exe nativo.
