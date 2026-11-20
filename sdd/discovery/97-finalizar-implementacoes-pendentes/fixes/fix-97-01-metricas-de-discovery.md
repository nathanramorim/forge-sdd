# fix-97-01 — A sessão de Discovery grava a telemetria genérica, sem bloco próprio de métricas da spec

**Branch:** `fix/97-metricas-de-discovery`
**Origem:** Iniciativa 97 (mesma superfície do `fix-90-02`: campos do `session-*.json`); causa-raiz histórica nas iniciativas 01 e 55, ambas arquivadas
**Tamanho:** `M`
**Status:** `spec`

## Descrição de negócio (release notes)

*   **Métricas de Discovery no mesmo arquivo de sessão:** a sessão que fecha o `/discovery` passa a gravar, em `sdd/.metrics/session-*.json`, o bloco `spec` com o tamanho dos artefatos e a contagem de features e fixes da iniciativa; sessões de Build, Review e arquivos antigos seguem sem o bloco (`fix-97-01`).

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

1. `monitor.SessionMetrics` ganha `Spec *SpecMetrics` com `json:"spec,omitempty"`. Ponteiro com `omitempty`: sessão que não é de Discovery não tem a chave e o arquivo fica idêntico ao de hoje (diferente de `complexidade`, sempre `null` por ser campo reservado).
2. `SpecMetrics` (só números e o id; nenhum caminho):

| Campo | Tipo | Fonte (derivável de arquivo real) |
|---|---|---|
| `iniciativa` | string (máx. 80, `notify.Sanitize`) | nome da pasta `<id>-<nome>` extraído de `--feature` |
| `discovery_bytes` | int | `os.Lstat` de `discovery.md` |
| `criteria_bytes` | int | `os.Lstat` de `criteria.md` |
| `plan_bytes` | int | `os.Lstat` de `plan.md` |
| `status_bytes` | int | `os.Lstat` de `status.md` |
| `features_count` | int | contagem de `features/*.md` regulares |
| `fixes_count` | int | contagem de `fixes/*.md` regulares |

   Arquivo ausente, não regular (link simbólico, FIFO) ou com erro de leitura: o campo de bytes fica 0 e o registro é gravado mesmo assim; a coleta nunca devolve erro. Valor inválido (bytes negativo, `iniciativa` vazia ou acima de 80) não entra no bloco.
3. Gatilho: só quando `--feature` é exatamente `sdd/discovery/<id>-<nome>/discovery.md` (o que o passo 10 de `discovery.md` já passa) e a pasta existe dentro do projeto (sem sair por `..` nem por link simbólico). Qualquer outro `--feature`, inclusive `features/` e `fixes/` do Builder, não ganha o bloco. Não muda skill, template nem golden de skill.
4. `sdd/.metrics/schema.json` e `internal/scaffold/templates/sdd/.metrics/schema.json.tmpl` documentam `spec` como opcional (golden, se houver, no mesmo PR; TST-02).
5. `RecordSessionMetrics` grava por `scaffold.WriteFileSafe` (temporário exclusivo e rename), mantendo o sufixo `-2`, `-3` quando o nome existe. Não há ciclo de importação: `go list -deps ./internal/scaffold` não inclui `monitor`.
6. `session record --dry-run`: calcula e imprime o registro (JSON) na saída padrão, sem `MkdirAll` nem gravação.
7. Fora de escopo: tokens zerados e `init_at` (105/3.1); `ClassifySessionType`; resultado do Gate Camada A, tentativas de Gate 1 e Conferência 1, contagem de RNs e de tasks do plano (os formatos variam entre iniciativas) e duração: "a verificar", só entram com fonte medida.

## Relações (sem editar nada delas)

- **105 task 3.1** (tokens zerados e `init_at`): mesmo arquivo `session-*.json`, mesma struct `SessionMetrics` e mesmo `sessionRecordCmd`; colisão textual provável e nenhuma semântica (campos distintos). Sequência recomendada: este fix entra primeiro (é `M` e a 3.1 está na onda 3 da 105) e a 3.1 rebaseia sobre a struct; se a 3.1 sair antes, este fix só soma o campo.
- **105 task 3.2** (agregador): lê com `json.Unmarshal` frouxo; o campo novo não quebra e pode ser consumido depois.
- **105 task 3.6** (`achados.csv`) e **106** (veredito com `head_sha` e `ponto`): arquivos distintos (`sdd/memory/logs/achados.csv`, parecer); sem colisão.
- `forge report` e `doctor` (`AggregateSessionMetrics`) ignoram o campo novo.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Schema e gravação da sessão | `internal/monitor/session_record.go`, `sdd/.metrics/schema.json` | `SessionMetrics` plano, sem bloco por tipo | estender | Um campo opcional no mesmo registro atende o pedido (DUP-01) |
| Tipo da sessão | `ClassifySessionType` (`session.go:223`) | Heurística por caminho só para o `report`; confunde `features/` com Discovery | reusar | Fica sem mudança: não serve de gatilho, o gatilho é o caminho exato de `discovery.md` |
| Medição real do Build | `ReadJobUsage`, `recordJobTelemetry` | Só para jobs do barramento | reusar | Fica sem mudança: Discovery não é job |
| Estimativa por bytes | `EstimateSessionUsage` | Serve só para tokens | reusar | Fica sem mudança: o bloco novo mede artefatos, não tokens |
| Gravação segura | `scaffold.WriteFileSafe` (`hooks.go:236`), ADR-0018 | Exportada; `scaffold` não importa `monitor` | reusar | Sem ciclo; SEG-02 |
| Saneamento | `notify.Sanitize` (`notify.go:122`) | Remove ANSI e controle e limita o tamanho | reusar | SEG-01 e SEG-05 |
| `--dry-run` em `session record` | `sessionRecordCmd` | Flag inexistente | criar | ARQ-05 |
| Leitores do arquivo de sessão | `readSessionFiles`, `ReadSessions`, `countApprovedCycles` | Todos com `json.Unmarshal` frouxo | reusar | Fica sem mudança: compatíveis nos dois sentidos |

