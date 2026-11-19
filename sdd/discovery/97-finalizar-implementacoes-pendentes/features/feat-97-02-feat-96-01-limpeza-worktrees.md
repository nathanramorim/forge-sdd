# feat-97-02 — feat-96-01: detecção e limpeza de worktrees ociosos

**Iniciativa:** 97-finalizar-implementacoes-pendentes
**Status:** ready
**Tamanho:** M

## Objetivo

Ao iniciar qualquer sessão do forge-sdd (`forge up`, `forge worker`, `forge inbox`), verificar se há worktrees em `.forge/runtime/worktrees/` que podem ser removidos com segurança, e oferecer remoção interativa ao usuário.

## Escopo

**Nenhum pacote novo.** Dois arquivos Go novos no pacote `internal/pipeline` + integração em `cmd/`.

### 1. `internal/queue/queue.go` — helper novo

```go
// LoadForInitiative lê a fila de uma iniciativa específica pelo ID, sem depender da branch ativa.
// Retorna fila vazia (não erro) quando queue.json não existe.
func LoadForInitiative(dir, initiative string) (*Queue, error)
```

Implementação: `layout.InitiativeQueue(dir, initiative)` → `os.ReadFile` → `json.Unmarshal`.
`os.ErrNotExist` → retorna `&Queue{Items: []QueueItem{}}` sem erro.

### 2. `internal/pipeline/worktree_gc.go` — detecção e varredura

```go
type StaleWorktree struct {
    JobID      string
    TaskID     string
    Path       string   // caminho absoluto
    JobState   string   // estado no barramento; "" se orphan
    TaskStatus string   // status da task em queue.json; "" se não encontrada
    Safe       bool
    Reason     string
}

func ListStaleWorktrees(dir string) ([]StaleWorktree, error)
func SweepWorktrees(dir string, targets []StaleWorktree) []error
```

**Algoritmo de `ListStaleWorktrees`:**

1. `os.ReadDir(layout.Worktrees(dir))` → lista `<jobID>` subdirs. Dir ausente → slice vazio.
2. Para cada `<jobID>`, `os.ReadDir(wtRoot/<jobID>)` lista `<taskID>` dirs.
3. Para cada par `(jobID, taskID)`:
   - Busca job: `bus.Find(dir, layout.RoleBuilder, jobID)`, depois `layout.RoleReviewer` se não encontrado.
   - **Orphan** (não encontrado em nenhum role): `Safe=true`, `Reason="job não encontrado no barramento (orphan)"`.
   - **StatePending ou StateProcessing**: `Safe=false`, `Reason="job aguardando/em execução"`.
   - **StateFailed**: `Safe=false`, `Reason="job falhou; pode ser retentado"`.
   - **StateDone**: chama `queue.LoadForInitiative(dir, job.Job.Initiative)`, busca task por `taskID`.
     - Task não encontrada → `Safe=false`, `Reason="task não encontrada na fila (tratar como ativa)"`.
     - Task com status != `StatusApproved` → `Safe=false`, `Reason="task não aprovada (status: <status>)"`.
     - Task com `StatusApproved` → `Safe=true`, `Reason="job concluído e aprovado"`.

**`SweepWorktrees`:** itera `targets` onde `Safe=true`, chama `pipeline.RemoveWorktree(dir, t.Path)` para cada um. Acumula erros sem abortar. Ao final chama `git worktree prune` uma única vez (`layout.Git(dir, "worktree", "prune")`).

### 3. `internal/pipeline/worktree_gc_test.go`

- `TestListStaleWorktrees_DirAusente` → slice vazio
- `TestListStaleWorktrees_OrphanDetectado` → Safe=true
- `TestListStaleWorktrees_JobEmExecucao` → Safe=false
- `TestListStaleWorktrees_TaskAprovada` → Safe=true
- `TestSweepWorktrees_RemoveApenas SafeTrue` → remove Safe, não-toca não-Safe

### 4. `cmd/forge-sdd/up.go` — hook `checkStaleWorktrees`

Ao iniciar, antes do menu de sessões, `checkStaleWorktrees` apenas **informa** o usuário sobre worktrees ociosos encontrados, sem remover nem perguntar. A remoção real fica em `forge pipeline worktree gc` (§5), separando preocupações (ARQ-01).

```go
func checkStaleWorktrees(dir string, out io.Writer) {
    if !isInteractive() { return }
    stale, err := pipeline.ListStaleWorktrees(dir)
    if err != nil || len(stale) == 0 { return }
    var safe []pipeline.StaleWorktree
    for _, w := range stale { if w.Safe { safe = append(safe, w) } }
    if len(safe) == 0 { return }
    fmt.Fprintf(out, "\n%d worktree(s) ocioso(s) encontrado(s):\n", len(safe))
    for _, w := range safe {
        fmt.Fprintf(out, "  %s/%s — %s\n", w.JobID, w.TaskID, w.Reason)
    }
    fmt.Fprintln(out, "  → rode `forge pipeline worktree gc` para limpar.")
}
```

A mesma função é adicionada ao início de `cmd/forge-sdd/worker.go` (forge worker) e `cmd/forge-sdd/inbox.go` (forge inbox) — antes de qualquer ação principal.

### 5. `cmd/forge-sdd/worker.go` — subcomando `forge pipeline worktree gc`

Adicionar `gc` a `pipelineWorktreeCmd`:
```
forge pipeline worktree gc [--dry-run]
```
- Sem `--dry-run`: lista Safe=true, pede confirmação, chama `SweepWorktrees`.
- Com `--dry-run`: lista todos os worktrees ociosos detectados sem remover.

## Rules e IDs

`ARQ-01`, `ARQ-02`, `ARQ-04`, `ARQ-05`, `ARQ-10`, `ARQ-11`, `COD-01`, `COD-03`, `DUP-01`, `SEG-01`, `TST-01`, `TST-02`

## Critérios de Aceitação Executáveis

```bash
go build ./...
go test ./internal/pipeline/... -run TestWorktreeGC
go test ./internal/queue/... -run TestLoadForInitiative
forge pipeline worktree gc --dry-run
# deve listar worktrees ou imprimir "nenhum worktree ocioso"
```
