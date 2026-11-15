# feat-69-03 — Fila e Status por Iniciativa

**Branch:** `feat/69-otimizacao-estrutura-tokens-e-paralelismo`
**Fase:** 69-03
**Tamanho:** `M`
**Depende de:** feat-69-02
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
`progress.md` (154 commits), `index.md` (131) e `sdd/.queue/*.json` (96 KB) são arquivos únicos tocados por toda branch, a causa dos conflitos entre frentes paralelas. Esta task move o estado versionado para dentro da pasta da iniciativa e o estado de execução para fora do git.

## Modelo do módulo
- **Modelo:** `enxuto`, extensão de `internal/queue` e `internal/gate`.
- **Justificativa:** só troca a origem do caminho (via `internal/layout`) e adiciona um arquivo de estado por iniciativa; sem pacote novo (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Fila do Builder | `internal/queue/queue.go:77`, `Load`, `Save`, `Enqueue` | Caminho fixo `sdd/.queue/queue.json`; `Enqueue` sobrescreve `feature_id` e `branch` e acrescenta itens | estender | Recebe o caminho da iniciativa; mantém escrita atômica |
| Filas por estação e atividade | `internal/queue/stations.go`, `activity.go` | Estado de execução versionado em `sdd/.queue/` | estender | Mudam para `sdd/.runtime/` e entram no `.gitignore` |
| Orçamento de progresso | `internal/gate/gate.go` (P-21, linha 485) | Valida `sdd/memory/progress.md` ≤ 1.024 bytes | estender | Passa a validar o `status.md` da iniciativa ativa |
| Estado global | `sdd/memory/progress.md`, `progress-log.md`, `sdd/features/index.md` | Arquivos únicos | remover do fluxo | Substituídos por `status.md` local (leitura agregada em 69-04) |
| Telemetria | `sdd/.metrics/` | Um arquivo por sessão, sem colisão | reusar | Já é seguro para paralelismo |
| Caminhos | `internal/layout` (69-02) | `Status`, `Queue`, `Runtime` | reusar | Única fonte (`DUP-04`) |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `COD-01`, `COD-03`, `DUP-04`, `FLX-02`, `FLX-04`, `FLX-06`, `SEG-02`, `TST-03`, `TST-04`.
- Rules: `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md`.
- Gerais: `FLX-01`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. `queue.json` e `status.md` ficam em `sdd/discovery/<id>-<nome>/`; `Load`/`Save`/`Enqueue` resolvem pela iniciativa ativa.
2. `sdd/.runtime/` recebe atividade e filas por estação; o `.gitignore` deste repositório o ignora e o scaffold de projetos novos cria `sdd/.runtime/.gitignore` (entregue na 69-05).
3. `status.md`: no máximo 1.024 bytes, com seções `Status`, `Próximo passo` e `A confirmar`. P-21 valida este arquivo; sem iniciativa ativa, P-21 passa com observação.
4. Remover `sdd/memory/progress.md` e `sdd/features/index.md` do fluxo escrito pelos comandos; a visão geral vem de 69-04. Adiado: os comandos viram skills na 69-06 e os arquivos são arquivados na 69-05.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/queue/... ./internal/gate/... ./internal/layout/...
git ls-files sdd/.queue sdd/memory/progress.md sdd/features/index.md | wc -l   # 0 ao final da série
grep -q '^sdd/.runtime/' .gitignore
```
- Teste em repositório git temporário: duas branches criam iniciativas, enfileiram e atualizam `status.md`; o merge das duas termina sem conflito (`TST-04`).
- P-21: `status.md` com 1.024 bytes passa; com 1.025 falha (`TST-03`).
- `Enqueue` em iniciativa nova não herda itens de outra.
