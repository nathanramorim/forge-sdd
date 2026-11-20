# feat-104-04 — Servidor MCP forge_doctor, hook SessionStart e instrução

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-03

## Problema
A dor do dono: hoje só funciona no terminal; precisa alcançar o app desktop e o Codex. O diagnóstico precisa ser disparado nas três superfícies.

## Comportamento
1. `forge mcp serve` expõe por stdio a ferramenta `forge_doctor` (chama `harness.Diagnose`), somente leitura, em JSON-RPC sobre `internal/acp`.
2. `forge init/update` registra o servidor no `.mcp.json` do Claude Code e o hook `SessionStart` em `.claude/settings.json` (mesmo instalador do hook do inbox).
3. Onde não há hook (Codex, chat comum), o bloco do `AGENTS.md`/instruções manda chamar `forge_doctor` ou `forge doctor --harness` e avisar quando não rodou.
4. Suposição a validar: o app desktop do Claude Code usa os mesmos hooks e MCPs; se não, vale só a instrução.

## Modelo do módulo
**Enxuto**: estende `internal/acp` (transporte) e `internal/scaffold` (hook e instruções).

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0010; ARQ-04 (sem dependência).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Transporte JSON-RPC | `internal/acp/transport_stdio.go` | stdio pronto | **Reusar** | Sem dependência |
| Hook do inbox | `internal/scaffold` (`hooks_test.go`, `.claude/settings.json`) | `UserPromptSubmit` | **Estender** com `SessionStart`, idempotente | `DUP-01` |
| Instruções por agente | `internal/scaffold/templates` (`AGENTS.md`, `CLAUDE.md`) | Bloco por agente | **Estender** | Mesmo gerador |
| MCPs do projeto | `sdd/memory/mcps.md` | `context7`, `git` | **Reusar** formato | — |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-04`, `ARQ-06`, `COD-01`, `SEG-04`, `SEG-06`, `TST-01`. Rules: `sdd-acp`, `sdd-arquitetura`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestMCPForgeDoctor|TestHookSessionStartIdempotente|TestInstrucaoDoctorNoAgentsMd' ./internal/acp ./internal/scaffold | grep -q 'TestMCPForgeDoctor'
go test ./internal/acp ./internal/scaffold -run 'TestMCPForgeDoctor|TestHookSessionStartIdempotente|TestInstrucaoDoctorNoAgentsMd'
go vet ./... && go test ./...
```
- Teste negativo: chamada de ferramenta desconhecida devolve erro JSON-RPC; segundo `update` não duplica o hook.
- Manual (não automatizável aqui): abrir o app desktop e o Codex e conferir o aviso.

## A confirmar no Gate 1
O app desktop usa os mesmos hooks? (suposição). Nome do subcomando `mcp serve` (muda a lista de `ARQ-06`).
