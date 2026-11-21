# Critérios — Iniciativa 110 (régua de estimativa de tokens)

Legenda dos diagramas: **HUMANO** (dono), **IA** (agentes) e **AUTOMÁTICO** (código, sem LLM). Diagramas em `diagramas/` (diagram-studio, assinatura Forge SDD): `forge-110-dados-e-regua.png` e `forge-110-fluxo-discovery.png`.

## Restrições (Constitution)
- `ARQ-01`, `ARQ-02`, `ARQ-04` (só stdlib), `ARQ-11`; `SEG-01` (dado lido validado na borda), `SEG-02` (caminhos dentro do projeto), `SEG-05` (só contadores, nunca conteúdo de conversa); `DUP-01`, `DUP-02`, `DUP-04`; `FLX-01`, `FLX-05` (ADR do esquema do histórico); `TST-01`, `TST-03`, `TST-04` (sem rede), `COD-05`.

## Modelo do módulo
**Completo** (`ARQ-11`), no repositório do módulo: leitor da amostra, estatística, relatório e histórico de erro. Tem estado (histórico) e borda externa (arquivos de métricas e transcrição). Nada no core nesta discovery.

## Decisões do projeto aplicadas
- **109 (contrato de módulos):** o módulo é aditivo, sem alterar o core sem módulo (byte a byte), sem canal de volta; extensões do contrato só aditivas.
- **ADR do `ParseJobUsage`/`medicao` (97-01):** `real` e `estimado` nunca se misturam sem rótulo.
- **Memória do dono:** handoff é menu; PR abre em draft.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Esquema da sessão | `internal/monitor/session_record.go`, `sdd/.metrics/schema.json` | `SessionMetrics` com modelo, tokens, `medicao`, cache | **Reusar o esquema** (ler) | `DUP-01`: o core é dono do conceito |
| Soma por papel | `internal/monitor/tokens.go` (`TaskTokenMetrics`, `CalculateTokenMetrics`) | soma por Discovery, Builder, Reviewer | **Estender** no módulo com mediana e faixa | só soma, sem estatística |
| Tokens reais | `internal/monitor/usage.go` (`ParseJobUsage`, `SumUsage`, `EstimateSessionUsage`) | só o `result` do Claude; resto por bytes | **Reusar a técnica** no módulo | outras ferramentas: não verificado |
| Relatório | `cmd/forge-sdd/report.go` | tokens, modelos, sessões por item | **Não duplicar**; `estimar` olha o futuro | `DUP-02` |
| Modelo por papel | `agent_models` (104-05), `internal/config/forge_profiles.go` | modelo e perfis de CLI | **Reusar** como fonte de "ferramenta" | o pipeline já sabe |
| Bloco `spec` | `internal/monitor/spec_metrics.go` | só na branch do PR #131, aberto | **Dependência**, não existente | |
| Histórico | `sdd/archive/discovery/85-modelos-por-sessao-e-roteamento-economico`, `criteria-55-telemetria-cobertura-total-e-relatorio.md` | critérios de roteamento e telemetria | **Ler antes da onda 2** | evitar refazer |

## Dependências novas
Nenhuma dependência Go (`ARQ-04`). Depende da 109 (contrato; só onda 3) e do PR #131 (bloco `spec`; só onda 2).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-02`, `ARQ-04`, `ARQ-11`, `SEG-01`, `SEG-02`, `SEG-05`, `DUP-01`, `DUP-02`, `FLX-01`, `TST-01`, `TST-03`, `TST-04`, `COD-05`.

## Critérios de Aceitação Executáveis
Esta discovery é só especificação: os primeiros blocos conferem os artefatos; os testes `go test` são os nomes que as tasks devem criar no repositório do módulo (o portão confere a forma, não a execução).

```bash
set -o pipefail
# artefatos da discovery
test -s sdd/discovery/110-estimativa-de-tokens/discovery.md
test -s sdd/discovery/110-estimativa-de-tokens/plan.md
test "$(wc -c < sdd/discovery/110-estimativa-de-tokens/status.md)" -le 1024
test -s sdd/discovery/110-estimativa-de-tokens/diagramas/forge-110-dados-e-regua.png
test -s sdd/discovery/110-estimativa-de-tokens/diagramas/forge-110-fluxo-discovery.png
grep -F '## Descrição de negócio' sdd/discovery/110-estimativa-de-tokens/discovery.md
grep -F '## Decisões para o Gate 1' sdd/discovery/110-estimativa-de-tokens/discovery.md
# amostra e estatística (onda 1)
go test ./internal/amostra/... -run 'TestClassificaRealEstimadoEZero|TestIgnoraArquivoInvalido|TestSomenteContadoresNuncaConteudo'
go test ./internal/regua/... -run 'TestMedianaEFaixaP25P75|TestRotuloSemEstimativaAbaixoDeCinco|TestRotuloAmostraPequenaAteQuinze'
go test ./cmd/... -run 'TestRelatorioMostraNEDefinicaoDeToken|TestSemPrecosNaoMostraDinheiro|TestPrecosLidosDeArquivoComData'
# estimativa por feature e calibração (ondas 2 e 3)
go test ./internal/regua/... -run 'TestEstimativaPorFeatureESomaDaIniciativa|TestRecomendacaoSoComAmostraSuficiente'
go test ./internal/historico/... -run 'TestRegistraEstimadoVersusRealEErro|TestMetaDeErroMedianoComAmostraMinima'
# negativo: cada caso abaixo deve falhar, e o teste existe de fato
go test ./internal/regua/... -list 'TestSemAmostraNaoInventaNumero' | grep -F 'TestSemAmostraNaoInventaNumero'
go test ./internal/amostra/... -list 'TestNaoEnviaDadoParaRede' | grep -F 'TestNaoEnviaDadoParaRede'
go test ./internal/amostra/... -list 'TestNaoGravaEstadoDoCore' | grep -F 'TestNaoGravaEstadoDoCore'
go test ./cmd/... -list 'TestPrecoNuncaEmbutidoNoBinario' | grep -F 'TestPrecoNuncaEmbutidoNoBinario'
go test ./internal/regua/... -run 'TestSemAmostraNaoInventaNumero|TestNaoEnviaDadoParaRede|TestNaoGravaEstadoDoCore|TestPrecoNuncaEmbutidoNoBinario'
# negativo: a assinatura dos diagramas não leva nome pessoal nem LinkedIn
! grep -iE 'nathan|linkedin' sdd/discovery/110-estimativa-de-tokens/diagramas/forge-110-dados-e-regua.svg sdd/discovery/110-estimativa-de-tokens/diagramas/forge-110-fluxo-discovery.svg || exit 1
```
