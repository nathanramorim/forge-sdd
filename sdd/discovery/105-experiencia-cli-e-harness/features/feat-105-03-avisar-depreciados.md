# feat-105-03 — Avisar: depreciação de `queue stations`, `attach`, `queue handoff` e aliases

**Iniciativa:** 105-experiencia-cli-e-harness (onda 1, task 1.3)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** feat-105-01
**Paralelo com:** feat-105-02, feat-105-04, feat-105-05
**Sequenciar com:** 106/107: tocam `queue.go` só se a 106 mexer em `queue`; se sim, rebase e série; não editar arquivos da 106/107. Os arquivos `.claude/commands/implement.md` e `proxima-feature.md` são só desta task.

## Problema
Itens obsoletos (O3, O4, O5, O7, O13) seguem expostos sem aviso. A decisão do Gate 1: só avisar, nada removido.

## Comportamento
1. `queue stations`, `attach` e `queue handoff` ganham aviso próprio: `PreRun` com `avisoDepreciado` (stderr, pt-BR, diz o que usar e que sai em versão futura por ADR) e `Short` com prefixo `[depreciado]`. NÃO usar o `Deprecated` do cobra: ele esconde o comando do `--help`, e o item 4 manda manter o comando listado (desvio do Builder aceito pelo dono). Continuam funcionando com a mesma saída e código.
2. `/implement` e `/proxima-feature` ganham uma linha de aviso: "alias de `/implementar`; será removido depois".
3. `forge doctor` e `forge status` NÃO mudam aqui. O aviso O13 (5 pareceres antigos da 97) vira nota no `status.md` da 105 apontando `/archive` da 97; nada regravado.
4. O aviso vai para stderr; a saída de dados não muda; `--help` mantém o comando listado.

## Modelo do módulo
**Sem módulo novo.** `cmd/forge-sdd` só adapta flags e metadados (`ARQ-01`); lógica de `internal/queue` intacta.

## Decisões do projeto aplicadas
ADR da feat-105-01 (depreciação, `ARQ-06`), ADR-0010 (comando único `forge`), ADR-0004.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Subcomandos alvo | `cmd/forge-sdd/queue.go` (`handoff` ~305, `stations` ~337), `cmd/forge-sdd/attach.go:18` | Sem `Deprecated` | **Estender** | `COD-03` |
| Lógica das estações | `internal/queue/stations.go`, `activity.go` | Usada por `attach` e atividade | **Manter** | Nada removido |
| Aliases | `.claude/commands/implement.md`, `proxima-feature.md`, `implementar.md` | Apontam ao mesmo arquivo | **Estender** (aviso) | O7 |
| Teste de comando | `cmd/forge-sdd/*_test.go` | Execução de comandos por cobra em teste | **Reusar** o padrão | `TST-03` |
| Cópias do comando em templates | `internal/scaffold/templates/**/commands/` | Possível espelho dos aliases | **Verificar** no Builder; atualizar goldens se existir | `TST-02` |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-06`, `COD-04`, `TST-03`. Rules: `sdd-arquitetura`, `sdd-codigo`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go build ./... && go vet ./...
go test -list 'TestComandosDepreciadosAvisam' ./cmd/forge-sdd | grep -F 'TestComandosDepreciadosAvisam'
go test ./cmd/forge-sdd -run 'TestComandosDepreciadosAvisam'
# negativo: depreciado continua funcionando (código 0 e mesma saída de dados); nada foi removido
go test -list 'TestDepreciadosContinuamFuncionando' ./cmd/forge-sdd | grep -F 'TestDepreciadosContinuamFuncionando'
go test ./cmd/forge-sdd -run 'TestDepreciadosContinuamFuncionando'
# negativo: comando NÃO depreciado (ex.: `queue enqueue`, `status`) não ganha aviso
go test -list 'TestComandosVigentesSemAviso' ./cmd/forge-sdd | grep -F 'TestComandosVigentesSemAviso'
go test ./cmd/forge-sdd -run 'TestComandosVigentesSemAviso'
go test ./...
```
- Aviso em pt-BR (`COD-05`), no stderr, sem ANSI; `--dry-run` inalterado (`ARQ-05`).
- Código morto: nenhum (nada removido); `COD-04` cobra apenas não deixar a mensagem de depreciação sem teste.
