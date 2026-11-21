# feat-106-15 — Instalador do hook PreToolUse e aviso de cobertura no doctor

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 1
**Depende de:** feat-106-13, feat-106-14

## Problema
Task `15` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`internal/scaffold/hooks.go`, `cmd/forge-sdd` (`doctor.go`), templates e goldens.
**Colisões e ordem:** Ordem: depois da 13 (templates e goldens, se embutir); `hooks.go` e `doctor.go` só aqui nas ativas (0.4 é opcional); colide com a 105.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `internal/scaffold/hooks.go` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`SEG-02`, `ARQ-05`, `TST-02`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/scaffold/... -run 'TestMergePreToolUseHook'
# negativo: TestMergePreToolUseHookJSONInvalidoNaoETocado existe de fato e passa
go test ./internal/scaffold/... -list 'TestMergePreToolUseHookJSONInvalidoNaoETocado' | grep -F 'TestMergePreToolUseHookJSONInvalidoNaoETocado'
go test ./internal/scaffold/... -run 'TestMergePreToolUseHookJSONInvalidoNaoETocado'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
- Nomes novos (não constam no `criteria.md`; o Builder os cria): `TestMergePreToolUseHookJSONInvalidoNaoETocado`.
