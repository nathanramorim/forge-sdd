# feat-106-32 — Verificação do atestado por PR de feature (modo local)

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** feat/106-pr-draft-revisao-e-ci (por feature só após a `constitution-exception` do `plan.md`, `GOV-02`)
**Status:** queued (bloqueada pela ordem)
**Onda:** 3
**Depende de:** feat-106-05, feat-106-06, feat-106-11, feat-106-21

## Problema
Task `32` do `plan.md` da 106 (escopo reduzido aprovado em 2026-11-20). Descrição completa e critérios da área: `plan.md` (tabela da onda) e `criteria.md`; esta spec não repete o texto.

## Modelo do módulo
**Completo** (`ARQ-11`), como na 106: `internal/vcs` com dependência para dentro e sem importar `cmd/`; esta feature mexe só no que consta em "Arquivos que toca".

## Arquivos que toca
`scripts/atestado.py`, workflow (job ubuntu).
**Colisões e ordem:** `persistVerdict`: só leitura aqui, depois da 21.

## Decisões do projeto aplicadas
ADR-0004, ADR-0012, ADR-0022 (PR único da Constituição, task 0.8). Decisões da 106 aprovadas (xiv a xxii) conforme `plan.md`. `GOV-02`: Regra 15 em exceção do dono, bloco `constitution-exception` no `plan.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Código e testes da área | `scripts/atestado.py` | Ver tabela `O que já existe` do `criteria.md` | **Reusar** e estender | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
`SEG-04`, `SEG-05`, `FLX-06`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
uv run scripts/atestado.py --verificar --sha HEAD --modo local
uv run scripts/atestado.py --verificar --sha HEAD --modo desligado
go test ./internal/vcs/... -run 'TestAtestadoDesligadoNaoExigeProva|TestAtestadoLocalConfereVeredito|TestModoDeAtestadoNoStatus|TestPRDaSpecSempreAtestadoLocal'
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/atestado.py --verificar --sha 0000000 --modo local; then exit 1; fi
# negativo: o caso abaixo deve falhar (saída 1 se passar)
if uv run scripts/atestado.py --verificar --sha HEAD --modo inexistente; then exit 1; fi
# negativo: o caso abaixo deve falhar (saída 1 se passar)
out=$(uv run scripts/atestado.py --verificar --sha HEAD --modo vcs 2>&1) && exit 1; printf '%s' "$out" | grep -F 'modo adiado'
# negativo: TestPRDaSpecIgnoraModoVCS existe de fato e passa
go test ./internal/vcs/... -list 'TestPRDaSpecIgnoraModoVCS' | grep -F 'TestPRDaSpecIgnoraModoVCS'
go test ./internal/vcs/... -run 'TestPRDaSpecIgnoraModoVCS'
go vet ./... && go test ./...
```
- Teste negativo: cada linha `# negativo:` acima deve provar a recusa; sem ele a feature não passa (`TST-03`).
