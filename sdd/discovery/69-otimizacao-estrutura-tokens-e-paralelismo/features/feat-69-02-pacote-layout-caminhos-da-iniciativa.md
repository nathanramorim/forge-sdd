# feat-69-02 — Pacote `internal/layout`: Dono dos Caminhos da Iniciativa

**Branch:** `feat/69-otimizacao-estrutura-tokens-e-paralelismo`
**Fase:** 69-02
**Tamanho:** `M`
**Depende de:** —
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
O caminho `sdd/...` é montado com `filepath.Join(dir, "sdd", ...)` em 15 arquivos não-teste. Para mover estado para a pasta da iniciativa sem caçar literais, um único pacote passa a resolver todo caminho do layout e a iniciativa ativa a partir da branch.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** pacote de funções puras sobre `path/filepath`, mais uma chamada a `git` com argv explícito; sem estado nem dependência externa (`ARQ-11`, `SEG-03`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Caminhos `sdd/` literais | `cmd/forge-sdd/{lessons,destroy,autopilot,session,governance,doctor}.go`, `internal/{config,queue,gate,skills,scaffold}` | 15 arquivos montam o caminho à mão | criar `internal/layout` e migrar | Definição única de nomes canônicos (`DUP-04`) |
| Branch atual | `internal/gate/gate.go` (P-01), `cmd/forge-sdd/gate.go` | O gate não executa git: recebe a branch pela flag `--branch` | criar | `layout.CurrentBranch` é o primeiro executor de git do repositório; argv explícito, sem shell (`SEG-03`) |
| Validação de caminho | `internal/config` (`shouldPreserve`), `acp.ValidateWorkspace` | Confinamento de diretório | reusar o padrão | Resolução da iniciativa rejeita `..` e caminho absoluto (`SEG-01`) |
| Convenção de branch | `sdd/memory/naming-convention.md` | `<prefixo>/<id>-<nome>` | reusar | Mapeia direto para `sdd/discovery/<id>-<nome>/` |
| Constantes de nomes | `internal/queue`, `internal/config` | `.queue`, `.metrics`, `.sddrc` repetidos | estender | Pacote `layout` exporta as constantes |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-02`, `DUP-01`, `DUP-04`, `SEG-01`, `SEG-03`, `COD-01`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. `internal/layout`: `Root`, `InitiativeDir(root, id)`, `FromBranch(branch)`, `Current(root)` (lê a branch), `Status`, `Queue`, `Runtime`, `Archive`, `Features`, `Fixes`, `Memory`, `Metrics`.
2. Migrar os 15 arquivos para o pacote, sem mudar comportamento nesta task (o layout antigo continua resolvendo igual).
3. Branch fora do padrão (`main`, `release/*`) resolve para "sem iniciativa ativa", sem erro fatal.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/layout/...
! grep -rnE 'filepath\.Join\([^)]*"sdd"' --include='*.go' cmd internal | grep -v _test.go | grep -v internal/layout
go test ./...
```
- Casos conformes: `feat/69-x`, `fix/12-y`, `chore/...` sem iniciativa.
- Casos de violação (`TST-03`): `..`, caminho absoluto, id vazio, caracteres de controle.
- Repositório git em `t.TempDir()` para `Current` (`TST-04`); sem rede.
- Saída de `forge-sdd doctor` e `queue` idêntica à anterior (golden inalterado).
