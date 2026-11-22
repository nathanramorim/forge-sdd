# feat-106-27 — ADR final dos fluxos e concepts.md

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 2
**Depende de:** feat-106-26b, feat-106-08

## Problema
Task `27` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`sdd/spec/adr/` (próxima livre), `sdd/spec/concepts.md`, textos que citam os modos (apontam para a rule).
**Colisões e ordem:** Não edita a Constituição.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `sdd/spec/adr/` (próxima livre) | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`GOV-01`, `FLX-05`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
grep -F 'internal/vcs' sdd/spec/concepts.md
go test ./internal/gate/... -run 'TestConstitutionDoRepositorioMantemTodosOsIDs'
# negativo: o caso abaixo deve falhar (saída 1 se passar)
test "$(wc -c < sdd/memory/constitution.md)" -le 9216
# negativo: TestCheckP23RulesGovernance existe de fato e passa
go test ./internal/gate/... -list 'TestCheckP23RulesGovernance' | grep -F 'TestCheckP23RulesGovernance'
go test ./internal/gate/... -run 'TestCheckP23RulesGovernance'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
