# Feature 113-04 — Subcomando `forge queue card`

## Objetivo
Expor a operação de vínculo da concepção: criar ou reaproveitar a issue (GitHub), registrar card vindo de MCP, ou sair em silêncio (desligado), com `--dry-run` e falha externa só como aviso.

## Modelo do módulo
**Enxuto** (ARQ-11): `cmd/forge-sdd/queue.go` só adapta flags (ARQ-01); a regra fica em `internal/` (113-01..03). Decisão do dono: `card` é SUBCOMANDO de `queue` (existente, `queue.go:486-494`), sem comando novo em ARQ-06 e sem ADR.

## Decisões do projeto aplicadas
- **ADR-0017** (existe): saída de `gh`/MCP é não confiável e saneada (princípio).
- **ADR-0004** (`adr-0004-pipeline-multi-sessao.md`, existe): o vínculo gravado alimenta `job.Card`. A task não depende de `forge up` nem de tmux: roda no chat (fix-89-01 já na `main`).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| `queueCmd` e subcomandos | `cmd/forge-sdd/queue.go:17,486-494` | `enqueue,next,complete,review,list,panel,handoff,stations,activity`; sem `card` | Estender | Novo `queueCardCmd` registrado com `queueCmd.AddCommand` |
| Padrão de `--dry-run` | `cmd/forge-sdd/dispatch.go:73`, `archive.go:104` | Flag booleana por comando | Reusar | ARQ-05 |
| Peças 113-01..03 | `internal/vcs`, `internal/config`, `internal/pipeline` | Não existem ainda | Reusar (dependência) | Esta task só orquestra |
| `forge queue card registrar` no discovery | `discovery.md` (113) | Texto fala em `card registrar`; o plano fala em `forge queue card` | Divergência | Adotar `forge queue card` com `--provider --id --url --feature` para o caso MCP; sem flags de registro, resolve o gestor |

## Dependências novas
Nenhuma (SEG-07, ARQ-04): cobra já é dependência do projeto.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-05`, `COD-03`, `SEG-01`, `SEG-08`, `TST-03`, `TST-04`.

## Critérios de Aceitação Executáveis
Testes com `Runner` falso e `t.TempDir()`; NUNCA chamam o `gh` real nem a rede (TST-04).
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./cmd/forge-sdd -run 'TestQueueCardDesligadoCustoZero|TestQueueCardGitHubCriaIssueEGravaEvidencia|TestQueueCardIdempotenteNaoDuplica|TestQueueCardRegistraRetornoMCP|TestQueueCardDryRunNaoCriaNemGrava|TestQueueCardFalhaExternaViraAvisoSemErroFatal|TestQueueCardRecusaCamposInvalidos'
go test ./internal/pipeline -run 'TestConcepcaoCustoZeroQuandoGestorNaoConfigurado|TestConcepcaoCriaIssueEGravaCardEvidence|TestConcepcaoFalhaExternaViraAvisoSemBloquearSpec'

# Os testes citados existem de verdade (sem isso o -run passa vazio):
test "$(go test ./cmd/forge-sdd -list 'TestQueueCardDesligadoCustoZero|TestQueueCardGitHubCriaIssueEGravaEvidencia|TestQueueCardIdempotenteNaoDuplica|TestQueueCardRegistraRetornoMCP|TestQueueCardDryRunNaoCriaNemGrava|TestQueueCardFalhaExternaViraAvisoSemErroFatal|TestQueueCardRecusaCamposInvalidos' | grep -c '^Test')" -ge 7   # pelo menos 7 (go test -run sem correspondência sai com 0)
test "$(go test ./internal/pipeline -list 'TestConcepcaoCustoZeroQuandoGestorNaoConfigurado|TestConcepcaoCriaIssueEGravaCardEvidence|TestConcepcaoFalhaExternaViraAvisoSemBloquearSpec' | grep -c '^Test')" -ge 3   # pelo menos 3 (go test -run sem correspondência sai com 0)
```
Casos: desligado não chama runner, não grava, não imprime erro; GitHub cria e grava `card-evidence.json`; segunda execução reaproveita; `--dry-run` não executa `gh issue create` nem toca o disco (ARQ-05); falha do `gh` sai com código 0 e aviso (RN-5); `--provider/--id/--url/--feature` inválidos (controle, tamanho, URL não http(s), provider fora de `supported`) recusados (SEG-01, negativo).

## Depende de
113-03.
