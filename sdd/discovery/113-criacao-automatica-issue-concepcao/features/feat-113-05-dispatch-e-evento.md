# Feature 113-05 — `Dispatch` aproveita a evidência e evento aditivo `spec.aberta`

## Objetivo
Fazer o `Dispatch` carregar `card-evidence.json` em `job.Card` quando o arquivo existir (mesmo com `require_existing_card: false`) e, havendo módulo ativo, emitir `spec.aberta` de forma aditiva.

## Modelo do módulo
**Enxuto** (ARQ-11): mudança localizada em `internal/pipeline/dispatch.go`; o evento é um gancho opcional. Sem pacote novo.

## Decisões do projeto aplicadas
- **ADR-0004** (existe): `job.Card` é o canal Discovery para Builder e Reviewer; o contrato do barramento não muda.
- **ADR-0002** (existe): a evidência fica na pasta da iniciativa.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Carga da evidência no `Dispatch` | `internal/pipeline/dispatch.go:200-212,240` | `loadEvidence` só roda se `cfg.PMMCP.RequireExistingCard`; `Card: bus.Card{Provider,ID,URL}` sai vazio senão. Plano confirmado | Estender | Com o arquivo presente, carregar sem bloquear; evidência inválida/vencida vira aviso quando `require_existing_card` é `false` |
| `loadEvidence` | `dispatch.go:431-497` | Retorna pendências (frescura, provider, feature) | Reusar | Com `require_existing_card: true` o comportamento atual NÃO muda |
| `internal/modulo` | `internal/modulo/**` | NÃO existe na `main` (a 109 não está mesclada). Divergência: o plano diz "quando presente" | Criar gancho aditivo | Sem o pacote, emitir o evento é no-op por interface/função injetável (nil = custo zero); a ligação real fica para quando a 109 mesclar |
| Testes de dispatch | `internal/pipeline/dispatch_test.go` | `TestDispatch*` cobrem evidência ausente/vencida/inválida | Reusar | Não podem regredir |

## Dependências novas
Nenhuma (SEG-07, ARQ-04).

## Rules e IDs da Constituição Aplicáveis
`ARQ-02`, `DUP-01`, `FLX-06`, `SEG-01`, `TST-03`, `TST-04`.

## Critérios de Aceitação Executáveis
Sem rede, sem `gh` e sem módulo real: o emissor é um fake injetado (TST-04).
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/pipeline -run 'TestDispatchCarregaCardEvidenceOpcionalQuandoPresente|TestDispatchSemEvidenciaComCardOpcionalNaoBloqueia|TestDispatchEvidenciaInvalidaComCardOpcionalViraAviso|TestDispatchRequireExistingCardContinuaBloqueando|TestEmiteSpecAbertaSemModuloEhNoOp|TestEmiteSpecAbertaComModuloLevaIssue'
go test ./internal/pipeline -run 'TestDispatch'
```
Casos: evidência válida preenche `job.Card` com `require_existing_card: false`; ausente não bloqueia; inválida ou vencida não bloqueia nesse modo (negativo) e `require_existing_card: true` segue bloqueando como hoje; evento sem módulo ativo faz zero chamadas; com módulo (fake) leva número, URL e título e não duplica o cartão.

## Depende de
113-03.
