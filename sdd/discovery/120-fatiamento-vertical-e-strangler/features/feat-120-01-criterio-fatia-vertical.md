# feat-120-01 — Critério de fatia vertical no split-features

Task 1 da Onda 1 (piloto, baixo risco). Branch própria `feat/120-01-criterio-fatia-vertical-<sid>`.

## Modelo do módulo
**Enxuto** (`ARQ-11`): só texto de skill e seu `.tmpl`; sem pacote, tipo nem fluxo de código.

## Decisões do projeto aplicadas
- `ARQ-03`: a mudança nasce no `.tmpl` embutido; a cópia em `.agents/` é espelho idêntico.
- `DUP-01`: o critério de fatiamento vive só aqui (passo 4); as outras skills apontam para ele.
- `COD-05`: texto em pt-BR.
- Gate 1 (DA-1..DA-4): sem Constituição/ADR nova; sem template de `plan.md`; contrato do Parecer inalterado.

## O que já existe
| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| critério de quebra | `split-features.md` passo 4 e `.tmpl` | "testável sozinha", "baixo para cima" | estender (mesmo passo) |
| golden | `internal/scaffold/testdata/golden`, `golden_test.go` | comparação com `-update` | reusar |

## Dependências novas
Nenhuma (`ARQ-04`).

## Arquivos que toca
- `.agents/skills/sdd-planejamento/references/split-features.md`
- `internal/scaffold/templates/.agents/skills/sdd-planejamento/references/split-features.md.tmpl`
- golden: `internal/scaffold/testdata/golden/.agents/skills/sdd-planejamento/references/split-features.md` (regenerado por `-update`)

Passo 4 passa a exigir **fatia vertical** que valida e entra na main sozinha; "baixo para cima" vale só dentro da fatia; manter "testável sozinha". Sem conflito com feat-120-02 e feat-120-03.

## Rules e IDs da Constituição Aplicáveis
`ARQ-03`, `DUP-01`, `COD-05`

## Critérios de Aceitação Executáveis
- `go test ./internal/scaffold/ -update` e, em seguida, `go test ./internal/scaffold/` (sem a flag) sai 0.
- `grep -q "fatia vertical" .agents/skills/sdd-planejamento/references/split-features.md` sai 0.
- `grep -q "fatia vertical" internal/scaffold/templates/.agents/skills/sdd-planejamento/references/split-features.md.tmpl` sai 0.
- `grep -q "testável sozinha" .agents/skills/sdd-planejamento/references/split-features.md` sai 0.
- `go test ./internal/harness/... && go build ./... && go vet ./...` sai 0.
- `! git diff --name-only origin/main | grep -E '^(sdd/memory/constitution\.md|sdd/spec/adr/)|\.go$' | grep -v '/testdata/'` sai 0 (não toca Constituição, ADR nem `*.go` fora de testdata).
