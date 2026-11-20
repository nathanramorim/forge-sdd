# fix-97-01 — A sessão de Discovery grava a telemetria genérica, sem bloco próprio de métricas da spec

**Branch:** `fix/97-metricas-de-discovery`
**Origem:** Iniciativa 97 (mesma superfície do `fix-90-02`: campos do `session-*.json`); causa-raiz histórica nas iniciativas 01 e 55, ambas arquivadas
**Tamanho:** `M`
**Status:** `spec`

## Descrição de negócio (release notes)

*   **Métricas de Discovery no mesmo arquivo de sessão:** as sessões de `/discovery`, `/nova-feature` e `/novo-fix` passam a gravar, em `sdd/.metrics/session-*.json`, o bloco `spec` com o tipo e o tamanho do que foi especificado; `/split-features`, Build, Review e arquivos antigos seguem sem o bloco (`fix-97-01`).

## Problema

O pedido do dono: o Build gera métricas e o Discovery não; o Discovery deveria gerar dentro do mesmo arquivo, numa propriedade nova `spec: {...}`.

O diagnóstico com evidência corrige a premissa em parte. O Discovery já grava o arquivo (`discovery.md`, passo 10, manda rodar `forge session record`; há exemplo real em `origin/feat/107-orquestrador-de-sessoes:sdd/.metrics/session-2026-11-19T155818Z.json`, com `feature: sdd/discovery/107-.../discovery.md`). O que falta é o conteúdo próprio:

- O Build é um job do barramento: `recordJobTelemetry` (`cmd/forge-sdd/worker.go:68-99`) lê o log do job com `monitor.ReadJobUsage` (`internal/monitor/usage.go:168-174`: tokens, cache, turnos, duração, modelo e custo medidos pelo evento `result`) e chama `session record` com `--phase <id do job>`, `--agent-path <papel>` e `--medicao`.
- O Discovery é sessão interativa (`internal/pipeline/launcher.go:104`, sem job): quem grava é o agente seguindo a skill, sem `--agent-path`, sem `--phase`, sem turnos, duração nem modelo. Os tokens saem de `EstimateSessionUsage` (`usage.go:130-166`, bytes do arquivo da spec) com `medicao: estimado`, ou saem zerados quando o agente os passa (exemplo na `main`: `session-2026-11-19T214726Z.json` tem `agent_path: null` e tokens 0; o conserto é a task 3.1 da 105).
- O schema (`internal/monitor/session_record.go:14-37`, `sdd/.metrics/schema.json`) não tem papel nem bloco por tipo. O único ramo por tipo é `ClassifySessionType` (`cmd/forge-sdd/session.go:223-236`), só para o `report`, decidido de propósito "sem exigir um campo novo no schema" (commit `641ff89`, iniciativa 55). Ele também classifica como `discovery` qualquer caminho sob `sdd/discovery/`, inclusive `features/` e `fixes/`: não serve de gatilho.

## Evidência

- `git blame cmd/forge-sdd/session.go` (`ClassifySessionType`): `641ff89`, 2026-09-05. `git log -S"func RecordSessionMetrics"`: schema e `session record` nascem em `073b4ee` (feat-01-01).
- Na `main`, nenhum arquivo de `sdd/.metrics/` tem `"medicao": "real"`: nenhuma sessão commitada tem medida real; o caminho real existe só no código do worker (a verificar em execução).
- Nenhum leitor de `session-*.json` usa `DisallowUnknownFields`: `readSessionFiles` (`session.go:157-190`), `ReadSessions` (`usage.go:228-249`) e `countApprovedCycles` (`autopilot.go:71-99`) usam `json.Unmarshal`. O modo estrito está em `forge.json` (`internal/config/forge.go:272`; a referência 198-200 do pedido está defasada), em `internal/harness/harness.go:22` (contratos) e em `eval.go:150` (arquivo `sim`); nenhum lê sessão.

## Feature Correlacionada

- Origem: `internal/monitor/session_record.go` (`SessionMetrics`) e `cmd/forge-sdd/session.go` (`sessionRecordCmd`), introduzidos na feat-01-01 e estendidos na feat-55-01/02 (cobertura de `/discovery`, `/split-features`, `/nova-feature`, `/archive`).
- Causa-raiz: o schema nasceu único e plano; a feat-55 deu cobertura ao Discovery só registrando a sessão, sem dado próprio do tipo, e o worker preenche o resto só para jobs.
- Evidência do defeito: acima. No arquivo de Discovery não há nenhum dado da spec produzida (artefatos, features, fixes).

## Decisões do projeto aplicadas

