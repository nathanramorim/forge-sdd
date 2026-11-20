# feat-104-06 — Laço no Go: lock, retorno estruturado e spike de aninhamento

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** implementada em 2026-11-20
**Depende de:** feat-104-01

## Problema
O fluxo, os limites e a ordem precisam morar em código determinístico, não em prompt. Falta lock por iniciativa, retomada após falha no meio e tratamento de falha parcial no leque (A passou, B falhou). Também falta provar, num spike, se um subagente cria outro (decide o desenho).

## Comportamento
1. `pipeline.RunChat` conduz o ciclo `spec -> build -> review -> retorno` com estados gravados em `.forge/runtime/`; o modelo só devolve conteúdo pelo contrato (feat-104-01).
2. Lock por iniciativa (arquivo com dono e lease, `os.Rename` atômico como no `bus`); segundo coordenador na mesma iniciativa é recusado.
3. Retorno: `ApplyVerdict` devolve achados (`arquivo`, `linha`) ao Builder até `max_rework_attempts` (3); no teto escala ao humano pelo inbox; `retomar` continua do último estado gravado.
4. Leque: o resultado de cada subagente é gravado; falha parcial deixa o lote em `parcial`, repete só o que falhou.
5. Spike (documentado em `sdd/discovery/104-.../spike-aninhamento.md`): tentar criar subagente a partir de subagente e de teammate; o resultado decide o desenho A/D da ADR-0015.

## Modelo do módulo
**Completo**: `internal/pipeline` (laço) usa `internal/harness` (contratos, lock) e `internal/bus` (jobs).

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0012, ADR-0013 (modo avançado), FLX-08.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Jobs e estados | `internal/bus` | Claim atômico com lease, `Recover`, `Renew` | **Reusar** para o lock | `DUP-01` |
| Retorno | `internal/pipeline/review.go` (`ApplyVerdict`), `pr_rework` | Teto `max_rework_attempts`, escalada | **Reusar**, só estruturar os achados | ADR-0004 D-3 |
| Ondas | `internal/pipeline/waves.go` (`PlanWaves`) | Lotes por dependência | **Reusar** | — |
| Worker headless | `internal/pipeline/worker.go` | Executa jobs | **Reusar** no modo `sessions` | Convivência |
| Lock por iniciativa | — | Não existe (workers e outra sessão no mesmo worktree) | **Criar** em `harness` | Dor real |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-02`, `COD-01`, `COD-03`, `FLX-01`, `FLX-08`, `TST-01`, `TST-05`. Rules: `sdd-arquitetura`, `sdd-fluxo-e-portao`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestLockPorIniciativa|TestRetornoEscalaNoTeto|TestLegueParcialRepeteSoFalho|TestRetomaDoUltimoEstado' ./internal/pipeline | grep -F 'TestLockPorIniciativa' >/dev/null
go test -list 'TestRetornoEscalaNoTeto' ./internal/pipeline | grep -F 'TestRetornoEscalaNoTeto' >/dev/null
go test ./internal/pipeline -run 'TestLockPorIniciativa|TestRetornoEscalaNoTeto|TestLegueParcialRepeteSoFalho|TestRetomaDoUltimoEstado'
# negativo: segundo coordenador na mesma iniciativa falha; na 4a devolução escala
go test ./internal/pipeline -run 'TestLockPorIniciativa|TestRetornoEscalaNoTeto'
go vet ./... && go test ./...
```
- Teste negativo: segundo coordenador na mesma iniciativa falha; na 4a devolução escala em vez de voltar.
- Spike registrado com comando e saída literais em `spike-aninhamento.md`.

## Decisões da implementação
- **Teto de voltas continua 3:** `LimiteDeVoltas` lê o `max_rework_attempts` de `forge.json` (padrão 3). Com teto 3, as 3 primeiras violações voltam ao Builder e a 4ª escala, como em `ApplyVerdict`.
- **Spike não muda o desenho:** a evidência local não contradiz a opção A (ver `spike-aninhamento.md`); D e C ficam sem decisão, pendentes de execução real fora de `go test` e de nova ADR.
- **Lock em `harness`, não no `bus`:** o `bus` reivindica jobs, não iniciativas. O lock é publicado por `os.Link` (atômico, sem sobrescrever) e `os.Rename` para tomar órfão, com lease de 15 min renovada a cada task. Processo morto: o arquivo `.forge/runtime/chat/locks/<iniciativa>.lock` fica até a lease vencer; depois o próximo coordenador o recupera sozinho. Não se checa PID vivo (portabilidade).
- **Escalada pelo retorno, não pelo inbox:** o laço devolve `Retorno` com achados (arquivo e linha, sanitizados, FLX-08) e o menu (`Opcoes`, `Recomendada`, RN-9); notificar a pessoa fica para a camada de entrada (feat-104-09), para não duplicar `ApplyVerdict` do modo `sessions`.
- **Leque sequencial:** as tasks do lote rodam em ordem; paralelismo com worktree por subagente é a feat-104-07. Falha de build ou review deixa a task em `falhou` e o lote em `parcial`; `Retomar` repete só o que falhou ou parou no meio.
- Estado gravado em `.forge/runtime/chat/<iniciativa>.json`.