## Dependências novas

Nenhuma (`go.mod` intacto).

## Rules e IDs da Constituição Aplicáveis

ARQ-01, ARQ-05, ARQ-11, SEG-01, SEG-02, SEG-05, DUP-01, DUP-02, COD-03, TST-01, TST-02, TST-03.

## A confirmar com o dono

1. Origem: escolhida a iniciativa 97 (viva, mesma superfície do `fix-90-02`) porque 55 e 01 estão arquivadas (55 é arquivo solto em `sdd/archive/discovery/`, sem `fixes/`) e a 66 trata de `activity.json`, não de `session-*.json`. Confirma a 97?
2. `/novo-fix`, `/nova-feature` e `/split-features` também geram sessões de spec: querem `spec` ali? Esta spec cobre só o `discovery.md`.
3. Falha de gravação do JSON hoje devolve erro e saída diferente de 0 (`session.go:116-119`). Mantém ou vira aviso com saída 0? Esta spec mantém o atual e deixa best-effort só a coleta do bloco.
4. O bloco é fino de propósito (só o que se mede em arquivo). Resultado do Gate Camada A, tentativas e tempo dependem de fonte medida (task 3.1 da 105); entram depois?

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go build ./...
go test -list 'TestSessionRecordDiscoveryGravaSpec' ./cmd/forge-sdd | grep -F 'TestSessionRecordDiscoveryGravaSpec' >/dev/null
go test ./cmd/forge-sdd -run 'TestSessionRecordDiscoveryGravaSpec|TestSessionRecordArquivoAntigoSemSpecAindaLe|TestSpecMetricsContaFeaturesEFixes' -count=1
# negativo: sessão de Build, de features/ e de fixes/ não grava spec; pasta ausente, link simbólico e caminho com .. não gravam spec nem quebram o comando
go test -list 'TestSessionRecordBuildNaoGravaSpec|TestSessionRecordFeaturesEFixesNaoGravamSpec|TestSessionRecordDiscoveryPastaAusenteGravaSemSpec|TestSessionRecordDiscoverySimbolicoNaoSegue|TestSessionRecordDiscoveryComPontosPontoNaoGravaSpec' ./cmd/forge-sdd | grep -F 'TestSessionRecordBuildNaoGravaSpec' >/dev/null
go test ./cmd/forge-sdd -run 'TestSessionRecordBuildNaoGravaSpec|TestSessionRecordFeaturesEFixesNaoGravamSpec|TestSessionRecordDiscoveryPastaAusenteGravaSemSpec|TestSessionRecordDiscoverySimbolicoNaoSegue|TestSessionRecordDiscoveryComPontosPontoNaoGravaSpec' -count=1
# negativo: --dry-run não cria diretório nem arquivo; iniciativa com ANSI e controle sai saneada; o JSON não carrega caminho nem conteúdo
go test -list 'TestSessionRecordDryRunNaoGrava|TestSpecMetricsIniciativaSaneada|TestSpecMetricsJSONSemCaminhoNemConteudo' ./cmd/forge-sdd ./internal/monitor | grep -F 'TestSessionRecordDryRunNaoGrava' >/dev/null
go test ./cmd/forge-sdd ./internal/monitor -run 'TestSessionRecordDryRunNaoGrava|TestSpecMetricsIniciativaSaneada|TestSpecMetricsJSONSemCaminhoNemConteudo' -count=1
# negativo: campo inválido (bytes negativo, iniciativa vazia ou acima de 80) não entra no bloco; leitores frouxos seguem lendo o arquivo com spec
go test -list 'TestSpecMetricsCampoInvalidoNaoEntra|TestLeitoresDeSessaoToleramSpec' ./internal/monitor ./cmd/forge-sdd | grep -F 'TestSpecMetricsCampoInvalidoNaoEntra' >/dev/null
go test ./cmd/forge-sdd ./internal/monitor -run 'TestSpecMetricsCampoInvalidoNaoEntra|TestLeitoresDeSessaoToleramSpec' -count=1
go vet ./... && go test ./internal/scaffold/ ./internal/monitor/ ./cmd/forge-sdd/ -count=1
```