- ARQ-01: a regra de coletar e a estrutura vivem em `internal/monitor`; `cmd/` só repassa a flag nova.
- ARQ-05: `--dry-run` novo em `session record` não cria diretório nem arquivo (hoje a flag não existe: `session.go:231-251`).
- SEG-01/SEG-05: o bloco guarda só números e o id da iniciativa saneado por `notify.Sanitize`; nenhum caminho, conteúdo de spec nem texto livre.
- SEG-02 / ADR-0018: a gravação do JSON passa a usar `scaffold.WriteFileSafe` (hoje `os.WriteFile`, `session_record.go:61`).
- DUP-01: sem arquivo novo nem tipo novo de sessão; um único registro por sessão.
- Constitution e ADRs: sem conflito; nenhuma ADR nova (campo opcional no schema 1.0).

## Modelo do módulo

`enxuto`: uma struct opcional e uma função de coleta em pacote existente (`internal/monitor`), sem pacote novo e sem dependência nova.

## Correção

1. `monitor.SessionMetrics` ganha `Spec *SpecMetrics` com `json:"spec,omitempty"`. Ponteiro com `omitempty`: sessão sem o bloco gera arquivo idêntico ao de hoje (diferente de `complexidade`, sempre `null` por ser campo reservado).
2. `SpecMetrics` (só números, enum e o id; nenhum caminho):

| Campo | Tipo | Quando | Fonte (derivável de arquivo real) |
|---|---|---|---|
| `tipo` | `discovery`, `feature` ou `fix` | sempre | caminho de `--feature` |
| `iniciativa` | string (máx. 80, `notify.Sanitize`) | sempre | pasta `<id>-<nome>` de `--feature` |
| `bytes` | int | sempre | `os.Lstat` do arquivo de `--feature` |
| `secoes_obrigatorias` | int (de 6) | `feature` e `fix` | quantos dos seis títulos `## ` obrigatórios da skill aparecem no arquivo (leitura local; só a contagem é gravada) |
| `criteria_bytes`, `plan_bytes`, `status_bytes` | int | `discovery` | `os.Lstat` dos arquivos irmãos |
| `features_count`, `fixes_count` | int | `discovery` | contagem de `features/*.md` e `fixes/*.md` regulares |

   Arquivo ausente, não regular (link simbólico, FIFO) ou com erro de leitura: o número fica 0. Valor inválido (negativo, `iniciativa` vazia ou acima de 80, `tipo` fora do enum) não entra no bloco.
3. **Discriminador (recomendação, ver "A confirmar"):** flag explícita nova `--sessao spec`, passada só pelas references de `discovery.md` (passo 10), `nova-feature.md` (passo 8) e `novo-fix.md` (passo 7). O bloco só é gravado quando a flag está presente E `--feature` é arquivo regular dentro de `sdd/discovery/<id>-<nome>/` (`<id>-<nome>` casando `^[0-9a-z]+-[a-z0-9-]+$`): `discovery.md` (tipo `discovery`), `features/feat-*.md` (`feature`) ou `fixes/fix-*.md` (`fix`). Sem `..`, sem link simbólico e dentro do projeto (`EvalSymlinks`). Flag presente com caminho inválido: grava o arquivo sem o bloco e avisa em stderr (`notify.Sanitize`).
   - Por que o caminho sozinho não basta: o Builder grava com `--feature` apontando para o mesmo `features/feat-*.md` ou `fixes/fix-*.md`. Evidência: `implementar.md:27` passa `--feature "<spec>" --phase "<id>" --agent-path "lider,builder,governance"` e `recordJobTelemetry` (`worker.go:86-97`) passa `--feature job.SpecPath --phase job.ID --agent-path <papel>`; `worker-job.md` não chama `session record`.
   - Alternativa descartada, "ausência de `--phase`": hoje as references de concepção não passam `--phase` nem `--agent-path`, então funcionaria sem mudar texto; mas depende de o agente do Build nunca omitir `--phase` e de o de concepção nunca incluí-lo. Uma omissão desfaz o isolamento em silêncio e grava `spec` em sessão de Build. A flag explícita falha para o lado seguro (sem flag, sem bloco), custa uma linha por reference e é testável.
