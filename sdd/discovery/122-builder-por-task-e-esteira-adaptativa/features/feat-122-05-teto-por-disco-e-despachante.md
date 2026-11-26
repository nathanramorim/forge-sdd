# feat-122-05 — Teto de paralelo por disco, onda e janela

Tasks 2.1 e 2.2. Trilha A. Depende de feat-122-02.

## Modelo do módulo
**Completo** (`ARQ-11`). Módulos: internal/pipeline.

## Decisões do projeto aplicadas
RN-5, RN-6, RN-7, RN-10 da spec; ARQ-04.

## O que já existe
Nenhuma medição de disco hoje: criar função só com stdlib (`syscall.Statfs` em unix, `GetDiskFreeSpaceExW` via `syscall` no Windows, com build tags). Tamanho do worktree medido no primeiro checkout e guardado em `.forge/runtime/` por máquina, remedido a cada 20 tasks. Despachante: `pares = min(vagas de disco, largura da onda por PlanWaves+PodeComecar, teto da janela)`, recalculado a cada vaga liberada; `max_parallel_tasks` vira teto máximo.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-04`, `ARQ-02`, `DUP-01`, `TST-04`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/pipeline/ -run "Teto|Liberacao|Disco"
GOOS=windows go build ./... && GOOS=darwin go build ./... && go build ./... && go vet ./...
```
