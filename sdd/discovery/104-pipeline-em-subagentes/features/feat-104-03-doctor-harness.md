# feat-104-03 — forge doctor --harness --json

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-03-doctor
**Depende de:** feat-104-01

## Problema
Nada confere o harness ao abrir a sessão: versão do binário, rules e skills sincronizadas, divergência skill x binário, worktrees órfãos, locks, MCPs e modelos.

## Comportamento
1. `forge doctor --harness --json` emite lista de checagens (`id`, `status` ok|aviso|falha, `detalhe`, `sugestao`), sem corrigir nada (RN-7).
2. Checagens (`id`): `versao` (binário x `sdd/.sddrc`/`.sdd-version`), `forge-json` (leitura estrita, inclui modelos e esforços), `skills-rules` (5 skills de estação e rules presentes), `contrato-skill-binario` (reusa `harness.VerificarContratoParecer`), `pareceres-gravados` (decodifica `sdd/discovery/*/verdicts/*.json` com o leitor atual), `worktrees` (órfãos em `.forge/runtime/worktrees/` e entradas de `.claude/worktrees/`) e `workers` (processos `forge worker` de versão diferente).
3. Saída em até 2 KB sem terminal; sai com código 0 (ok/aviso) ou 3 (falha).
4. `/doctor` (skill `doctor.md`) passa a citar a flag.

## Modelo do módulo
**Completo** em `internal/doctor` (`doctor.Diagnose`), que reusa os contratos de `internal/harness` (feat-104-01); `cmd/forge-sdd/doctor.go` só liga a flag (`ARQ-01`).

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
go test -list 'TestDiagnoseJSON|TestDiagnoseSemEfeitoColateral|TestDiagnoseDetectaWorktreeOrfao|TestDiagnoseContratoDivergenteFalha' ./internal/doctor | grep -F 'TestDiagnoseContratoDivergenteFalha' >/dev/null
go test ./internal/doctor -run 'TestDiagnoseJSON|TestDiagnoseSemEfeitoColateral|TestDiagnoseDetectaWorktreeOrfao'
# negativo: contrato divergente produz falha e código 3
go test ./internal/doctor -run 'TestDiagnoseContratoDivergenteFalha'
go vet ./... && go test ./...
```
- Teste negativo: contrato divergente produz `falha` e código 3; binário mais antigo que o projeto também falha.
- Worktree órfão, parecer ilegível, worker de outra versão e skills ausentes são `aviso` (não bloqueiam).
- `--dry-run`/diagnóstico não altera a árvore (`git status --porcelain` igual antes e depois).

## Decisões da implementação
- **Pacote:** `internal/doctor`, não `internal/harness`: `harness` é importado por `queue`, que `pipeline` importa; `Diagnose` precisa de `pipeline.ListStaleWorktrees` (`DUP-01`) e fecharia um ciclo. O contrato skill x binário ficou em `harness.VerificarContratoParecer`, reusado pelo teste da 104-01.
- **Limiar:** `falha` só quando o harness não pode rodar certo (binário mais antigo que o projeto, `sdd/.sddrc` ilegível, `forge.json` inválido, contrato skill x binário divergente). O resto é `aviso`; o doctor sempre sugere o passo e nunca executa.
- **Locks vencidos e MCPs:** fora desta entrega. O código não grava arquivos de lock (só mutex em memória) e `sdd/memory/mcps.md` é tabela livre sem leitor; sem fonte determinística, a checagem seria inventada. Reabrir se surgir lock em disco.
- **Modelos:** cobertos por `forge-json`, pois `LoadForgeConfig` já valida modelos e esforços.
- **Workers:** versão obtida executando `<exe> version` (3 s de limite) dos processos `forge worker` listados por `ps`; falha ao obter a versão é ignorada.
- **Saída:** `{"versao":1,"checagens":[{id,status,detalhe,sugestao}],"falhas":N,"avisos":N}`, `detalhe` até 160 caracteres e no máximo 5 itens listados (sempre abaixo de 2 KB). `--dir` ou argumento posicional escolhe o projeto.
- **`ARQ-06`:** inalterado; `doctor` já está na lista.

## Status
Implementada em 2026-11-20.
