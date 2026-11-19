# fix-90-02 — Tokens zerados em sessões de métricas e campo `complexidade` reservado

**Branch:** `fix/90-02-metrics-token-e-complexidade` (a partir de `feat/88-modo-pipeline-avancado`; PR contra ela)
**Origem:** Iniciativa 90, `feat-90-02` (`internal/monitor/usage.go`, `internal/monitor/session_record.go`; skills `.agents/skills/sdd-*/references/`)
**Tamanho:** `M`
**Status:** `ready`

> **Nota de concepção:** a demanda chegou em dois pedaços distintos.
> - Parte A (bug) é independente e urgente: vários skills não passam `--tokens-input`/`--tokens-output` ao chamar `forge session record`, e o próprio CLI não estima quando as flags são omitidas.
> - Parte B (campo `complexidade`) é preparação estrutural: o campo deve aparecer no schema e na struct Go, sem flag de CLI e sem lógica de preenchimento. A ADR que definirá como calcular complexidade está prevista para ser cadastrada antes de implementar o preenchimento real. Dois momentos de uso previstos pelo usuário: (a) em nível de discovery e (b) conferido depois de implementar. O campo não recebe `omitempty` para distinguir "não preenchido" (null) de ausente.

---

## Sintoma

Arquivos `sdd/.metrics/session-*.json` têm `tokens_input: 0` e `tokens_output: 0` mesmo em sessões com interação real. Dois grupos históricos explicam os zeros:

| Grupo | Exemplo | Phase | Evidence |
|---|---|---|---|
| **1 — Skill-template call (pré e pós-90-02)** | `session-2026-11-19T041917Z.json` | — | Corresponde ao template de `nova-feature.md` sem flags de token |
| **2 — Old-worker path (pré-90-02)** | `session-2026-11-15T015806Z.json` | `job-20261115-…-review` | Worker anterior ao commit `5486a27` não chamava `ReadJobUsage` |
| **3 — Agente ignorou flags (implementar.md)** | `session-2026-11-15T144210Z.json` | — | `implementar.md` já inclui as flags; o agente não as usou — evidência de que a conformidade do template não garante o preenchimento; justifica o fallback no CLI |

**Foco deste fix:**
- **Grupos 1 e 3 (skill-template + CLI sem estimativa):** corrigir os templates e adicionar estimativa automática no CLI.
- **Grupo 2** não será corrigido retroativamente: são registros anteriores a `5486a27`.

**Observação sobre o estado do binário:** nenhum dos arquivos rastreados em `sdd/.metrics/` contém o campo `medicao`, e o binário instalado nesta máquina rejeita `--medicao`. Isso confirma que o binário em uso pré-data o commit `5486a27` (feat-90-02). O código na base `feat/88-modo-pipeline-avancado` tem `ReadJobUsage`, mas o usuário precisa instalar o binário compilado a partir dessa branch para que o worker grave tokens nos jobs do pipeline. As sessões de skill-template call persistirão com zeros mesmo após instalar o novo binário, porque a coleta pelo worker não substitui as chamadas manuais dos agentes — essa é a parte do "ainda ficam" que este fix resolve.

## Feature Correlacionada

- **Arquivo de origem:** `sdd/discovery/90-esteira-economica-e-autonoma/features/feat-90-02-telemetria-real-por-job.md`
- **Causa-raiz (verificada nos skill templates e no CLI):**
  1. **Templates de skill incompletos.** `feat-90-02` introduziu `monitor.ReadJobUsage` no worker e corrigiu `implementar.md` com `--tokens-input`/`--tokens-output`. Os demais templates não foram atualizados:
     - `.agents/skills/sdd-concepcao/references/novo-fix.md` → sem tokens
     - `.agents/skills/sdd-concepcao/references/nova-feature.md` → sem tokens
     - `.agents/skills/sdd-concepcao/references/discovery.md` → sem tokens
     - `.agents/skills/sdd-governanca/references/revisar.md` → sem tokens
     - `.agents/skills/sdd-operacao/references/lider.md` → parcial (sem tokens)
     - `.agents/skills/sdd-planejamento/references/split-features.md` → sem tokens
     - `.agents/skills/sdd-operacao/references/archive.md` → sem tokens
  2. **Cópias nos templates do scaffold.** Cada um dos 7 arquivos acima tem um correspondente `.tmpl` em `internal/scaffold/templates/.agents/skills/…`. Esses templates são embutidos via `//go:embed all:templates` em `internal/scaffold/scaffold.go` e instanciados por `forge init`. Sem a correção dos `.tmpl`, todo projeto gerado pelo scaffold herda o bug.
  3. **CLI sem fallback de estimativa.** `forge session record` aceita `tokens_input=0` sem alerta e sem estimativa automática. O cobra permite distinguir "flag não usada" de "flag definida como 0" via `cmd.Flags().Changed(...)`, mas o comando atual não usa essa distinção. `feat-90-02` não adicionou esse fallback ao CLI — apenas ao worker.