4. **Fora do bloco:** `/split-features` (`split-features.md`, passo 10, `--feature ".../features/"`) não passa a flag e grava como hoje, sem `spec`. Mesmo que o agente passasse a flag, `--feature` é pasta e não arquivo: sem bloco.
5. A coleta do bloco é melhor esforço: qualquer falha ao montá-lo grava o arquivo sem ele e avisa em stderr. Falha ao gravar o JSON mantém o comportamento atual (erro e saída diferente de 0, `session.go:116-119`). Decisão do dono.
6. `sdd/.metrics/schema.json` e `internal/scaffold/templates/sdd/.metrics/schema.json.tmpl` documentam `spec` como opcional, e `--sessao` entra no `Long` do comando.
7. `RecordSessionMetrics` grava por `scaffold.WriteFileSafe` (temporário exclusivo e rename), mantendo o sufixo `-2`, `-3`. Sem ciclo: `go list -deps ./internal/scaffold` não inclui `monitor`.
8. `session record --dry-run` (flag nova): imprime o registro em JSON na saída padrão, sem `MkdirAll` nem gravação.
9. **Textos afetados (uma linha por reference, acréscimo de ` --sessao spec`, cerca de 14 bytes cada):** `.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md`; espelhos `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md.tmpl`; goldens `internal/scaffold/testdata/golden/.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md` (TST-02). `sdd-concepcao/SKILL.md` (2045 de 2048 bytes) NÃO muda. O P-24 mede `SKILL.md`, adaptadores e instruções (`gate.go:786-`), não `references/`.
10. Fora de escopo: tokens zerados e `init_at` (105/3.1); `ClassifySessionType`; resultado do Gate Camada A, tentativas, RNs, tasks do plano e duração ("a verificar", só entram com fonte medida).

## Relações (sem editar nada delas)

- **105 task 3.1** (tokens zerados e `init_at`): mesmo arquivo `session-*.json`, mesma struct `SessionMetrics` e mesmo `sessionRecordCmd`; colisão textual provável e nenhuma semântica (campos distintos). Sequência recomendada: este fix entra primeiro (é `M` e a 3.1 está na onda 3 da 105) e a 3.1 rebaseia sobre a struct; se a 3.1 sair antes, este fix só soma o campo.
- **105 task 2.7 e 106 task 2.6** reescrevem essas mesmas references (e templates e goldens): **Sequenciar com** elas, em série e rebaseando; aqui é uma linha por arquivo. Não editei nada delas (não conferi o conteúdo dessas tasks sobre as references).
- **105 task 3.2** (agregador): lê com `json.Unmarshal` frouxo; o campo novo não quebra e pode ser consumido depois.
- **105 task 3.6** (`achados.csv`) e **106** (veredito com `head_sha` e `ponto`): arquivos distintos (`sdd/memory/logs/achados.csv`, parecer); sem colisão.
- `forge report` e `doctor` (`AggregateSessionMetrics`) ignoram o campo novo.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Schema e gravação da sessão | `internal/monitor/session_record.go`, `sdd/.metrics/schema.json` | `SessionMetrics` plano, sem bloco por tipo | estender | Um campo opcional no mesmo registro atende o pedido (DUP-01) |
| Tipo da sessão | `ClassifySessionType` (`session.go:223`) | Heurística por caminho só para o `report`; confunde `features/` do Builder com Discovery | reusar | Fica sem mudança: não distingue especificar de construir; o gatilho é a flag `--sessao spec` mais o caminho |
| Distinguir especificação de construção | `implementar.md:27`, `worker.go:86`, references de concepção | Build passa `--phase` e `--agent-path`; concepção não passa nenhum | criar | Flag explícita `--sessao`; a ausência de `--phase` é frágil |
| Medição real do Build | `ReadJobUsage`, `recordJobTelemetry` | Só para jobs do barramento | reusar | Fica sem mudança: Discovery não é job |
| Estimativa por bytes | `EstimateSessionUsage` | Serve só para tokens | reusar | Fica sem mudança: o bloco novo mede artefatos, não tokens |
| Gravação segura | `scaffold.WriteFileSafe` (`hooks.go:236`), ADR-0018 | Exportada; `scaffold` não importa `monitor` | reusar | Sem ciclo; SEG-02 |
| Saneamento | `notify.Sanitize` (`notify.go:122`) | Remove ANSI e controle e limita o tamanho | reusar | SEG-01 e SEG-05 |
| Contagem de seções obrigatórias | `checkP03ReuseTable`, `gate.CriteriaHeading` | Só P-03 e P-25 olham seções da spec | criar | Contar os seis títulos é local ao bloco; reusar o gate fica a verificar |
| `--dry-run` em `session record` | `sessionRecordCmd` | Flag inexistente | criar | ARQ-05 |
| Leitores do arquivo de sessão | `readSessionFiles`, `ReadSessions`, `countApprovedCycles` | Todos com `json.Unmarshal` frouxo | reusar | Fica sem mudança: compatíveis nos dois sentidos |

## Dependências novas

Nenhuma (`go.mod` intacto).

## Rules e IDs da Constituição Aplicáveis

ARQ-01, ARQ-05, ARQ-11, SEG-01, SEG-02, SEG-05, DUP-01, DUP-02, COD-03, TST-01, TST-02, TST-03.

## Decisões do dono (20/11/2026)

- Origem: iniciativa 97 (`fix-97-01`).
- Falha ao gravar o JSON mantém o comportamento atual (erro); só a coleta do `spec` é melhor esforço, com aviso em stderr.
- Gatilho: `/discovery`, `/nova-feature` e `/novo-fix` gravam `spec`; `/split-features` não ("todas elas menos split feature").

