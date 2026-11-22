# feat-104-08 — Papéis Discovery, Builder e Reviewer no chat, com permissão pelo harness

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-07, feat-104-05

## Problema
O dono quer o fluxo no mesmo chat, cada papel como líder que coordena subagentes. Hoje as ferramentas por papel vivem em `.claude/agents/` só para Claude; Codex e Agy dependem de pedido no prompt.

## Comportamento
1. Modo `pipeline_mode: chat`: o comando `/implementar` no chat chama `pipeline.RunChat` e instrui o chat a criar subagentes A/B por papel (um nível), conforme o resultado do spike.
2. Novos agentes `builder-a`, `builder-b`, `reviewer-a`, `reviewer-b` (até 420 bytes): Builder com `Edit,Write,Bash`, Reviewer só `Read,Grep,Glob,Bash` (`SEG-06`), sem `Agent`.
3. Retorno: parecer volta pelo contrato; o líder (papel no chat) só repassa os achados como dados (`FLX-08`).
4. Para Codex e Agy, o gerador escreve a restrição de ferramentas no perfil/config quando existir recurso equivalente; senão, registra no diagnóstico (feat-104-03) que a permissão é só por prompt.

## Modelo do módulo
**Enxuto**: estende `internal/scaffold` (agentes e adaptadores), `internal/pipeline` e `internal/config`.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0013 (Gate 1), ADR-0003 (420 bytes), SEG-06.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Agentes por papel | `.claude/agents/builder.md`, `reviewer.md`, `lider.md` | Ferramentas por papel, sem `Agent` | **Estender** com A/B | `SEG-06` |
| Skills | `.agents/skills/sdd-build`, `sdd-governanca`, `sdd-concepcao` | Instruções por estação | **Reusar** | `DUP-05` |
| `/implementar`, `forge dispatch`, `forge queue review` | `.claude/commands/implementar.md`, `internal/pipeline/dispatch.go` | Despacho e revisão | **Reusar** | — |
| Modo avançado | ADR-0013 | Autonomia | **Reusar** | Gate 1 igual |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `FLX-01`, `FLX-08`, `SEG-06`, `COD-01`, `TST-01`. Rules: `sdd-fluxo-e-portao`, `sdd-arquitetura`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestAgentesABSemFerramentaAgent|TestReviewerABSomenteLeitura|TestAgentesAte420Bytes' ./internal/scaffold | grep -F 'TestReviewerABSomenteLeitura' >/dev/null
go test ./internal/scaffold -run 'TestAgentesABSemFerramentaAgent|TestReviewerABSomenteLeitura|TestAgentesAte420Bytes'
# negativo: reviewer com Edit, Write ou Agent reprova
go test ./internal/scaffold -run 'TestReviewerABSomenteLeitura|TestAgentesABSemFerramentaAgent'
go vet ./... && go test ./...
```
- Teste negativo: definição de reviewer com `Edit` ou `Write` reprova; com `Agent` reprova.

## A confirmar no Gate 1
Se o spike mostrar aninhamento, a spec passa a prever líder-agente (desenho D)?
