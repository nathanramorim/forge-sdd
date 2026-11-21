# feat-105-02 — Corrigir texto: "5 estações" vira "3 sessões" e versões

**Iniciativa:** 105-experiencia-cli-e-harness (onda 1, task 1.2)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** feat-105-01
**Paralelo com:** feat-105-03 e feat-105-04 (arquivos disjuntos, exceto `doctor` na 04: não tocar `doctor.go` aqui)
**Sequenciar com:** 106 (tasks 1.7, 2.5 e 2.6 mexem em templates, goldens e `split-features.md`) e 107: esta task entra EM SÉRIE com elas e regrava os goldens, então rebase obrigatório e regeração dos goldens sobre o que já estiver na `main`; não editar arquivos da 106 nem da 107. Também antes da task 2.4 da 105.

## Problema
Arquivos entregues no `init` e a doc ainda dizem "5 estações" e carregam versão antiga (O1, O8, O9, O11).

## Comportamento
1. Trocar "5 estações" por "3 sessões (Discovery, Builder, Reviewer)" em: `internal/scaffold/templates/agents/claude/CLAUDE.md.tmpl`, `.../gemini/GEMINI.md.tmpl`, `.../openai/AGENTS.md.tmpl`, `.github/copilot-instructions.md.tmpl`, as descrições de `lider.md.tmpl` e `orquestrador.chatmode.md.tmpl` (só texto), `docs/acp-ide-setup.md:3`, `.agents/rules/sdd-acp.md:2`, `Long` de `cmd/forge-sdd/attach.go` e `Long`/`Short` do comando raiz (`main.go`, também o texto "pronta para uso com GitHub Copilot").
2. Cabeçalhos de versão de `docs/pipeline.md` e `sdd/spec/flows.md` apontam para a versão atual do binário.
3. Docs que citam `forge review` passam a citar `forge bus review`.
4. Regravar goldens em `internal/scaffold/testdata/golden/` no mesmo PR (`TST-02`).
5. Não alterar `lider` além da descrição (a task 1.6 cuida do texto interno), nem nomes de arquivo, comandos ou flags.

## Modelo do módulo
**Sem módulo novo.** Só texto de template, golden e `Long` de comando (`ARQ-01`).

## Decisões do projeto aplicadas
ADR-0004 (3 sessões), ADR da feat-105-01 (obsolescência), ADR-0003 (parcialmente obsoleta, não editada aqui).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Texto "5 estações" gerado | `internal/scaffold/templates/**`, `testdata/golden/**` | CLAUDE, GEMINI, AGENTS, copilot-instructions, lider, orquestrador (busca) | **Corrigir** | O1 do `obsoletos.md` |
| Texto fora de template | `docs/acp-ide-setup.md`, `.agents/rules/sdd-acp.md`, `cmd/forge-sdd/attach.go` | "5 estações" | **Corrigir** | O1 |
| Texto legado | `internal/scaffold/legacy/roles/lider.md.tmpl` | "5 estações" | **Decidir no Builder** (se ainda é gravado no `init`; se morto, `COD-04`) | Não verificado aqui |
| Teste de golden | `internal/scaffold/*_test.go` | Comparação com `testdata/golden` | **Reusar** | `TST-02` |
| Docs históricas | `sdd/archive/**`, `docs/showcase-v3-lp.html`, `brag-output/` | "5 estações" histórico/LP | **Manter** | Fora do escopo; arquivo histórico |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`COD-05`, `TST-02`, `DUP-06`. Rules: `sdd-codigo`, `sdd-testes`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/scaffold/... ./cmd/forge-sdd/...
go test -list 'TestTemplatesSemCincoEstacoes' ./internal/scaffold | grep -F 'TestTemplatesSemCincoEstacoes'
go test ./internal/scaffold -run 'TestTemplatesSemCincoEstacoes'
# negativo: nenhum template nem golden entregue ao usuário volta a dizer "5 estações"
go test -list 'TestGoldensSemTextoDeCincoEstacoes' ./internal/scaffold | grep -F 'TestGoldensSemTextoDeCincoEstacoes'
go test ./internal/scaffold -run 'TestGoldensSemTextoDeCincoEstacoes'
# negativo: texto de `Long` do raiz e do attach sem "5 estações"
go test -list 'TestLongSemCincoEstacoes' ./cmd/forge-sdd | grep -F 'TestLongSemCincoEstacoes'
go test ./cmd/forge-sdd -run 'TestLongSemCincoEstacoes'
go test ./...
```
- `TestTemplatesSemCincoEstacoes` e `TestGoldensSemTextoDeCincoEstacoes` varrem `templates/` e `testdata/golden/` (texto, não arquivo histórico).
- `--dry-run` do `init` e do `update` continua sem escrever (`ARQ-05`); revisar o `update` ao regravar templates corrigidos.
