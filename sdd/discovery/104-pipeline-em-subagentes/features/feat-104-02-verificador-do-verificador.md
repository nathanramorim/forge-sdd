# feat-104-02 — Verificar o verificador: teste negativo nos critérios

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** implementada (2026-11-20)
**Depende de:** nenhuma

## Problema
Três critérios erraram por defeito do próprio teste: `grep -c` contando linhas num SVG de uma linha, pipeline que só passa sem saída e quebra sob `bash -e`, e extrator que juntou o bloco "como regenerar" aos critérios. Critério que nunca falha não prova nada.

## Comportamento
1. `forge gate --spec <spec> --criterios` (check `P-25`) extrai só o bloco `## Critérios de Aceitação Executáveis`, roda `bash -n` (e, com `--executar-criterios`, `bash -e -o pipefail`) e exige a presença de uma linha de teste negativo (marcador `# negativo:`) com resultado esperado de falha.
2. Rejeita padrões ruins conhecidos: `grep -c` sem `-o`/`wc`, pipeline sem `set -o pipefail`, `! cmd` sem `|| exit 1`, pipeline terminado em `grep -v`, bloco que não é o dos critérios.
3. O `gate` só lê (`ARQ-05`); não há `--dry-run` porque nada é gravado. Lógica em `internal/gate`, comando só adapta.

## Modelo do módulo
**Enxuto** (`ARQ-11`): estende `internal/gate` (já dono do portão e de `P-xx`); flag nova no `forge gate` em `cmd/forge-sdd`, sem subcomando novo.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento das skills e agentes), mais: ADR-0014 (critério executável é fonte de verdade).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Extrator de critérios | `internal/gate`, `internal/pipeline/specheck.go` | Lê a seção da spec | **Estender** com sintaxe e negativo | `DUP-01` |
| Portão Camada A | `forge gate` (`P-xx`) | Valida seções, não a qualidade do critério | **Estender** com nova checagem | Mesmo portão |
| Subcomandos | `cmd/forge-sdd/gate.go` | `forge gate` com flags | **Estender** com `--criterios` e `--executar-criterios` (`ARQ-06`) | Decisão 8 do Gate 1 |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-05`, `ARQ-06`, `COD-01`, `TST-01`, `TST-05`. Rules: `sdd-arquitetura`, `sdd-fluxo-e-portao`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestCriteriaVerifyRejeitaGrepC|TestCriteriaVerifyExigeNegativo|TestCriteriaVerifyIgnoraComoRegenerar' ./internal/gate | grep -F 'TestCriteriaVerifyExigeNegativo' >/dev/null
go test ./internal/gate -run 'TestCriteriaVerifyRejeitaGrepC|TestCriteriaVerifyExigeNegativo|TestCriteriaVerifyIgnoraComoRegenerar'
# negativo: spec com grep -c num SVG de uma linha e spec sem marcador negativo são reprovadas
go test ./internal/gate -run 'TestCriteriaVerifyRejeitaGrepC|TestCriteriaVerifyExigeNegativo'
go run ./cmd/forge-sdd gate --dir . --criterios --spec internal/gate/testdata/criterios/grep-c-svg.md | grep -F '✗ [P-25]' >/dev/null
go vet ./... && go test ./...
```
- Teste negativo: spec cujo critério usa `grep -c` num SVG de uma linha é reprovada; spec sem `# negativo:` é reprovada.
- Os critérios das features desta iniciativa passam em `forge gate --criterios`.

## Decisões da implementação
- Flag `--criterios` no `forge gate` (decisão 8 do Gate 1); o resultado é o check `P-25` (`TST-01`), que só roda com a flag e com `--spec`.
- Execução com `bash -e` fica na flag opcional `--executar-criterios`: a checagem padrão é estática para não rodar `go test ./...` dentro do portão nem reprovar specs ainda não implementadas.
- `grep -c` é reprovado sempre (com ou sem `-o`, pois `-c` conta linhas mesmo com `-o`); a forma aceita é `grep -o ... | wc -l`.
- O marcador `# negativo:` vale em linha própria (seguido do comando) ou inline; o portão confere a presença e o comando, não a semântica do teste negativo.
- Nas 12 specs: `set -o pipefail` no início, `grep -q` trocado por `grep -F ... >/dev/null` (evita SIGPIPE sob `pipefail`) e uma linha `# negativo:` com o teste do caso que deve falhar. Na feat-104-10 e na feat-104-12 o teste negativo ganhou nome próprio (`TestPipelineModeRecusaInvalido`, `TestTraceRecusaNoSemMotivo`).
- Ajuste após a Governança: o `P-25` exige que todo nome em `go test -run` de uma linha `# negativo:` apareça num `go test -list` do mesmo bloco (senão o critério passaria com teste inexistente); 104-10 e 104-12 passaram a listar seus testes negativos. A detecção de pipefail aceita qualquer `set` com `-o pipefail` (ex.: `set -euo pipefail`, `set -o errexit -o pipefail`).

## Status
- [x] Implementada em 2026-11-20 (branch `feat/104-02-verificador`).
