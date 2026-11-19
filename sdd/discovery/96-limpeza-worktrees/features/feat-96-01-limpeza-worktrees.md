# feat-96-01 — Deteccao e limpeza de worktrees ociosos na inicializacao da sessao

**Iniciativa:** 96-limpeza-worktrees
**Branch:** feat/96-limpeza-worktrees
**Status:** ready
**Depende de:** nenhuma

## Modelo do modulo

- **Modelo:** enxuto.
- **Justificativa:** a logica de GC pertence a `internal/pipeline` (ja dono de `CreateWorktree`,
  `RemoveWorktree` e `WorktreePath`; `DUP-01`, conceitos.md). Nenhum pacote novo. Um arquivo
  `worktree_gc.go` e um arquivo de testes `worktree_gc_test.go` sao adicionados ao mesmo pacote. A
  integracao no CLI e feita em `cmd/forge-sdd/up.go` (hook pre-launch) e em um novo subcomando
  `forge pipeline worktree gc` (estende o `pipelineWorktreeCmd` existente em `worker.go`).

## Decisoes do projeto aplicadas

- ADR-0004 (Pipeline Multi-Sessao, `internal/bus` como barramento, `internal/queue` como fila de
  tasks; workers nao promovem interacao, so relatam).
- ADR-0010 (`forge pipeline worktree gc`, nao um novo toplevel).
- ARQ-01: logica de GC fica em `internal/pipeline`, nao em `cmd/`.
- ARQ-02: `internal/pipeline` ja importa `internal/bus` e `internal/queue`; sem ciclo novo.
- ARQ-04: nenhuma dependencia nova de runtime.
- ARQ-05: `--dry-run` lista worktrees ociosos sem remover nada.
- SEG-01: caminhos de worktree validados via `WorktreePath` (regex existente) antes de qualquer
  operacao. `RemoveWorktree` ja verifica que o caminho esta dentro de `.forge/runtime/worktrees`.
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## O que ja existe

| Procurei | Onde | Achei | Decisao |
|---|---|---|---|
| Remocao de worktree | `internal/pipeline/worktree.go` (`RemoveWorktree`) | Funcao completa com validacao de path e branch | reusar direto; nao re-implementar |
| Criacao de worktree | `internal/pipeline/worktree.go` (`CreateWorktree`) | Funcao completa | sem alteracao |
| Caminho dos worktrees | `internal/layout/layout.go` (`Worktrees`) | `.forge/runtime/worktrees` | reusar |
| Estados de job no barramento | `internal/bus/job.go`, `internal/layout/layout.go` | `StatePending`, `StateProcessing`, `StateDone`, `StateFailed` | reusar como criterio de protecao |
| Estados de task na fila | `internal/queue/queue.go` | `StatusQueued`, `StatusBuilding`, `StatusReviewPending`, `StatusApproved`, `StatusChangesRequested`, `StatusBlocked` | reusar; task e segura so quando `StatusApproved` |
| Listagem de jobs por estado | `internal/bus/store.go` (`List`) | `bus.List(dir, role, state)` | reusar para detectar jobs `processing` |
| Leitura da fila de tasks | `internal/queue/queue.go` (`Load`) | Carrega `queue.json` de uma iniciativa | reusar |
| Comando worktree CLI | `cmd/forge-sdd/worker.go` (`pipelineWorktreeCmd`) | `forge pipeline worktree add/remove` | estender com subacao `gc` |
| Verificacao de TTY | `cmd/forge-sdd/menu.go` (`isInteractive`) | Variavel injetavel para testes | reusar; GC nunca promove se nao-interativo |
| `--dry-run` | `cmd/forge-sdd/up.go`, convencao geral | Flag Cobra padrao | reusar; GC em dry-run so lista |
| Limpeza de metadados git | `git worktree prune` | Sem chamada atual pos-remocao | novo: chamar apos cada `RemoveWorktree` bem-sucedido |

## Dependencias novas

- Nenhuma (`ARQ-04`, `SEG-07`). `internal/pipeline` ja depende de `internal/bus`, `internal/queue`
  e `internal/layout`. O unico comando git novo e `git worktree prune`, chamado via
  `layout.Git(dir, "worktree", "prune")`, sem flags de caminho (sem risco de path injection).

