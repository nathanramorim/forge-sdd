# Plano — Iniciativa 89 (de baixo para cima)

## Onda 1
- **89-01 Motor do `/toca-ai`** — `internal/pipeline/goal.go`, `forge pipeline toca-ai`, `--preflight`, `consent`. Rules: ARQ-01, ARQ-06, SEG-01, SEG-03, COD-03, TST-01. Deps: —. CA: `go test ./internal/pipeline -run 'TestGoal|TestPreflightGoal|TestConsent|TestPendencias'`

## Onda 2
- **89-02 Comando `/toca-ai` e skill** — referência, adaptadores, templates, golden, cheatsheet, docs. Rules: ARQ-01, DUP-01, COD-05. Deps: 89-01. CA: `go test ./internal/scaffold -run 'TestCheatsheet|TestGolden|TestCommands'`
