# feat-104-01 — Contratos entre etapas validados em código

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** nenhuma

## Problema
Skill e binário divergem: a Governança exige `regra_descricao` e o `forge queue review` do binário usado recusou o campo ("unknown field"). Nada impede a divergência de voltar.

## Comportamento
1. `internal/harness` define um esquema por passagem: `SpecParaBuild` (Discovery->Builder), `ResultadoBuild` (Builder->Reviewer) e `Parecer` com `Retorno` (achados com `arquivo`, `linha`, `evidencia`, `regra`, `regra_descricao`).
2. Leitura estrita (`DisallowUnknownFields`) e escrita só por `harness.Validate*`; `queue.Verdict` passa a ser o tipo do `Parecer` (um dono).
3. Teste de contrato: o exemplo JSON de `.agents/skills/sdd-governanca/references/veredito.md` é extraído e validado pelo esquema; skill e binário não divergem sem teste vermelho.
4. Mensagem de erro cita o campo e o contrato.

## Modelo do módulo
**Completo** (`ARQ-11`): pacote novo `internal/harness` (contratos), com `queue.Verdict` reexportado; sem ciclo (`ARQ-02`).

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento das skills e agentes), mais: ADR-0012 (veredito gravado pelo Go), ADR-0013.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Parecer e veredito | `internal/queue/queue.go` (`Verdict`, `RegraDescricao`) | Esquema com o campo na `main` | **Reusar** como fonte do contrato | `DUP-01` |
| Exemplo de veredito | `.agents/skills/sdd-governanca/references/veredito.md` | JSON com `regra_descricao` | **Reusar** no teste de contrato | Evita deriva |
| Aplicação do veredito | `internal/pipeline/review.go` (`ApplyVerdict`) | Máquina de estados | **Reusar**, passando pelo esquema | `ADR-0004` |
| Catálogo de conceitos | `sdd/spec/concepts.md` | Sem linha de contratos | **Adicionar** linha | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-02`, `ARQ-11`, `COD-01`, `COD-03`, `DUP-01`, `TST-01`, `TST-05`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestContratoVeredito|TestContratoRecusaCampoDesconhecido' ./internal/harness | grep -q 'TestContratoVeredito'
go test ./internal/harness -run 'TestContratoVeredito|TestContratoRecusaCampoDesconhecido'
go vet ./... && go test ./...
```
- Teste negativo: parecer com campo desconhecido é recusado com o nome do campo.
- Teste de contrato: o JSON de exemplo da skill passa no esquema.

## A confirmar no Gate 1
O `regra_descricao` hoje já existe na `main`; a falha vista pelo dono veio de um binário mais antigo? Se o binário em uso for anterior, a feature só garante que não volta.