- **Evidência:**
  - `session-2026-11-19T041917Z.json`: feature 95-01, `agent_path: null`, sem model, sem turns → padrão exato de `nova-feature.md`.
  - `session-2026-11-15T144210Z.json`: builder + fix-84-01; `implementar.md` já inclui as flags, ainda assim zeros → o agente não seguiu a instrução. Isso confirma que a conformidade do template não garante o preenchimento, e justifica o fallback no CLI como segunda linha de defesa.
  - Comparação direta entre `implementar.md` (flags incluídas) e demais 6 arquivos (flags ausentes).

---

## Modelo do módulo

**Enxuto** (ARQ-11): sem módulo novo. Altera arquivos existentes em:
- `.agents/skills/sdd-*/references/` (7 arquivos de referência de skill)
- `internal/scaffold/templates/.agents/skills/sdd-*/references/` (7 arquivos `.tmpl` correspondentes)
- `cmd/forge-sdd/session.go` (verificação de `Changed()` + chamada à função de estimativa)
- `internal/monitor/usage.go` (nova função `EstimateSessionUsage(dir, feature string, filesTouched []string) JobUsage`)
- `sdd/.metrics/schema.json` (campo `complexidade`)
- `internal/scaffold/templates/sdd/.metrics/schema.json.tmpl` (campo `complexidade`; também sincronizar campos opcionais de feat-90-02 ausentes: `tokens_cache_read`, `tokens_cache_creation`, `cost_usd_tabela`, `medicao`)
- `internal/monitor/session_record.go` (struct `SessionMetrics` — campo `complexidade`)

Golden tests do scaffold devem ser regenerados após atualizar os `.tmpl` (flag `-update` em `golden_test.go`).

---

## Decisões do projeto aplicadas

- **ADR-0004** (Pipeline Multi-Sessão): a telemetria de sessão é o mecanismo de rastreabilidade; corrigir os templates mantém o contrato da ADR.
- **ADR-0003** (Skills por Estação): os arquivos de referência são as instruções canônicas por estação; corrigi-los é a única forma de propagar a mudança a todos os agentes.
- **ARQ-01** (cmd/ só adapta CLI): a verificação de `Changed()` e a chamada a `monitor.EstimateSessionUsage` ficam em `session.go` (adaptação de flags); a lógica de estimativa fica em `internal/monitor/usage.go` (domínio) — conforme. `feat-90-02` usou o mesmo padrão ao mover `estimateJobTokens` de `cmd/` para `internal/monitor`.
- **ARQ-11** (Spec declara o modelo do módulo): módulo enxuto documentado acima.
- **COD-04** (código morto fora): o campo `complexidade` é adicionado como reservado, sem lógica de preenchimento. O usuário solicitou "deixar o caminho preparado" para a ADR futura; nenhuma branch de execução depende do campo neste fix.
- **COD-06** (TODO com ID): comentário de placeholder no campo `complexidade` deve ser `// reservado; preenchimento definido por ADR futura` (sem referência a fix-90-02, que ficará fechado).
- **DUP-01** (um conceito, uma casa): `EstimateSessionUsage` compartilha constantes e padrões de `EstimateJobUsage` para evitar duplicação.
- **GOV-01** (ADR para conflito com Constitution): a adição de campo opcional nulo é retrocompatível. A lógica de preenchimento virá com ADR futura.
- **SEG-07** (dependências novas): nenhuma.

