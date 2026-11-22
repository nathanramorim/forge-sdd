# feat-106-06 — Cadeia de fallback do atestado (github e nenhum)

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 0
**Depende de:** feat-106-02, feat-106-05

## Problema
Task `06` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`internal/vcs`, `internal/pipeline` (pergunta ao humano).
**Colisões e ordem:** sem colisão com outra feature ativa além da ordem de `Depende de`.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `internal/vcs` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`FLX-01`, `SEG-05`, `TST-03`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/vcs/... -run 'TestFallbackTerminaNoLocal'
# negativo: TestFallbackNaoAprovaEmSilencio existe de fato e passa
go test ./internal/vcs/... -list 'TestFallbackNaoAprovaEmSilencio' | grep -F 'TestFallbackNaoAprovaEmSilencio'
go test ./internal/vcs/... -run 'TestFallbackNaoAprovaEmSilencio'
# negativo: TestFallbackNaoBloqueiaParaSempre existe de fato e passa
go test ./internal/vcs/... -list 'TestFallbackNaoBloqueiaParaSempre' | grep -F 'TestFallbackNaoBloqueiaParaSempre'
go test ./internal/vcs/... -run 'TestFallbackNaoBloqueiaParaSempre'
# negativo: TestNenhumRecusaModoVCS existe de fato e passa
go test ./internal/vcs/... -list 'TestNenhumRecusaModoVCS' | grep -F 'TestNenhumRecusaModoVCS'
go test ./internal/vcs/... -run 'TestNenhumRecusaModoVCS'
# negativo: TestModoVCSRecusadoComModoAdiado existe de fato e passa
go test ./internal/vcs/... -list 'TestModoVCSRecusadoComModoAdiado' | grep -F 'TestModoVCSRecusadoComModoAdiado'
go test ./internal/vcs/... -run 'TestModoVCSRecusadoComModoAdiado'
# negativo: TestAzureDevOpsRecusaNaoVerificado existe de fato e passa
go test ./internal/vcs/... -list 'TestAzureDevOpsRecusaNaoVerificado' | grep -F 'TestAzureDevOpsRecusaNaoVerificado'
go test ./internal/vcs/... -run 'TestAzureDevOpsRecusaNaoVerificado'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
- Nomes novos (não constam no `criteria.md`; o Builder os cria): `TestFallbackTerminaNoLocal`.
