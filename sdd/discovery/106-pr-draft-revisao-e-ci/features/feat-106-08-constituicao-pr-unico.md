# feat-106-08 — PR único da Constituição e da rule de modos

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** P
**Depende de:** nenhuma

## Problema
Task `08` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`sdd/memory/constitution.md`, `.agents/rules/sdd-modos-de-projeto.md` (nova), `.agents/rules/sdd-fluxo-e-portao.md`, `sdd/spec/adr/adr-0022*`.
**Colisões e ordem:** Única task que edita a Constituição. Sem dependência; abre em paralelo.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `sdd/memory/constitution.md` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`GOV-01`, `GOV-04`, `FLX-05`, `TST-03`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/gate/... -run 'TestConstitutionDoRepositorioMantemTodosOsIDs|TestCheckP24_RepositorioDoForgeSddPassaNoProprioPortao|TestCheckP23RulesGovernance'
grep -F '| Modo do atestado | local |' sdd/memory/constitution.md
grep -F '| Modo de revisão humana | com-humano |' sdd/memory/constitution.md
grep -F '| WIP humano | 3 |' sdd/memory/constitution.md
grep -F '| Módulos permitidos | nenhum |' sdd/memory/constitution.md
grep -F '| Início automático do Builder | desligado |' sdd/memory/constitution.md
grep -F '.agents/rules/sdd-modos-de-projeto.md' sdd/memory/constitution.md
grep -F 'regras:' .agents/rules/sdd-modos-de-projeto.md
ls sdd/spec/adr | grep -F 'adr-0022'
# negativo: o caso abaixo deve falhar (saída 1 se passar)
test "$(wc -c < sdd/memory/constitution.md)" -le 9216
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if grep -F 'Quadro de acompanhamento' sdd/memory/constitution.md; then exit 1; fi
# negativo: TestCheckP24_ReprovaArquivosGeradosAcimaDoLimite existe de fato e passa
go test ./internal/gate/... -list 'TestCheckP24_ReprovaArquivosGeradosAcimaDoLimite' | grep -F 'TestCheckP24_ReprovaArquivosGeradosAcimaDoLimite'
go test ./internal/gate/... -run 'TestCheckP24_ReprovaArquivosGeradosAcimaDoLimite'
# negativo: TestCheckP23RulesGovernance existe de fato e passa
go test ./internal/gate/... -list 'TestCheckP23RulesGovernance' | grep -F 'TestCheckP23RulesGovernance'
go test ./internal/gate/... -run 'TestCheckP23RulesGovernance'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
