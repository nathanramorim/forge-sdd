# feat-104-03 — forge doctor --harness --json

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-01

## Problema
Nada confere o harness ao abrir a sessão: versão do binário, rules e skills sincronizadas, divergência skill x binário, worktrees órfãos, locks, MCPs e modelos.

## Comportamento
1. `forge doctor --harness --json` emite lista de checagens (`id`, `status` ok|aviso|falha, `detalhe`, `sugestao`), sem corrigir nada (RN-7).
2. Checagens: versão (`forge.json` x binário), rules/skills sincronizadas com o scaffold, contrato skill x binário (feat-104-01), worktrees órfãos em `.forge/runtime/worktrees/`, locks vencidos, MCPs de `sdd/memory/mcps.md`, modelos de `forge.json`.
3. Saída em até 2 KB sem terminal; sai com código 0 (ok/aviso) ou 3 (falha).
4. `/doctor` (skill `doctor.md`) passa a citar a flag.

## Modelo do módulo
**Completo** via `internal/harness` (feat-104-01): `harness.Diagnose`; `cmd/forge-sdd/doctor.go` só liga a flag (`ARQ-01`).

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0010 (comando único `forge`).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| `forge doctor` | `cmd/forge-sdd/doctor.go`, `.agents/skills/sdd-operacao/references/doctor.md` | Estrutura, orçamentos, telemetria | **Estender** | `/doctor` único |
| Worktrees | `internal/pipeline/worktree_gc.go` | Coleta de órfãos | **Reusar** a detecção | `DUP-01` |
| MCPs | `sdd/memory/mcps.md` | Estado dos MCPs | **Reusar** a leitura | Fonte única |
| Modelos | `internal/config/forge_profiles.go` | `Models`, `Efforts` | **Reusar** | `DUP-06` |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-05`, `COD-03`, `DUP-01`, `TST-01`. Rules: `sdd-arquitetura`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestDiagnoseJSON|TestDiagnoseSemEfeitoColateral|TestDiagnoseDetectaWorktreeOrfao' ./internal/harness | grep -F 'TestDiagnoseJSON' >/dev/null
go test ./internal/harness -run 'TestDiagnoseJSON|TestDiagnoseSemEfeitoColateral|TestDiagnoseDetectaWorktreeOrfao'
# negativo: worktree órfão e contrato divergente produzem falha e código 3
go test ./internal/harness -run 'TestDiagnoseDetectaWorktreeOrfao'
go vet ./... && go test ./...
```
- Teste negativo: worktree órfão e contrato divergente produzem `falha` e código 3.
- `--dry-run`/diagnóstico não altera a árvore (`git status --porcelain` igual antes e depois).

## A confirmar no Gate 1
Lista final de checagens e o limiar entre aviso e falha.
