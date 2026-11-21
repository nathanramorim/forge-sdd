# feat-105-06 — `lider`: trocar descrição e texto, sem renomear nem remover

**Iniciativa:** 105-experiencia-cli-e-harness (onda 1, task 1.6)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** feat-105-01 (e feat-105-02 para não regravar o mesmo template duas vezes)
**Paralelo com:** nenhuma (toca os mesmos templates e goldens da 1.2: entra depois dela)
**Sequenciar com:** 106 (templates, goldens, `hooks.go`) e 107: série com rebase; sem editar arquivos delas.

## Problema
`lider` descreve "coordena as 5 estações", mas é o agente da sessão Discovery usado por `forge up` (`launcher.go`, `tmux.go`, `preflight.go`). Remover ou renomear quebra o `up` (O2).

## Comportamento
1. Trocar a descrição e o corpo de `.claude/agents/lider.md`, do template `agents/claude/.claude/agents/lider.md.tmpl` e do `chatmode` equivalente para: agente da sessão Discovery (3 sessões), gates humanos e fila; sem "5 estações".
2. O nome `lider`, o caminho do arquivo e as referências do Go ficam intactos.
3. Goldens regravados no mesmo PR (`TST-02`).
4. Testes de `launcher`, `tmux`, `preflight` e `up_sessions` seguem verdes sem edição de lógica.

## Modelo do módulo
**Sem módulo novo.** Texto de template e golden; Go não muda.

## Decisões do projeto aplicadas
ADR da feat-105-01 (decisão sobre `lider`), ADR-0004, ADR-0008 (modo `sessions`).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Arquivo do agente | `.claude/agents/lider.md`, `internal/scaffold/templates/agents/claude/.claude/agents/lider.md.tmpl` | Descrição com "5 estações" | **Corrigir texto** | O2 |
| Uso no Go | `internal/pipeline/launcher.go`, `tmux.go`, `preflight.go` | Usam `lider` como agente Discovery | **Manter** | Não quebrar `forge up` |
| Testes | `launcher_test`, `tmux_test`, `preflight_test`, `up_sessions_test` | Citam `lider` | **Reusar** como guarda | `TST-02` |
| Texto legado | `internal/scaffold/legacy/roles/lider.md.tmpl` | "5 estações" | **Tratar junto com feat-105-02** (um dono só) | `DUP-06` |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`COD-04`, `TST-02`, `FLX-02`. Rules: `sdd-codigo`, `sdd-testes`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/pipeline/... ./internal/scaffold/...
go test -list 'TestLiderDescricaoSemCincoEstacoes' ./internal/scaffold | grep -F 'TestLiderDescricaoSemCincoEstacoes'
go test ./internal/scaffold -run 'TestLiderDescricaoSemCincoEstacoes'
# negativo: o agente `lider` continua existindo no template e no repositório (nada renomeado ou removido)
go test -list 'TestLiderContinuaNoScaffold' ./internal/scaffold | grep -F 'TestLiderContinuaNoScaffold'
go test ./internal/scaffold -run 'TestLiderContinuaNoScaffold'
# negativo: `forge up` ainda resolve o agente `lider` da sessão Discovery
go test -list 'TestLauncherDiscoveryUsaLider' ./internal/pipeline | grep -F 'TestLauncherDiscoveryUsaLider'
go test ./internal/pipeline -run 'TestLauncherDiscoveryUsaLider'
go test ./...
```
- `TestLauncherDiscoveryUsaLider` pode ser um teste novo mínimo se nenhum existente já travar o nome `lider`.
- Se algum teste de `launcher`/`tmux`/`preflight`/`up_sessions` precisar de mudança além de golden, parar e consultar o dono (indica que o nome vazou).
