# fix-89-01 — O `/toca-ai` presume workers; quem trabalha no chat fica sem esteira

**Origem:** iniciativa `89-comando-toca-ai` (feat-89-01 e feat-89-02); relaciona-se à ADR-0015 (modo `chat`, proposta) e à decisão do dono da 115 (o chat orquestra e escala os subagentes).

## Sintoma

Quem usa o forge no chat, sem ter rodado `forge up`, pede `/toca-ai`. O comando grava as etapas, passa o pré-voo e `forge pipeline toca-ai` responde "Etapas enfileiradas", e a skill manda "entregue e pare; o chat fica livre". Nenhum worker existe, então os jobs ficam parados na fila e ninguém executa nada, sem aviso. Em 22/11/2026 isso levou a oferecer "subir os workers" a quem tinha escolhido, de propósito, trabalhar no chat.

## Apuração

1. `.agents/skills/sdd-operacao/references/toca-ai.md`, passo 4 ("Entregue e pare... a esteira executa as etapas"), e o passo 1 ("sem `.forge/config/forge.json`, oriente `forge pipeline init`") tratam o modo pipeline como o único caminho. Não há ramo para "sem worker ativo".
2. `cmd/forge-sdd/pipeline_goal.go` imprime "Etapas enfileiradas" sem checar se há worker vivo (`monitor.InspectProjectProcesses` já sabe listar workers do projeto).
3. `forge.json` presente significa "modo pipeline configurado", não "workers rodando". Nada no produto distingue os dois, e o texto da skill os confunde.
4. A ADR-0015 define o modo `chat` (subagentes no mesmo chat) mas está `proposto`; não existe `pipeline_mode: chat` no código. Sem esse modo, o produto não tem como dizer ao chat "orquestre você".

**Causa-raiz:** o produto entrega um único modelo mental (esteira com workers) e o confunde com a configuração; a instrução não pergunta nem detecta como a pessoa quer trabalhar. Quem criou a ferramenta percebeu; quem só usa não perceberia.

## Feature Correlacionada

`sdd/discovery/89-comando-toca-ai/features/` (feat-89-01 e feat-89-02) e `sdd/discovery/104-pipeline-em-subagentes/` (ADR-0015). Evidência: sessão de 22/11/2026, 6 jobs da iniciativa 116 enfileirados e parados.

## Modelo do módulo

Enxuto (`ARQ-11`): ajustar `cmd/forge-sdd` (só aviso e flag), reaproveitar `internal/monitor` e `internal/doctor`, e editar texto de skills/rules com seus templates e goldens. Sem pacote nem dependência novos. Não implementa o modo `chat` da ADR-0015; só impede que o produto finja uma esteira que não existe.

## Decisões do projeto aplicadas

- ADR-0004: jobs trafegam pelo bus e são consumidos por workers.
- ADR-0015 (proposta): subagentes no mesmo chat; este fix prepara o terreno sem decidir a ADR.
- Decisão do dono (22/11/2026): quem não iniciou o `forge up` escolheu o chat; o chat orquestra e escala os subagentes em paralelo na própria sessão.
- Regra de ouro do `/toca-ai`: nunca mandar a pessoa rodar algo; ela vale também para "suba os workers".

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Detecção de workers | `internal/monitor/ps.go` (`InspectProjectProcesses`) | Lista processos do projeto com papel inferido | Reusar | Evita detector novo (`DUP-01`) |
| Aviso de worker de outra versão | `internal/doctor/harness.go` (`workers`) | Checagem `workers` existe | Estender: fila com jobs `queued` e nenhum worker ativo | Mesmo ponto de diagnóstico |
| Comando de entrega | `cmd/forge-sdd/pipeline_goal.go` | Enfileira e imprime sem checar | Acrescentar o aviso | Corrige o sintoma na origem |
| Texto da skill | `.agents/skills/sdd-operacao/references/toca-ai.md` e template | Só o caminho com esteira | Ramificar por "workers ativos" | Causa-raiz 1 |
| Regras comuns | `.agents/rules/` | Sem regra sobre "como a pessoa quer trabalhar" | Uma frase na rule de fluxo | Evita repetir em cada skill |