## A confirmar com o dono

1. Discriminador: recomendo a flag `--sessao spec` (acima), em vez de ausência de `--phase`. A escolha é sua.
2. Resultado do Gate Camada A, tentativas e tempo dependem de fonte medida (105/3.1); entram depois?

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go build ./...
go test -list 'TestSessionRecordDiscoveryGravaSpec|TestSessionRecordNovaFeatureGravaSpec|TestSessionRecordNovoFixGravaSpec' ./cmd/forge-sdd | grep -F 'TestSessionRecordNovoFixGravaSpec' >/dev/null
go test ./cmd/forge-sdd -run 'TestSessionRecordDiscoveryGravaSpec|TestSessionRecordNovaFeatureGravaSpec|TestSessionRecordNovoFixGravaSpec|TestSessionRecordArquivoAntigoSemSpecAindaLe|TestSpecMetricsContaFeaturesEFixes' -count=1
# negativo: /split-features, Builder com --phase e --agent-path e sessão sem a flag não gravam spec, mesmo com --feature em features/ ou fixes/
go test -list 'TestSessionRecordSplitFeaturesNaoGravaSpec|TestSessionRecordBuilderNaoGravaSpecMesmoEmFeatures|TestSessionRecordSemFlagSessaoNaoGravaSpec|TestSessionRecordFlagComPastaNaoGravaSpec' ./cmd/forge-sdd | grep -F 'TestSessionRecordBuilderNaoGravaSpecMesmoEmFeatures' >/dev/null
go test ./cmd/forge-sdd -run 'TestSessionRecordSplitFeaturesNaoGravaSpec|TestSessionRecordBuilderNaoGravaSpecMesmoEmFeatures|TestSessionRecordSemFlagSessaoNaoGravaSpec|TestSessionRecordFlagComPastaNaoGravaSpec' -count=1
# negativo: caminho com .., link simbólico, fora de sdd/discovery ou pasta ausente grava o arquivo sem spec e avisa em stderr, sem quebrar o comando
go test -list 'TestSessionRecordSpecComPontosPontoNaoGrava|TestSessionRecordSpecSimbolicoNaoSegue|TestSessionRecordSpecForaDeDiscoveryNaoGrava|TestSessionRecordSpecPastaAusenteGravaSemSpec' ./cmd/forge-sdd | grep -F 'TestSessionRecordSpecSimbolicoNaoSegue' >/dev/null
go test ./cmd/forge-sdd -run 'TestSessionRecordSpecComPontosPontoNaoGrava|TestSessionRecordSpecSimbolicoNaoSegue|TestSessionRecordSpecForaDeDiscoveryNaoGrava|TestSessionRecordSpecPastaAusenteGravaSemSpec' -count=1
# negativo: --dry-run não cria diretório nem arquivo; iniciativa com ANSI sai saneada; o JSON não carrega caminho nem conteúdo; falha ao gravar o JSON ainda devolve erro
go test -list 'TestSessionRecordDryRunNaoGrava|TestSpecMetricsIniciativaSaneada|TestSpecMetricsJSONSemCaminhoNemConteudo|TestSessionRecordFalhaDeGravacaoDevolveErro' ./cmd/forge-sdd ./internal/monitor | grep -F 'TestSessionRecordDryRunNaoGrava' >/dev/null
go test ./cmd/forge-sdd ./internal/monitor -run 'TestSessionRecordDryRunNaoGrava|TestSpecMetricsIniciativaSaneada|TestSpecMetricsJSONSemCaminhoNemConteudo|TestSessionRecordFalhaDeGravacaoDevolveErro' -count=1
# negativo: campo inválido (negativo, iniciativa vazia ou acima de 80, tipo fora do enum) não entra no bloco; leitores frouxos seguem lendo o arquivo com spec
go test -list 'TestSpecMetricsCampoInvalidoNaoEntra|TestLeitoresDeSessaoToleramSpec' ./internal/monitor ./cmd/forge-sdd | grep -F 'TestSpecMetricsCampoInvalidoNaoEntra' >/dev/null
go test ./cmd/forge-sdd ./internal/monitor -run 'TestSpecMetricsCampoInvalidoNaoEntra|TestLeitoresDeSessaoToleramSpec' -count=1
# as três references de concepção passam a flag
grep -F -- '--sessao spec' .agents/skills/sdd-concepcao/references/discovery.md .agents/skills/sdd-concepcao/references/nova-feature.md .agents/skills/sdd-concepcao/references/novo-fix.md >/dev/null
go vet ./... && go test ./internal/scaffold/ ./internal/monitor/ ./cmd/forge-sdd/ ./internal/gate/ -count=1
```