## Rules e IDs da Constituicao Aplicaveis

- `ARQ-01`, `ARQ-02`, `ARQ-04`, `ARQ-05`, `ARQ-10`, `ARQ-11`
- `COD-01`, `COD-03`, `COD-06`
- `DUP-01`, `DUP-02`
- `SEG-01`, `SEG-07`
- `TST-01`, `TST-02`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`,
  `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo

### 1. `internal/pipeline/worktree_gc.go` — funcao de deteccao

```go
// StaleWorktree representa um worktree encontrado em disco com o estado do job associado.
type StaleWorktree struct {
    JobID      string
    TaskID     string
    Path       string   // caminho absoluto
    JobState   string   // bus state; "" se nao encontrado (orphan)
    TaskStatus string   // queue status; "" se nao encontrado
    Safe       bool     // verdadeiro se pode ser removido
    Reason     string   // motivo legivel para o usuario
}

// ListStaleWorktrees varre .forge/runtime/worktrees/<jobID>/<taskID> e classifica cada worktree
// como seguro (Safe=true) ou protegido/orphan (Safe=false).
//
// Regra de seguranca (task pode ser removida quando TODOS os criterios abaixo sao verdadeiros):
//  1. O job existe no barramento com estado "done" para o papel builder OU reviewer.
//  2. A task tem status "approved" em queue.json da iniciativa do job.
//     Se queue.json nao existir, a task e tratada como terminal (aprovada).
//  3. Nenhum job com o mesmo Feature em qualquer papel esta em "processing" no barramento.
//
// Jobs em "failed" nao sao oferecidos (podem ser retentados; evita perda de contexto).
// Worktrees sem job correspondente no barramento (orphans) sao listados com Safe=false,
// Reason="job nao encontrado no barramento".
func ListStaleWorktrees(dir string) ([]StaleWorktree, error)
```

**Algoritmo detalhado:**

1. Chama `os.ReadDir(layout.Worktrees(dir))` para listar subdiretorios de primeiro nivel
   (`<jobID>`). Se o diretorio nao existir, retorna slice vazio sem erro.
2. Para cada `<jobID>`, chama `os.ReadDir(filepath.Join(wtRoot, jobID))` para listar `<taskID>`.
3. Para cada par `(jobID, taskID)`, constroi `path = layout.Worktrees(dir)/<jobID>/<taskID>` e
   verifica que e um diretorio (nao um arquivo ou link que aponte para fora — ja coberto por
   `RemoveWorktree`, mas filtrar aqui evita exibir lixo ao usuario).
4. Busca o job no barramento:
   - `bus.List(dir, layout.RoleBuilder, layout.StateDone)`
   - `bus.List(dir, layout.RoleBuilder, layout.StateProcessing)`
   - `bus.List(dir, layout.RoleReviewer, layout.StateDone)`
   - `bus.List(dir, layout.RoleReviewer, layout.StateProcessing)`
   - Filtra pelo `jobID`. Se nao encontrado em nenhum estado, marca como orphan.
5. Se encontrado em `processing` (qualquer papel) → `Safe=false`, `Reason="job em execucao"`.
6. Se encontrado em `failed` → `Safe=false`, `Reason="job falhou; pode ser retentado"`.
7. Se encontrado em `done`:
   - Le queue.json via `queue.Load(layout.InitiativeQueuePath(dir, job.Initiative))`.
   - Se falhar ao ler queue.json, trata como terminal (Safe para remoção).
   - Verifica o status da task `taskID` na fila; so permite remoção se `StatusApproved`.
   - Se task esta em estado nao-terminal (`queued`, `building`, `review_pending`,
     `changes_requested`, `blocked`) → `Safe=false`, `Reason="task ainda na fila (<status>)"`.
   - Verifica adicionalmente: se qualquer job com o mesmo `job.Feature` esta em `processing`
     (resultado do passo 4 reutilizado) → `Safe=false`,
     `Reason="outra tarefa do mesmo feature em execucao"`.
   - Se nao ha restricao → `Safe=true`, `Reason="job concluido e todas as tasks aprovadas"`.

### 2. `cmd/forge-sdd/up.go` — hook pre-launch

Antes de chamar `env.LaunchAll(...)`, quando `isInteractive() && !dry`:

