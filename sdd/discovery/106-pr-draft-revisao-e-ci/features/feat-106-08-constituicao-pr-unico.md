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

## Itens que migram das Regras 11 e 12 para `sdd-fluxo-e-portao.md`
Conferido contra o texto de `origin/main` (Regras 11 e 12 e `FLX-07`). O `FLX-07` encurtado guarda só: beta não mescla em `main`, Release Notes reais e retenção de 3 estáveis. O texto pronto está em `proposta-0.8/sdd-fluxo-e-portao.proposta.md`. A rule hoje já cobre: tag na branch da feature, PR aberto até a promoção, destaques reais e prévias de comandos; não cobre a retenção.

| # | Item (texto original) | Origem | Entra na rule |
|---|---|---|---|
| 1 | "uma beta pode acumular fixes e features" | Regra 11 | sub-item novo, bloco Release Beta |
| 2 | "`-beta` ou `-rc`" (vale também para `-rc`) | Regra 12 | sub-item das versões, troca `-beta` por `-beta` e `-rc` |
| 3 | "o workflow falha com o texto genérico" | Regra 12 | sub-item novo |
| 4 | "as 3 releases estáveis mais recentes e as betas posteriores à mais antiga delas" | Regra 12 | sub-item de retenção (a rule não trazia a retenção) |
| 5 | "o NPM não é afetado" | Regra 12 | mesmo sub-item da retenção |
| 6 | "Novos comandos atualizam a documentação, com prévia e agente explícito" | Regra 12 | sub-item novo (a rule só tinha "prévias de comandos") |

Os critérios estão no bloco P-25 do `criteria.md`.

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
grep -F 'acumular fixes e features' .agents/rules/sdd-fluxo-e-portao.md
grep -F 'vale também para `-rc`' .agents/rules/sdd-fluxo-e-portao.md
grep -F 'texto genérico' .agents/rules/sdd-fluxo-e-portao.md
grep -F 'betas posteriores à mais antiga' .agents/rules/sdd-fluxo-e-portao.md
grep -F 'NPM não é afetado' .agents/rules/sdd-fluxo-e-portao.md
grep -F 'com prévia e agente explícito' .agents/rules/sdd-fluxo-e-portao.md
# negativo: a Constituição não pode mais citar as Regras 11 e 12 removidas
if grep -E 'Regras? 1[12]' sdd/memory/constitution.md; then exit 1; fi
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