---

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Flags de token em session record | `implementar.md` | `--tokens-input`/`--tokens-output` com nota "nunca zerados" | **replicar** nos demais 6 templates + 7 tmpl | Uniformidade de instrução entre estações |
| Estimativa de tokens a partir de bytes | `internal/monitor/tokens.go` | `EstimateTokensFromBytes` | **reusar** dentro de `EstimateSessionUsage` | Sem nova lógica; ARQ-01 |
| `cmd.Flags().Changed(...)` | `cmd/forge-sdd/session.go` | Não usa para tokens | **adicionar** | Distingue omissão de zero explícito |
| Estimativa de tokens no pipeline | `internal/monitor/usage.go` | `ReadJobUsage`, `EstimateJobUsage` (padrões: spec 4000 bytes, log 2000 bytes) | **estender** com `EstimateSessionUsage` compartilhando os mesmos padrões e constantes; evitar duplicação de lógica (DUP-01) | ARQ-01 |
| Cópias de template nos scaffold templates | `internal/scaffold/templates/.agents/skills/` | 7 arquivos `.tmpl` com mesmo texto do `.agents/skills/` | **corrigir** junto com os originais | Todo novo projeto herdaria o bug |
| schema.json nos scaffold templates | `find internal/scaffold/templates -name 'schema.json*'` | `schema.json.tmpl` existe | **corrigir** junto com `sdd/.metrics/schema.json`; também sincronizar campos opcionais de feat-90-02 ausentes no `.tmpl` | Todo projeto gerado por `forge init` recebe um schema desatualizado sem a correção |
| Campo complexidade | toda a árvore `internal/`, `cmd/`, `internal/scaffold/templates/` | Ausente | **criar** | Primeira inclusão |
| Outros callers de RecordSessionMetrics | `grep -rn RecordSessionMetrics --include=*.go` | `worker.go` (tokens via ReadJobUsage), `archive.go` (tokens via SumUsage), `session.go` | **sem alteração** em worker e archive | Esses dois caminhos já fornecem tokens |

---

## Dependências novas

Nenhuma. Sem `go get` necessário.

---

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-11`, `COD-04`, `COD-06`, `DUP-01`, `DUP-02`, `GOV-01`, `SEG-07`, `TST-01`, `TST-04`.
Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

---

## Critérios de Aceitação Executáveis

### Parte A — Bug: tokens zerados

```bash
# A1a. Função de estimativa em internal/monitor
go test ./internal/monitor/... -run TestEstimateSessionUsage

# O teste deve verificar:
# - Com feature existente no disco → tokens_input = EstimateTokensFromBytes(bytesDaFeature * 4)
# - Com filesTouched não vazios → tokens_output = EstimateTokensFromBytes(somaDosBytes)
# - Com feature ausente ou inexistente → tokens_input usa fallback de 4000 bytes (padrão de EstimateJobUsage)
# - Com filesTouched vazio → tokens_output usa fallback de 2000 bytes (padrão de EstimateJobUsage)
# - Medicao retornado = "estimado" em todos os casos
# (EstimateSessionUsage usa os mesmos padrões e constantes de EstimateJobUsage — DUP-01)

# A1b. Comportamento de Changed() em cmd (integração):
go test ./cmd/forge-sdd/... -run TestSessionRecord_EstimaQuandoFlagsOmitidas

# O teste deve verificar:
# - Chamada sem --tokens-input e sem --tokens-output → tokens_input > 0, tokens_output > 0, medicao = "estimado"
# - Chamada com --tokens-input 0 explícito → tokens_input = 0 (não estima input)
# - Chamada com --tokens-input 1500 → tokens_input = 1500 (mantém o valor)
# - Chamada com apenas --tokens-input 1500 (sem --tokens-output) →
#   tokens_input = 1500 (Changed = true: mantém), tokens_output estimado (Changed = false: estima)
# Regra: cada flag é avaliada independentemente via Changed(). Se Changed = false, estima; se Changed = true, mantém.

