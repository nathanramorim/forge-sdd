# feat-96-01 — Detecção e limpeza de worktrees ociosos na inicialização da sessão

**Iniciativa:** 96-limpeza-worktrees
**Branch:** feat/96-limpeza-worktrees
**Status:** ready
**Depende de:** nenhuma

## Modelo do módulo

- **Modelo:** enxuto.
- **Justificativa:** a lógica de GC pertence a `internal/pipeline` (já dono de `CreateWorktree`,
  `RemoveWorktree` e `WorktreePath`; `DUP-01`, conceitos.md). Nenhum pacote novo. Dois arquivos
  (`worktree_gc.go` e `worktree_gc_test.go`) são adicionados ao pacote existente. A integração
  no CLI fica em `cmd/forge-sdd/up.go` (hook pré-launch) e num novo subcomando
  `forge pipeline worktree gc` (estende `pipelineWorktreeCmd` em `worker.go`). A lógica
  interativa de remoção fica em `pipeline.SweepWorktrees` (não em `cmd/`), respeitando `ARQ-01`.

## Decisões do projeto aplicadas

- ADR-0004 (Pipeline Multi-Sessão; workers não promovem interação, apenas relatam; `ARQ-06`
  lista os subcomandos do pipeline).
- ARQ-01: lógica de GC fica em `internal/pipeline`, não em `cmd/`.
- ARQ-02: `internal/pipeline` já importa `internal/bus` e `internal/queue`; sem ciclo novo.
- ARQ-04: nenhuma dependência nova de runtime.
- ARQ-05: `--dry-run` lista worktrees ociosos sem remover nada.
- SEG-01: caminhos validados via `WorktreePath` (regex existente) antes de qualquer operação;
  `RemoveWorktree` já verifica que o caminho está dentro de `.forge/runtime/worktrees`.
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Remoção de worktree | `internal/pipeline/worktree.go` (`RemoveWorktree`) | Função completa com validação de caminho e branch | reusar | Já valida path traversal e chama `git worktree remove --force` |
| Caminho dos worktrees | `internal/layout/layout.go` (`Worktrees`) | `.forge/runtime/worktrees` | reusar | Fonte única do caminho (DUP-01) |
| Busca de job por ID | `internal/bus/store.go` (`Find`) | `bus.Find(dir, role, id)` — percorre todos os estados | reusar | Busca atômica por role; chamar para builder e reviewer separadamente |
| Estados de job no barramento | `internal/layout/layout.go` | `StatePending`, `StateProcessing`, `StateDone`, `StateFailed` | reusar como critério de proteção | |
| Estados de task na fila | `internal/queue/queue.go` | `StatusQueued`, `StatusBuilding`, `StatusReviewPending`, `StatusApproved`, `StatusChangesRequested`, `StatusBlocked` | reusar; task segura apenas com `StatusApproved` | |
| Caminho do queue.json por iniciativa | `internal/layout/layout.go` (`InitiativeQueue`) | `layout.InitiativeQueue(dir, id)` devolve `sdd/discovery/<id>/queue.json` | reusar | Permite ler a fila de uma iniciativa específica sem depender da branch ativa |
| Carregamento da fila | `internal/queue/queue.go` (`Load`) | `queue.Load(dir)` lê da iniciativa ativa (via branch git) | **não adequado** — ler de iniciativa específica exige path direto | O Builder precisa verificar qualquer iniciativa, não só a ativa |
| Prompt MultiSelect | `internal/survey/survey.go` (`huh.NewMultiSelect`) | `huh.NewMultiSelect[string]()` já usado em `survey.go:52` e `survey.go:155` | reusar via `internal/survey` | `cmd/` nunca chama `huh` diretamente; `ARQ-01` |
| Subcomando worktree CLI | `cmd/forge-sdd/worker.go` (`pipelineWorktreeCmd`) | `forge pipeline worktree add/remove` | estender com ação `gc` | Evita novo toplevel (ADR-0004, ARQ-06) |
| Verificação de TTY | `cmd/forge-sdd/menu.go` (`isInteractive`) | Variável injetável para testes | reusar | GC nunca promove se não-interativo |
| Limpeza de metadados git | `git worktree prune` | Sem chamada atual pós-remoção | novo | Necessário para remover entradas obsoletas do `.git/worktrees` |

## Dependências novas

- Nenhuma dependência de runtime nova (`ARQ-04`, `SEG-07`). `internal/pipeline` já importa
  `internal/bus`, `internal/queue` e `internal/layout`. O único comando git novo é
  `git worktree prune`, chamado via `layout.Git(dir, "worktree", "prune")`, sem flags de
  caminho (sem risco de path injection).
- Novo helper interno: `queue.LoadForInitiative(dir, initiative string) (*Queue, error)` —
  lê `layout.InitiativeQueue(dir, initiative)` diretamente, sem depender da branch ativa.

## Rules e IDs da Constituição Aplicáveis

- `ARQ-01`, `ARQ-02`, `ARQ-04`, `ARQ-05`, `ARQ-10`, `ARQ-11`
- `COD-01`, `COD-03`, `COD-06`
- `DUP-01`, `DUP-02`
- `SEG-01`, `SEG-07`
- `TST-01`, `TST-02`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`,
  `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo

