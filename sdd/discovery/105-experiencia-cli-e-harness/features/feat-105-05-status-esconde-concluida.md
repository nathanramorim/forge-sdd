# feat-105-05 — `status` esconde iniciativa concluída; `--all` mostra

**Iniciativa:** 105-experiencia-cli-e-harness (onda 1, task 1.5)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** nada
**Paralelo com:** feat-105-02, feat-105-03, feat-105-04
**Sequenciar com:** 106/107: se mexerem em `status.go` ou menus, série com rebase; sem editar arquivos delas. Destrava o `/archive` das iniciativas 79 a 88 (decisão do dono, fora deste PR).

## Problema
Iniciativas com fila concluída aparecem como ativas no `forge status` (O14).

## Comportamento
1. Iniciativa cuja fila só tem itens `completed`/`approved` (estado por iniciativa, ADR-0002) some da listagem padrão.
2. `forge status --all` mostra todas, marcando as concluídas.
3. Iniciativa sem fila, ou com fila vazia, continua visível (não é "concluída").
4. Somente leitura; saída curta sem TTY preservada (`status.go:60-77`).

## Modelo do módulo
**Estende** o cálculo do status em `internal/`; a regra "concluída" é função pura testável; `cmd/forge-sdd/status.go` só liga `--all` (`ARQ-01`, `ARQ-10`).

## Decisões do projeto aplicadas
ADR-0002 (estado por iniciativa), ADR-0005 (menu chama o `RunE` do `status`; não duplicar), ADR-0004.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Listagem do status | `cmd/forge-sdd/status.go`, `internal/pipeline` (status) | Lista iniciativas, saída curta sem TTY | **Estender** | `DUP-06` |
| Estado da fila | `queue.json` por iniciativa em `sdd/discovery/<id>/` e `internal/queue` | Status dos itens | **Reusar** o leitor | `DUP-01` |
| Teste de status | `cmd/forge-sdd/status_test.go` | Testes existentes | **Estender** | `TST-03` |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-10`, `DUP-06`, `TST-03`. Rules: `sdd-arquitetura`, `sdd-testes`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go build ./... && go vet ./...
go test -list 'TestStatusEscondeIniciativaConcluida' ./cmd/forge-sdd | grep -F 'TestStatusEscondeIniciativaConcluida'
go test ./cmd/forge-sdd -run 'TestStatusEscondeIniciativaConcluida'
go test -list 'TestStatusAllMostraConcluida' ./cmd/forge-sdd | grep -F 'TestStatusAllMostraConcluida'
go test ./cmd/forge-sdd -run 'TestStatusAllMostraConcluida'
# negativo: fila vazia ou com item pendente NÃO é escondida
go test -list 'TestStatusNaoEscondeFilaPendenteOuVazia' ./cmd/forge-sdd | grep -F 'TestStatusNaoEscondeFilaPendenteOuVazia'
go test ./cmd/forge-sdd -run 'TestStatusNaoEscondeFilaPendenteOuVazia'
go test ./...
```
- `status` não escreve no disco; saída sem TTY continua curta.
- Depois do merge, a pessoa decide o `/archive` de 79 a 88; este PR não arquiva nada.