```go
if err := runWorktreeGCSweep(cmd.OutOrStdout(), dir, false); err != nil {
    // log de aviso, nao bloqueia o launch
}
```

`runWorktreeGCSweep` chama `pipeline.ListStaleWorktrees(dir)`, filtra os `Safe=true` e, se houver
pelo menos um, exibe a lista com `huh.NewSelect` (ou `huh.NewMultiSelect` se disponivel) e remove
os escolhidos via `pipeline.RemoveWorktree`. Apos cada remocao bem-sucedida, chama
`layout.Git(dir, "worktree", "prune")`. Em dry-run, apenas imprime a lista.

A confirmacao nao bloqueia `forge up` em caso de recusa ou erro: o pipeline abre normalmente.

**Opcoes apresentadas ao usuario:**
- `[x] Limpar <jobID>/<taskID> (job <jobID> — concluido e aprovado)` — um item por worktree seguro
- `[ ] Manter todos`

### 3. `forge pipeline worktree gc` — subcomando CLI

Novo `ValidArgs` `gc` no `pipelineWorktreeCmd` existente:

```
forge pipeline worktree gc [--dir <dir>] [--dry-run]
```

- Sem TTY (`isInteractive() == false`): lista todos os stale worktrees em stdout (formato
  `SAFE|PROTECTED <jobID>/<taskID> <reason>`) e retorna sem remover. Exit 0.
- Com TTY e sem `--dry-run`: mesma interacao do hook de `forge up`.
- Com `--dry-run`: lista, sem promover interacao.

**Integracao com `forge status`:** a saida de `forge status` ja chama `forge ps` internamente;
nao e necessario alterar `status.go` nesta feature. Worktrees ociosos sao visivels via
`forge pipeline worktree gc --dry-run`.

### 4. Comportamento em cenarios especiais

| Cenario | Comportamento |
|---|---|
| `--dry-run` | Lista todos os stale (safe e protegidos), sem remover |
| Sem TTY / autopilot | Apenas lista safe em stdout, sem prompt, sem remover |
| Worktree orphan (sem job) | Listado com `PROTECTED`, nunca oferecido para remocao automatica |
| Job `failed` | Listado com `PROTECTED`, nunca oferecido |
| Task `review_pending` | Protegido; PR pode retornar como `pr_rework` |
| Erro em `RemoveWorktree` | Aviso no stderr; continua com os proximos; nao aborta |
| `.forge/runtime/worktrees` inexistente | Retorna vazio silenciosamente |
| `git worktree prune` falha | Aviso no stderr; nao aborta |

## Criterios de Aceitacao Executaveis

```bash
go test -v -race ./internal/pipeline/... -run TestWorktreeGC
go test -v -race ./cmd/forge-sdd/... -run TestPipelineWorktreeGCCmd
go vet ./...
```

Cenarios obrigatorios (cada um = um `Test*` independente):

1. **TestWorktreeGC_JobDoneTaskAprovada**: worktree presente, job em `done` no barramento e task
   `approved` em queue.json → `Safe=true`.
2. **TestWorktreeGC_JobProcessing**: job em `processing` no barramento → `Safe=false`,
   reason contem "em execucao".
3. **TestWorktreeGC_JobFailed**: job em `failed` → `Safe=false`, reason contem "falhou".
4. **TestWorktreeGC_TaskReviewPending**: job `done` mas task `review_pending` → `Safe=false`,
   reason contem "review_pending".
5. **TestWorktreeGC_Orphan**: nenhum job com esse ID encontrado no barramento → `Safe=false`,
   reason contem "nao encontrado".
6. **TestWorktreeGC_SemWorktreeDir**: `.forge/runtime/worktrees` nao existe → retorna slice vazio,
   sem erro.
7. **TestWorktreeGC_FeatureIrmaEmProcessing**: job `done` e task `approved`, mas outro job com
   mesmo `Feature` esta `processing` → `Safe=false`.
8. **TestPipelineWorktreeGCCmd_DryRun**: `forge pipeline worktree gc --dry-run` lista sem remover.
9. **TestPipelineWorktreeGCCmd_SemTTY**: sem TTY imprime lista e retorna exit 0 sem prompt.

Todos os testes passam em `go test -race`. `go vet ./...` limpo.
