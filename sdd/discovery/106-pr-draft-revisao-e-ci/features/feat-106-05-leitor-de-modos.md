# feat-106-05 — Leitor único dos campos de modo da Constituição

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 0
**Depende de:** feat-106-02, feat-106-08

## Problema
Task `05` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`internal/vcs` (leitor), `internal/config`.
**Colisões e ordem:** sem colisão com outra feature ativa além da ordem de `Depende de`.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `internal/vcs` (leitor) | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-10`, `SEG-01`, `GOV-01`, `TST-03`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/vcs/... -run 'TestLeModosDaConstituicaoDaBranchBase|TestAtestadoPadraoELocal'
# negativo: TestModoDeAtestadoInvalidoRecusado existe de fato e passa
go test ./internal/vcs/... -list 'TestModoDeAtestadoInvalidoRecusado' | grep -F 'TestModoDeAtestadoInvalidoRecusado'
go test ./internal/vcs/... -run 'TestModoDeAtestadoInvalidoRecusado'
# negativo: TestModoNaBranchDaFeatureNaoSobrepoeABase existe de fato e passa
go test ./internal/vcs/... -list 'TestModoNaBranchDaFeatureNaoSobrepoeABase' | grep -F 'TestModoNaBranchDaFeatureNaoSobrepoeABase'
go test ./internal/vcs/... -run 'TestModoNaBranchDaFeatureNaoSobrepoeABase'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