### 1. `internal/queue/queue.go` — helper novo

```go
// LoadForInitiative lê a fila de uma iniciativa específica pelo ID, independentemente da
// branch git ativa. Retorna fila vazia (não erro) quando queue.json não existe.
func LoadForInitiative(dir, initiative string) (*Queue, error)
```

Implementação: `layout.InitiativeQueue(dir, initiative)` → `os.ReadFile` → `json.Unmarshal`.
Se `ErrNotExist` → retorna `&Queue{Items: []QueueItem{}}` sem erro (fila vazia = sem tasks
ativas = conservador: o chamador verifica ausência de task como "não encontrada", não como
"aprovada"). Qualquer outro erro de I/O → retorna o erro (fail closed).

### 2. `internal/pipeline/worktree_gc.go` — detecção e varredura

```go
// StaleWorktree representa um worktree encontrado em disco com o estado do job associado.
type StaleWorktree struct {
    JobID      string
    TaskID     string
    Path       string   // caminho absoluto
    JobState   string   // estado no barramento; "" se não encontrado (orphan)
    TaskStatus string   // status da task em queue.json; "" se não encontrada
    Safe       bool     // verdadeiro se pode ser removido com segurança
    Reason     string   // mensagem legível para o usuário
}

// ListStaleWorktrees varre .forge/runtime/worktrees/<jobID>/<taskID> e classifica cada
// worktree encontrado. Escopo: somente worktrees criados pelo forge (caminhos que existem
// como subdiretórios dentro de layout.Worktrees(dir)).
func ListStaleWorktrees(dir string) ([]StaleWorktree, error)

// SweepWorktrees remove os worktrees listados (apenas os com Safe=true) e chama
// git worktree prune uma única vez ao final. Erros de remoção individual são acumulados
// e reportados sem abortar os demais.
func SweepWorktrees(dir string, targets []StaleWorktree) []error
```

**Algoritmo de `ListStaleWorktrees`:**

1. `os.ReadDir(layout.Worktrees(dir))` lista subdirs de primeiro nível (`<jobID>`).
   Se o diretório não existir → retorna slice vazio, sem erro.
2. Para cada `<jobID>`, `os.ReadDir(wtRoot/<jobID>)` lista `<taskID>` (apenas diretórios).
3. Para cada par `(jobID, taskID)`, constrói `path` e verifica que é diretório.
4. Busca o job:
   - `bus.Find(dir, layout.RoleBuilder, jobID)` — percorre pending/processing/done/failed.
   - Se `ErrNotFound` para o builder, também tenta `bus.Find(dir, layout.RoleReviewer, jobID)`.
   - Se não encontrado em nenhum papel → orphan: `Safe=true`, `Reason="job não encontrado no barramento (orphan)"` — usuário decide remover com aviso visual no prompt.
5. Se job encontrado em `StateProcessing` (qualquer papel) → `Safe=false`, `Reason="job em execução"`.
6. Se job encontrado em `StatePending` → `Safe=false`, `Reason="job aguardando execução"`.
7. Se job encontrado em `StateFailed` → `Safe=false`, `Reason="job falhou; pode ser retentado"`.
8. Se job encontrado em `StateDone`:
   a. Chama `queue.LoadForInitiative(dir, job.Job.Initiative)`.
   b. Busca task `taskID` nos `Items`. Se não encontrada → trata como **não terminal**:
      `Safe=false`, `Reason="task não encontrada na fila"` (fail closed).
   c. Se task encontrada com `StatusApproved` → `Safe=true`,
      `Reason="job concluído e task aprovada"`.
   d. Para qualquer outro status (`queued`, `building`, `review_pending`,
      `changes_requested`, `blocked`) → `Safe=false`, `Reason="task ainda na fila (<status>)"`.

**Reclassificação antes da remoção:** `SweepWorktrees` chama `ListStaleWorktrees` novamente
antes de remover cada item (workers do `forge up` anterior podem ter reclamado um job enquanto
o prompt estava aberto). Remove somente se ainda `Safe=true`.

### 3. Hooks nos comandos `forge up`, `forge worker` e `forge inbox`

Mesmo padrão nos três pontos de entrada:

| Arquivo | Momento de execução |
|---|---|
| `cmd/forge-sdd/up.go` | Antes de `env.LaunchAll(...)` |
| `cmd/forge-sdd/worker.go` | Antes de iniciar a sessão do worker |
| `cmd/forge-sdd/inbox.go` | No início de `runInbox`, antes de listar notificações |

```go
stale, _ := pipeline.ListStaleWorktrees(abs)
safe := filterSafe(stale) // inclui orphans com Safe=true
if len(safe) > 0 {
    chosen, err := survey.ChooseWorktreesToClean(safe)
    if err == nil && len(chosen) > 0 {
        for _, e := range pipeline.SweepWorktrees(abs, chosen) {
            fmt.Fprintln(out, "aviso:", e)
        }
    }
}
```