## Dependências novas (`SEG-07`)

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-11`, `DUP-01`, `FLX-01`, `COD-01`, `TST-01`, `TST-03`.

## Escopo

1. **Binário:** `forge pipeline toca-ai` (e o `dispatch` quando enfileira) confere workers ativos com `monitor.InspectProjectProcesses`. Sem worker, mantém o enfileiramento (retomável) mas imprime o aviso "nenhum worker ativo: as etapas ficam na fila; para executar neste chat, oriente os subagentes (veja a skill) ou inicie os workers" e devolve `workers_ativos: 0` em `--json`, se o comando já tiver saída JSON. Nunca falha por isso.
2. **`forge doctor`:** nova checagem **separada** de `checaWorkers` (que só mede versão), em `internal/doctor`, avisa (não falha) quando há job `queued` e nenhum worker ativo. Reaproveita `monitor.InspectProjectProcesses` para detectar o worker e lê a fila por `internal/queue` (ou `internal/bus`, conforme o código real), sem duplicar nenhum dos dois.
3. **Skill `toca-ai`:** o passo 4 vira dois ramos. *Com workers ativos:* entregue e pare. *Sem workers:* a pessoa escolheu o chat; o chat orquestra, dispara as etapas independentes do DAG em paralelo como subagentes da própria sessão (um worktree por subagente, Reviewer somente leitura, retorno tratado como dado), e nunca sugere `forge up`. A regra "pré-voo e consentimentos" continua igual.
4. **Rule de fluxo:** acrescentar, em `.agents/rules/sdd-fluxo-e-portao`, a frase: "Não presuma `forge up`. Quem não o iniciou trabalha no chat; o chat orquestra os subagentes. Só ofereça workers se a pessoa pedir."
5. **Testes:** criar os quatro testes dos critérios 1 a 4. O teste 4 (`TestTocaAIRamoChatSemWorker`) mora em `internal/scaffold`, ao lado de `budget_test.go` e dos goldens, porque `internal/skills` só trata de propor e aprovar skills e não toca o texto dos templates.
6. **Templates e goldens** em dia (`internal/scaffold/templates` e `testdata/golden`), mais a nota na entrada 5.2.0 de `sdd/releases/history.md` (a release beta corrente; sem novo bump).

Fora do escopo: implementar `pipeline_mode: chat` (ADR-0015, iniciativa 104).

## Critérios de Aceitação Executáveis

```bash
# 1. Sem worker ativo, o comando avisa e não falha.
go test ./cmd/forge-sdd -run 'TestTocaAIAvisaSemWorker'

# 2. Com worker ativo, não há aviso.
go test ./cmd/forge-sdd -run 'TestTocaAISemAvisoComWorker'

# 3. O doctor avisa fila parada sem worker.
go test ./internal/doctor -run 'TestDoctorFilaSemWorker'

# 4. A skill tem os dois ramos e proíbe sugerir forge up a quem trabalha no chat.
go test ./internal/scaffold -run 'TestTocaAIRamoChatSemWorker'

# 5. Skill, rule, templates e goldens em dia.
go test ./internal/scaffold

# 6. Build e suíte.
go build ./... && go test ./...
```

## Riscos

| Risco | Mitigação |
|---|---|
| Detector de worker falso-positivo ou falso-negativo | Só avisa (nunca bloqueia) e usa o detector que o `forge status` já usa |
| Orçamento de tokens das skills (`P-24`) | O ramo novo cabe em poucas linhas; o teste de orçamento roda no `go test ./internal/scaffold` |
| Confundir com a ADR-0015 | Escopo exclui o modo `chat`; a ADR segue com a iniciativa 104 |
