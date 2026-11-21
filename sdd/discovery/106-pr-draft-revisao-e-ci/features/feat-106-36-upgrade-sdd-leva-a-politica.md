# feat-106-36 — /upgrade-sdd leva a política; forge update e goldens; nota de versão

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 3
**Depende de:** feat-106-27

## Problema
Task `36` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`.claude/commands/upgrade-sdd.md`, `internal/scaffold` (goldens), `sdd/releases/`.
**Colisões e ordem:** Templates e goldens: depois da 26b.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `.claude/commands/upgrade-sdd.md` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`TST-02`, `FLX-07`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/scaffold/...
# negativo: TestUpgradeNaoSobrescreveSettingsDoUsuario existe de fato e passa
go test ./internal/scaffold/... -list 'TestUpgradeNaoSobrescreveSettingsDoUsuario' | grep -F 'TestUpgradeNaoSobrescreveSettingsDoUsuario'
go test ./internal/scaffold/... -run 'TestUpgradeNaoSobrescreveSettingsDoUsuario'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
- Nomes novos (não constam no `criteria.md`; o Builder os cria): `TestUpgradeNaoSobrescreveSettingsDoUsuario`.