# A2. Templates de skill incluem --tokens-input (go test que lê os arquivos do sistema de arquivos):
go test ./cmd/forge-sdd/... -run TestSkillTemplates_IncluemTokensInput

# O teste deve:
# - Percorrer explicitamente os seguintes arquivos de referência em .agents/skills/:
#   novo-fix.md, nova-feature.md, discovery.md, revisar.md, lider.md, split-features.md, archive.md
# - Percorrer os 7 arquivos .tmpl correspondentes em internal/scaffold/templates/.agents/skills/
# - Para cada arquivo que contenha "forge session record",
#   verificar que o trecho contém "--tokens-input"
# - Incluir implementar.md e seu .tmpl como regressão (já devem ter a flag)
# - Excluir explicitamente sdd-concepcao/SKILL.md(.tmpl) e os arquivos de instrução raiz
#   (CLAUDE.md.tmpl, GEMINI.md.tmpl, AGENTS.md.tmpl, copilot-instructions.md.tmpl) —
#   esses citam "forge session record" em contexto de documentação, não de instrução de execução

# A3. Golden tests do scaffold passam após regeneração:
go test ./internal/scaffold/... -run TestGoldenInit -update  # regenera
go test ./internal/scaffold/...  # valida

# A4. Budget de contexto não violado:
go test ./internal/scaffold/... -run TestContextBudget

# A5. Regressão completa
go vet ./... && go test ./...
```

**Comportamento esperado após a correção:**
- Todo chamador de `forge session record` nos skill templates inclui `--tokens-input <n> --tokens-output <n>` com a nota "estime de forma realista (1 token ≈ 4 caracteres); nunca zeros se houve interação real".
- O comando `forge session record` chama `monitor.EstimateSessionUsage` para cada flag de token que não foi `Changed`; o resultado estimado substitui o valor padrão 0 e `medicao` é definido como `estimado`.
- Uma flag `Changed` (passada explicitamente, mesmo com valor 0) é respeitada sem sobrescrita.
- Regra de estimativa simétrica: cada flag é avaliada independentemente via `Changed()`.

### Parte B — Melhoria: campo `complexidade`

```bash
# B1. O campo complexidade existe na struct e serializa como null quando não definido
go test ./internal/monitor/... -run TestSessionMetrics_Complexidade

# O teste deve verificar:
# - marshal de SessionMetrics sem complexidade (Complexidade == nil) → JSON contém "complexidade":null
# - marshal de SessionMetrics com Complexidade apontando para uma string qualquer
#   → JSON contém "complexidade":"<valor>"
# - unmarshal de JSON antigo (sem campo complexidade) → struct.Complexidade == nil
# (Tipo provisional: *string; a ADR futura pode redefinir o tipo sem breaking change
#  desde que seja antes de qualquer preenchimento real que gravaria o formato atual)

# B2. Golden tests passam após regeneração (schema.json.tmpl atualizado):
go test ./internal/scaffold/... -run TestGoldenInit -update  # regenera
go test ./internal/scaffold/...

# B3. Regressão completa
go vet ./... && go test ./...
```

**Comportamento esperado:**
- `schema.json` contém: `"complexidade": "string | null (opcional; aguarda ADR de complexidade; dois momentos de uso: discovery e pós-implementação)"`
- `schema.json.tmpl` contém a mesma linha, além dos campos opcionais de feat-90-02 que faltavam (`tokens_cache_read`, `tokens_cache_creation`, `cost_usd_tabela`, `medicao`).
- `SessionMetrics.Complexidade` é do tipo `*string`, com tag `json:"complexidade"` (sem `omitempty`) para que apareça como `null` em JSON quando não preenchido.
- Nenhuma flag `--complexidade` é adicionada ao `forge session record`.
- Nenhuma lógica preenche o campo automaticamente; fica nulo até a ADR de complexidade ser aprovada e a lógica implementada.
- Arquivos antigos sem o campo continuam válidos (unmarshal retorna `nil` para `*string`).
