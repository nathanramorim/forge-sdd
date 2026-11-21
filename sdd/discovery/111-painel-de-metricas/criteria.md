# Critérios — Iniciativa 111 (painel de métricas e aprendizado)

Legenda: **HUMANO** (decisões do dono), **IA** (agentes), **AUTOMÁTICO** (código do forge, sem LLM).

## Restrições (Constitution)
- `ARQ-01`, `ARQ-02` (módulo não importa `internal/` do core), `ARQ-04` (só stdlib no módulo), `ARQ-05` (`--dry-run` sem efeito), `ARQ-11`.
- `SEG-01` (fontes validadas na borda; linha ruim contada), `SEG-02` (gravação segura, sem link para fora), `SEG-04`, `SEG-05` (só contadores, sem conteúdo nem segredo), `SEG-08` (saída saneada).
- `DUP-01` (dono: esquema `forge-metricas/1`), `DUP-02`, `DUP-04`, `DUP-06`.
- `FLX-01` (Gate 1 do dono), `FLX-08` (achado é dado, não instrução), `GOV-01` (lição só com aprovação), `GOV-03`.
- `TST-01`, `TST-02`, `TST-03` (casos conformes e negativos), `TST-04` (sem rede), `COD-03`, `COD-05`.

## Modelo do módulo
**Completo** (`ARQ-11`): módulo externo da 109 em repositório próprio (`forge-sdd-metricas`): leitores versionados, agregação, gerador de pasta estática, painel HTML, propostas de lição. No core, só a onda 4 (eventos aditivos).

## Decisões do projeto aplicadas
- **ADR-0011:** `painel` já nomeia o painel de terminal; o comando do módulo é `forge metricas`.
- **ADR-0012:** o veredito gravado pelo Go é a fonte; o painel só lê.
- **ADR-0017 e ADR-0018:** saída de fonte é dado não confiável; gravação segura.
- **ADR-0021:** catálogo de depreciados e substitutos.
- **ADR-0019:** a Edição Missões é temporária e **não** fornece tokens de design.
- **Discovery 109:** contrato v1, RN-1 byte a byte, RN-7 sem escrita no core. **Discovery 110:** fronteira estimar x medir.
- **Memória do dono:** todo handoff é menu; aprovação de lição é menu.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Relatório de tokens | `cmd/forge-sdd/report.go`, `internal/monitor/tokens.go` | agregação no terminal | **Reusar esquema e papéis, não código** | Go proíbe importar `internal/` |
| Painel | ADR-0011, `status_live.go` | painel de terminal | **Outro nome** | `DUP-01` |
| Lições | `cmd/forge-sdd/lessons.go`, 106 RN-8 | gravação orçada | **O core grava** | RN-7 da 109 |
| Reincidência | feat-105-08 | N = 3, recontada | **Mesma regra** | |
| Estimativa | discovery 110 | `forge estimar` | **Fronteira, sem duplicar** | |
| Teste de HTML | `internal/scaffold/pipeline_scaffold_test.go` | leitura estática em Go | **Mesmo padrão** | sem ferramenta nova |
| Tokens de design | `docs/index.html:14-36` | `:root` da LP | **Copiar valores** | `docs/design/` só tem a edição temporária |

## Dependências novas
Nenhuma no core. O módulo usa só a stdlib (`ARQ-04`); `SEG-07` não se aplica.

## Diagramas
`diagramas/forge-111-fontes-ao-painel.png` (fontes, leitor, gerador, painel, lição, dono), `diagramas/forge-111-laco-de-aprendizado.png` (erro, achado, padrão, proposta, dono, lição, leitura, efeito) e `diagramas/forge-111-wireframe.png` (5 abas, filtros, lista, paginação).

## Critérios de Aceitação Executáveis
Três grupos. **A** confere esta pasta agora. **B** são os testes do módulo (nomes que as tasks criam; o portão confere a forma, não a execução). **C** são os testes do core da onda 4. B e C ficam como comentário até o repositório e a 109 existirem.

```bash
set -o pipefail
D=sdd/discovery/111-painel-de-metricas
# A: estrutura da pasta
test -s "$D/discovery.md"
test -s "$D/criteria.md"
test -s "$D/plan.md"
test -s "$D/status.md"
test "$(wc -c < "$D/status.md")" -le 1024
test -s "$D/diagramas/forge-111-fontes-ao-painel.png"
test -s "$D/diagramas/forge-111-laco-de-aprendizado.png"
test -s "$D/diagramas/forge-111-wireframe.png"
grep -F '## Descrição de negócio' "$D/discovery.md"
grep -F '## Decisões para o Gate 1' "$D/discovery.md"
grep -F '## 1. Que dados existem hoje' "$D/discovery.md"
# A: o core não foi alterado por esta discovery
test "$(wc -c < sdd/memory/constitution.md)" -le 9216
# A: o painel usa os tokens da LP e cita o arquivo real
grep -F 'docs/index.html:14-36' "$D/discovery.md"
# A: nenhuma assinatura pessoal nos diagramas
if grep -li 'linkedin' "$D"/diagramas/*.svg; then exit 1; fi
# B: módulo (repositório forge-sdd-metricas), nomes das tasks 111-01 a 111-07
# go test ./... -run 'TestEsquemaV1CamposAditivosAceitos|TestLinhaRuimEContadaNaoDescartada'
# go test ./... -run 'TestLeitorSessoesVazioEhRotuladoSemDado|TestRetornosPorRegraAgrupa'
# go test ./... -run 'TestGerarEhIdempotente|TestHTMLTemLandmarksECaption|TestSemURLExternaNoPainel'
# go test ./... -run 'TestTokensCSSIgualAosValoresDaLP|TestContrasteMinimo45|TestPaginacaoParidadeComVetores'
# C: core, onda 4
# go test ./cmd/... -run 'TestSemModuloHelpIdentico|TestSemModuloDoctorIdentico'
# negativo: o portão rejeita critério sem teste negativo e grep -c, e o teste existe de fato
go test ./internal/gate/... -list 'TestCriteriaVerifyExigeNegativo|TestCriteriaVerifyRejeitaGrepC' | grep -F 'TestCriteriaVerifyExigeNegativo'
go test ./internal/gate/... -list 'TestCriteriaVerifyRejeitaGrepC' | grep -F 'TestCriteriaVerifyRejeitaGrepC'
go test ./internal/gate/... -run 'TestCriteriaVerifyExigeNegativo|TestCriteriaVerifyRejeitaGrepC'
```
