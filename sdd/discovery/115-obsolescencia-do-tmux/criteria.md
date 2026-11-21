# Critérios — Iniciativa 115 (Obsolescência do tmux)

## Restrições (Constitution)
`ARQ-01` (cmd só adapta; lógica em `internal/`), `ARQ-04` (sem dependência nova: a remoção **reduz**), `COD-03` (erro com contexto na configuração antiga), `COD-04` (código morto sai na mesma entrega), `DUP-01` (uma casa por conceito: um caminho de orquestração), `FLX-05` + `GOV-01` (ADR no mesmo PR da mudança de arquitetura), `SEG-01` (valor antigo de `terminal` recusado na borda), `TST-01`, `TST-05`.

## Modelo do módulo
**Enxuto** (`ARQ-11`): não há pacote novo, só **remoção** e adaptação. O que sobra de `internal/pipeline` continua com a dependência para dentro e sem importar `cmd/`.

## Decisões do projeto aplicadas
- ADR-0015 (pipeline em subagentes no mesmo chat) — passa de `proposto` a **aceito**, é a base desta iniciativa.
- ADR-0008 (sessões com CLI aberta no tmux) — **substituída**.
- ADR-0021 (obsolescência das 5 estações) — precedente do formato de remoção.
- `GOV-01`: a Constituição só muda se o texto de `ARQ-06` citar comandos que somem; a conferir.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Abertura de painéis | `internal/pipeline/tmux.go` | 333 linhas: plano de janelas, `LaunchTmux`, `BuildTmuxPlan` | **remover** | O chat não abre painel |
| Instalador por gerenciador | `internal/pipeline/tmuxinstall.go` | 292 linhas: tabela brew/apt/dnf/pacman/apk/zypper, `InstallPlan`, `DecideTerminal` | **remover** | Sem tmux, não há o que instalar |
| Despachante | `internal/pipeline/deliver.go` | `capture-pane`, `load-buffer`, `paste-buffer` | **adaptar**: no modo chat o prompt vai ao subagente; confirmar o que resta | Evitar remoção cega |
| Painel ao vivo | `internal/pipeline/status_live.go` | tmux como uma das fontes | **adaptar**: manter as fontes que não são tmux | O painel continua útil |
| Aviso de tmux ausente | `internal/pipeline/preflight.go` | sugere instalar ou `--terminal native` | **substituir** pelo aviso do modo chat | O usuário não precisa saber de terminal |
| Rótulo do terminal no menu | `cmd/forge-sdd/menu.go` (`terminalLabel`) | "no tmux (3 painéis)" / "janelas do sistema" | **remover** | Não há escolha de terminal |
| Configuração `terminal` | `internal/config/forge_profiles.go` | `auto`, `tmux`, `native` | **reduzir**: recusar `tmux` com mensagem de transição | `SEG-01`, `RN-2` |
| Testes de tmux | `tmux_test.go`, `tmux_real_test.go`, `tmuxinstall_test.go` | 960 linhas | **remover** junto com o código | `COD-04` |
| Documentação | `README.md`, `docs/`, `docs/comandos.md`, `/tutorial` | ensinam instalar tmux | **atualizar** | Lista fechada no critério |

## Dependências novas
`Nenhuma` (`SEG-07`). A iniciativa **reduz** dependências.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-04`, `ARQ-11`, `COD-03`, `COD-04`, `DUP-01`, `FLX-05`, `GOV-01`, `SEG-01`, `TST-01`, `TST-05`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. Nenhum comando do produto executa tmux (fora de mensagem de transição).
! grep -rn 'exec.Command("tmux"' internal/ cmd/ || exit 1
# 2. O código de tmux saiu.
test ! -f internal/pipeline/tmux.go
test ! -f internal/pipeline/tmuxinstall.go
# 3. Configuração antiga recusada com mensagem clara.
grep -q 'terminal' internal/config/forge_profiles.go
go test ./internal/config/... -run 'TestTerminalTmuxRecusado'
# 4. Doctor não fala de tmux; fala do modo chat.
! grep -rq 'tmux' internal/doctor/ || exit 1
go test ./internal/doctor/...
# 5. Nenhuma dependência externa nova.
test ! -s go.mod.tmp
go mod tidy && git diff --exit-code go.mod
# 6. Suíte inteira, sem depender de tmux instalado.
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** `TestTerminalTmuxRecusado` prova que `terminal: tmux` no `forge.json` produz erro com a mensagem de transição, e **não** cai em silêncio para outro modo.
- **Prova de ausência:** o `grep` do item 1 é a prova de que nenhum caminho do binário chama tmux; ele falha se alguém reintroduzir.
- **Prova documental:** `grep -ri 'tmux' README.md docs/` retorna vazio, exceto na seção "o que saiu", que existe para quem procura.

## Riscos técnicos a verificar na implementação

| Risco | Verificação |
|---|---|
| `deliver.go` usado também no modo headless (`worker`) | ler `worker.go` e os testes de `deliver` antes de remover |
| `status_live.go` quebra sem a fonte tmux | rodar o teste de `status_live` com a fonte removida |
| Configuração de projeto existente com `terminal: tmux` | teste de migração com `forge.json` antigo |
| Documentação espalhada | `grep -ri tmux` com lista fechada no PR |

## A confirmar no Gate 1

- A transição é de **uma versão** (aviso, sem remover o campo) ou corta direto? A recomendação é uma versão de aviso, porque há projetos com `terminal: tmux` gravado.
- O `ARQ-06` da Constituição cita algum comando que some? Se sim, precisa de PR de Constituição junto (`GOV-01`).