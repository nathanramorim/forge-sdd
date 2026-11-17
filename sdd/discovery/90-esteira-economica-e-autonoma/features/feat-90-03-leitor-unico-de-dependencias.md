# feat-90-03 — Um leitor de dependências e checagem das specs antes do build

**Iniciativa:** 90-esteira-economica-e-autonoma
**Branch:** feat/90-esteira-economica-e-autonoma (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** nenhuma

## Problema
`queue.parseDependsOnContent` quebra `nenhuma (usa …)` em palavras (a 89-01 nunca sairia de `forge queue next`); `pipeline.readDependsOn` lê certo, mas só reconhece `feat-…` (ignora `fix-88-01` e `88-01`). Dois leitores, `DUP-01`.

## Comportamento
1. `internal/queue` vira o único dono: aceita `feat-90-02`, `fix-88-01` e `88-02`; `nenhuma`, `—` e texto livre entre parênteses não geram dependência. `waves.go` passa a chamá-lo.
2. Fila antiga com dependências lixo (`"nenhuma"`, `"(usa"`) é corrigida na leitura, sem editar o arquivo.
3. Checagem por task no `dispatch` e no `/toca-ai`: dependência que não existe na iniciativa ou critério sem comando executável bloqueia o despacho, listando de uma vez tudo o que falta.

## Modelo do módulo
**Enxuto** (ARQ-11): estende `internal/queue` e `internal/pipeline`.

## Decisões do projeto aplicadas
ADR-0004 (gate pré-build), ADR-0014.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Leitor da fila | `queue.parseDependsOnContent` | Divide em palavras | substituir | Bug |
| Leitor das ondas | `waves.readDependsOn`, `taskRefRe` | Só `feat-` | apagar; chamar o da fila | `DUP-01`, `DUP-06` |
| Seções da spec | `dispatch.missingSections` | Confere seções | estender por task | Mesmo portão |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`DUP-01`, `DUP-06`, `COD-04`, `TST-03`. Rules: `sdd-reuso-e-duplicacao`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
go test ./internal/queue ./internal/pipeline -run 'TestDependeDe|TestPlanWaves_MesmoLeitorDaFila'
go test ./internal/pipeline -run 'TestDispatch_SpecIncompleta|TestGoal_SpecIncompleta'
go vet ./... && go test ./...
```
- `**Depende de:** nenhuma (usa …)` não gera dependência; `feat-89-01 e 88-02` gera duas; `fix-88-01` é aceito.
- Fila e ondas dão o mesmo resultado para a mesma spec.
