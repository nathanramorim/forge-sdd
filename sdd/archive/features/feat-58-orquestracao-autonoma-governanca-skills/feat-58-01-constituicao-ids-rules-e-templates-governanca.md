# feat-58-01 — Constituição por IDs Estáveis, Rules Modulares (`.agents/rules/`) e Templates de Governança

**Branch:** `feat/58-orquestracao-autonoma-governanca-skills`
**Fase:** 58-01
**Depende de:** —
**Status:** `done`

## Objetivo

Reestruturar `sdd/memory/constitution.md` para que seja curta, estável e indexada por **IDs canônicos** (`ARQ-*`, `SEG-*`, `DUP-*`, `COD-*`, `TST-*`, `FLX-*`, `GOV-*`), distribuindo o "como cumprir" em arquivos modulares `.agents/rules/sdd-*.md` (com frontmatter `description`, `globs`, `regras: [IDs]`) sem duplicar o texto da Constituição (`GOV-04`), além de incorporar ao scaffold os templates genéricos de governança (`concepts.md`, `adr-0000-modelo.md` e `.github/pull_request_template.md`).

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `enxuto` (templates estáticos em `embed.FS` e artefatos de memória/regras SDD).
- **Justificativa:** Define a estrutura de governança, IDs da Constituição e templates do scaffold sem introduzir nova lógica de estado em runtime.

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Constituição atual | `sdd/memory/constitution.md` e `internal/scaffold/templates/sdd/memory/constitution.md.tmpl` | 15 regras numeradas (0 a 15) | `estender` | Mapear as 15 regras existentes para IDs estáveis (`ARQ`, `SEG`, `DUP`, `COD`, `TST`, `FLX`, `GOV`) + tabela de Divergências Conhecidas e Mapa de Roteamento |
| Pasta `.agents/rules/` | `.agents/rules/README.md` (`feat-02-01`) | `README.md` e `example.md.example` | `estender` | Adicionar as 5 rules nativas do SDD (`sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md`) |
| Templates de Governança | `busize-solutions/planning-brain-governanca` (`templates/`) | `catalogo-conceitos.md`, `adr-modelo.md`, `pull_request_template.md` | `reusar` (genérico) | Incorporar a versão agnóstica de cliente no repositório e nos templates do scaffold (`sdd/spec/concepts.md`, `sdd/spec/adr/adr-0000-modelo.md`, `.github/pull_request_template.md`) |

## Dependências novas (`SEG-07`)
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`
- **IDs:** `ARQ-01..06`, `DUP-01`, `DUP-10`, `GOV-01`, `GOV-04`, Regra 3 (`embed.FS`).

## Critérios de Aceitação Executáveis

1. `sdd/memory/constitution.md` (e seu `.md.tmpl` em `internal/scaffold/templates/`) contém os IDs estáveis (`ARQ-*`, `SEG-*`, `DUP-*`, `COD-*`, `TST-*`, `FLX-*`, `GOV-*`) preservando integralmente as 15 regras atuais, mais o Mapa de Roteamento para `.agents/rules/` e a seção `Divergências conhecidas`.
2. Existem em `.agents/rules/` (e em `internal/scaffold/templates/.agents/rules/`) os 5 arquivos `sdd-{arquitetura,seguranca,reuso-e-duplicacao,codigo-e-testes,fluxo-e-portao}.md` com frontmatter `description`, `globs` e `regras: [...]` citando apenas os IDs da Constituição, sem copiar o texto das regras (`GOV-04`).
3. Existem no repositório e no scaffold (`internal/scaffold/templates/`) os modelos genéricos `sdd/spec/concepts.md`, `sdd/spec/adr/adr-0000-modelo.md` e `.github/pull_request_template.md` (com suporte ao bloco `constitution-exception`).
4. Golden tests em `internal/scaffold/testdata/golden/` atualizados; `go test ./internal/scaffold/...` e `go vet ./...` passam sem erros.

## Handoff
Pode rodar em paralelo com `feat-58-02`. Desbloqueia `feat-58-03` e `feat-58-04`.
