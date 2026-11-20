# feat-104-02 — Verificar o verificador: teste negativo nos critérios

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** nenhuma

## Problema
Três critérios erraram por defeito do próprio teste: `grep -c` contando linhas num SVG de uma linha, pipeline que só passa sem saída e quebra sob `bash -e`, e extrator que juntou o bloco "como regenerar" aos critérios. Critério que nunca falha não prova nada.

## Comportamento
1. `forge criteria verify <spec>` extrai só o bloco `## Critérios de Aceitação Executáveis`, roda `bash -n` e `bash -e` e exige a presença de uma linha de teste negativo (marcador `# negativo:`) com resultado esperado de falha.
2. Rejeita padrões ruins conhecidos: `grep -c` sem `-o`/`wc`, pipeline sem `set -o pipefail`, bloco que não é o dos critérios.
3. `--dry-run` não grava nada (`ARQ-05`). Lógica em `internal/gate`, comando só adapta.

## Modelo do módulo
**Enxuto** (`ARQ-11`): estende `internal/gate` (já dono do portão e de `P-xx`); comando novo sob `cmd/forge-sdd`.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento das skills e agentes), mais: ADR-0014 (critério executável é fonte de verdade).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Extrator de critérios | `internal/gate`, `internal/pipeline/specheck.go` | Lê a seção da spec | **Estender** com sintaxe e negativo | `DUP-01` |
| Portão Camada A | `forge gate` (`P-xx`) | Valida seções, não a qualidade do critério | **Estender** com nova checagem | Mesmo portão |
| Subcomandos | `cmd/forge-sdd` | Sem `criteria` | **Adicionar** `criteria verify` (`ARQ-06`) | Coberto pela ADR-0015 |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-05`, `ARQ-06`, `COD-01`, `TST-01`, `TST-05`. Rules: `sdd-arquitetura`, `sdd-fluxo-e-portao`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestCriteriaVerifyRejeitaGrepC|TestCriteriaVerifyExigeNegativo|TestCriteriaVerifyIgnoraComoRegenerar' ./internal/gate | grep -q 'TestCriteriaVerifyExigeNegativo'
go test ./internal/gate -run 'TestCriteriaVerifyRejeitaGrepC|TestCriteriaVerifyExigeNegativo|TestCriteriaVerifyIgnoraComoRegenerar'
go vet ./... && go test ./...
```
- Teste negativo: spec cujo critério usa `grep -c` num SVG de uma linha é reprovada; spec sem `# negativo:` é reprovada.
- Os critérios das features desta iniciativa passam em `forge criteria verify`.

## A confirmar no Gate 1
Nome do subcomando (`criteria verify` ou `gate --criterios`)? Padrão conservador: flag em `gate`, sem subcomando novo.
