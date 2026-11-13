# feat-69-04 — Subcomandos `status` e `next-id`

**Branch:** `feat/69-otimizacao-estrutura-tokens-e-paralelismo`
**Fase:** 69-04
**Tamanho:** `M`
**Depende de:** feat-69-02
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Sem `index.md` e `progress.md` globais, a visão geral precisa ser calculada, e o próximo ID precisa enxergar outras branches para duas frentes paralelas não escolherem o mesmo número. A Constitution (`ARQ-06`) já lista `status`, mas o comando não está registrado no CLI.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** dois subcomandos finos em `cmd/` sobre funções em `internal/layout`, com leitura de refs via `git` com argv explícito (`ARQ-01`, `ARQ-10`, `SEG-03`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Subcomando `status` | `cmd/forge-sdd/*.go`; `ARQ-06` | Listado em `ARQ-06`, não registrado; existe `queue panel` | criar `status` | Substitui `index.md` e `progress.md` na leitura |
| Próximo ID | `sdd/memory/naming-convention.md` | Regra em prosa, olha só a árvore local; modos `sequencial`, `hash`, `workitem` | criar `next-id` | Implementa os três modos consultando todas as refs |
| Leitura de refs | `internal/gate` | `git` com argv explícito | reusar o padrão | `git for-each-ref`, `git show <ref>:<path>` (`SEG-03`) |
| Caminhos e iniciativa | `internal/layout` (69-02) | `FromBranch`, `Status` | reusar | Única fonte |
| Procedimento | `.agents/skills/novo-subcomando-cli-sdd` | Aprovado em `sdd/skills/index.md` | reusar | `DUP-05` |
| Config de naming | `internal/config` (`naming_convention`) | Lê `.sddrc` | reusar | Define o modo de `next-id` |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-06`, `ARQ-10`, `SEG-01`, `SEG-03`, `COD-01`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Skill: `novo-subcomando-cli-sdd`.
- Gerais: `FLX-01`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. `forge-sdd status [--all-branches]`: tabela iniciativa, branch, status e próximo passo; com `--all-branches`, lê `status.md` de cada ref sem fazer checkout.
2. `forge-sdd next-id [--kind discovery|fix]`: devolve o próximo ID considerando pastas locais e de todas as refs; respeita `naming_convention` (`sequencial`, `hash`, `workitem`).
3. Registrar ambos em `main.go` e atualizar `ARQ-06` apenas na ADR 0003 (69-07), sem editar a Constitution aqui.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./cmd/forge-sdd/... ./internal/layout/...
forge-sdd status --help && forge-sdd next-id --help
```
- Repositório git temporário (`TST-04`): pasta `70-x` existente só em outra branch faz `next-id` devolver `71` (ou superior).
- `status --all-branches` lista iniciativas de branches não mescladas.
- Violações (`TST-03`): ref com nome malicioso (`--upload-pack=...`, `;`), repositório sem commits e branch destacada falham com erro enriquecido, sem panic.
- `--dry-run` não existe nestes comandos porque são somente leitura; nenhum escreve em disco.