Não bloqueia o comando principal em caso de recusa ou erro.
Com `--dry-run` ou sem TTY: listar em stdout, sem promover nem remover.

### 4. `internal/survey/survey.go` — prompt de múltipla escolha

```go
// ChooseWorktreesToClean exibe a lista de worktrees seguros e retorna os que o usuário
// escolheu remover. Retorna slice vazio (sem erro) se o usuário escolher "Manter todos".
func ChooseWorktreesToClean(entries []pipeline.StaleWorktree) ([]pipeline.StaleWorktree, error)
```

Usa `huh.NewMultiSelect[string]()` já disponível em v0.3. Opções: uma por worktree seguro
(`<jobID>/<taskID> — job concluído e task aprovada`) mais "Manter todos" como ação de saída.

**Nota de dependência cruzada:** `internal/survey` importar `internal/pipeline` cria ciclo
se `internal/pipeline` importar `internal/survey`. Verificar importações antes de implementar;
se houver ciclo, mover `ChooseWorktreesToClean` para um novo `internal/pipeline/gc_prompt.go`
usando `huh` diretamente (já é dependência do módulo).

### 5. `forge pipeline worktree gc` — subcomando CLI

Novo `ValidArgs` `gc` no `pipelineWorktreeCmd` existente:

```
forge pipeline worktree gc [--dir <dir>] [--dry-run]
```

- **Com `--dry-run` ou sem TTY:** lista todos os stale em stdout no formato
  `SAFE|PROTECTED <jobID>/<taskID>  <reason>`, sem promover, sem remover. Exit 0.
- **Com TTY e sem `--dry-run`:** mesma interação do hook de `forge up`.
- **Lógica de remoção:** delega a `pipeline.SweepWorktrees` (não reimplemeneta em `cmd/`).

### 6. Comportamento em cenários especiais

| Cenário | Comportamento |
|---|---|
| `--dry-run` | Lista todos (safe e protegidos), sem remover |
| Sem TTY / autopilot | Apenas lista safe em stdout, sem prompt, sem remover |
| Worktree orphan (sem job) | `SAFE` com aviso no prompt: "job não encontrado — pode ser resíduo de branch excluída" |
| Job `failed` | `PROTECTED`, nunca oferecido |
| Task `review_pending` | `PROTECTED`; PR pode retornar como `pr_rework` |
| Task não encontrada na fila | `PROTECTED` (fail closed) |
| Erro em `RemoveWorktree` | Aviso no stderr; continua com os próximos |
| `.forge/runtime/worktrees` inexistente | Retorna vazio silenciosamente |
| `git worktree prune` falha | Aviso no stderr; não aborta |
| Workers ativos reclamam job entre prompt e remoção | `SweepWorktrees` reclassifica antes de cada remoção |

## Critérios de Aceitação Executáveis

```bash
go test -v -race ./internal/pipeline/... -run TestWorktreeGC
go test -v -race ./internal/queue/... -run TestLoadForInitiative
go test -v -race ./cmd/forge-sdd/... -run TestPipelineWorktreeGCCmd
go vet ./...
```

Cenários obrigatórios (cada um = um `Test*` independente):

1. **TestWorktreeGC_JobDoneTaskAprovada**: worktree presente, job `done` no barramento e task
   `StatusApproved` em queue.json via `LoadForInitiative` → `Safe=true`.
2. **TestWorktreeGC_JobProcessing**: job `StateProcessing` → `Safe=false`, reason contém
   "em execução".
3. **TestWorktreeGC_JobPending**: job `StatePending` → `Safe=false`, reason contém
   "aguardando".
4. **TestWorktreeGC_JobFailed**: job `StateFailed` → `Safe=false`, reason contém "falhou".
5. **TestWorktreeGC_TaskReviewPending**: job `done` mas task `review_pending` → `Safe=false`,
   reason contém "review_pending".
6. **TestWorktreeGC_TaskNaoEncontrada**: job `done` mas task ausente em queue.json →
   `Safe=false`, reason contém "não encontrada" (fail closed).
7. **TestWorktreeGC_Orphan**: nenhum job com esse ID encontrado no barramento → `Safe=false`,
   reason contém "não encontrado".
8. **TestWorktreeGC_SemWorktreeDir**: `.forge/runtime/worktrees` não existe → slice vazio,
   sem erro.
9. **TestWorktreeGC_ReclassificacaoAntesRemocao**: job `done`/`approved` ao listar, mas
   `processing` no momento da remoção → `SweepWorktrees` não remove.
10. **TestLoadForInitiative_SemFila**: queue.json ausente → fila vazia, sem erro.
11. **TestLoadForInitiative_IOError**: erro de leitura diferente de `ErrNotExist` → erro
    propagado (fail closed).
12. **TestPipelineWorktreeGCCmd_DryRun**: `forge pipeline worktree gc --dry-run` imprime lista
    e retorna exit 0 sem remover nada.
13. **TestPipelineWorktreeGCCmd_SemTTY**: sem TTY imprime lista e retorna exit 0 sem prompt.

Todos os testes passam em `go test -race`. `go vet ./...` limpo.
