# feat-89-01 — Motor do `/toca-ai`: `forge pipeline toca-ai` enfileira as etapas e o Go cuida do resto

**Iniciativa:** 89-comando-toca-ai
**Branch:** feat/89-comando-toca-ai (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** nenhuma (usa o modo avançado, o `spec_review`, o `STOP` e o barramento da iniciativa 88, já na base)

## Problema
O Líder, dentro do chat, não consegue digitar `/clear` nem `/goal clear`, e o contexto dele enche em trabalho longo. Se a limpeza, o archive e a parada dependerem do modelo lembrar, o trabalho longo falha. A solução é **encapsular** o que é mecânico em código Go e deixar ao Líder só o que é decisão: planejar e gerar as specs.

## Comportamento (`forge pipeline toca-ai <iniciativa>`)
1. **Entrada:** a pasta da iniciativa já tem `plan.md` (lista de etapas, com `**Depende de:**`) e uma spec por etapa em `features/`, geradas pelo Líder (feat-89-02). O comando não gera spec, só confere que existe.
2. **Portão da spec (FLX-01):** no modo avançado, cria um job `spec_review` (feat-88-04) e **só enfileira as etapas depois do parecer `aprovado`**; sem o modo avançado, exige a aprovação humana da spec e para com a instrução.
3. **Uma etapa = um job** de build no barramento, na ordem do DAG (`**Depende de:**`, `forge bus plan`), com `spec_path` apontando para a spec da etapa. Reenviar não duplica jobs.
4. **O Go faz o que o modelo não consegue:**
   - **Contexto limpo:** cada job roda em contexto novo — no worker headless o agente é um processo novo por job; no modo live, o despachante já envia o `/clear` (`deliverContextReset`, `context: clear`).
   - **Archive:** quando todas as etapas estão `approved`, o auto-archive existente (`CheckAndAutoArchive`) arquiva a iniciativa; o `status.md` acima de 1 KB aciona o Archivist.
   - **Parada:** `.forge/runtime/STOP` e os tetos do Settings (se ligados) já são respeitados pelos workers.
   - **Impasse:** etapa que esgota as tentativas vira `blocked` com motivo no inbox, e a fila segue nas independentes (feat-88-06).
5. **Sem `/goal`:** nada depende de meta do Claude Code; o estado está em `plan.md`, na fila e no barramento.
6. **Retomada:** rodar o comando de novo continua de onde parou (idempotente); o painel (`forge status`) mostra o andamento.

## Pré-voo: tudo que a execução precisa da pessoa é resolvido **antes**, uma única vez
Caso real (`reel-studio`/`marketing`): no meio do trabalho o agente pediu "rode no seu terminal `voice-studio perfil criar … --confirm-own-voice`" e "libere o comando na permissão". Isso quebra a promessa de não interagir. O motor garante que isso não acontece:
1. O Líder grava `requisitos.json` na iniciativa, com tudo que as etapas vão usar: `comandos` (argv exato, ex.: `uv run … voice-studio perfil criar`), `ferramentas` (ex.: `uv`, `python3.11`), `arquivos` de entrada (ex.: `amostra-voz.wav`) e `consentimentos` (id e texto, ex.: "a voz da amostra é minha").
2. `forge pipeline toca-ai --preflight <iniciativa>` confere sozinho o que dá: ferramentas no PATH (`LookPath`), arquivos existentes, e se cada comando está coberto pela permissão do pipeline (`allow_commands` no nível `project`, ou `full`). Mostra **uma única lista** do que falta.
3. **Consentimento** (confirmar que a voz é própria, aceitar custo etc.) só vale se a pessoa responder; o agente nunca confirma por ela. A resposta é gravada em `.forge/runtime/consents.json` (id, texto, data) por `forge pipeline consent <id>` (pergunta interativa, uma vez). Nas etapas, o agente usa o consentimento gravado para passar a flag; sem registro, a etapa nem começa.
4. **Permissão de comando:** no pipeline, o pré-voo propõe os comandos exatos para `allow_commands` e a pessoa aprova uma vez. O agente **executa o comando ele mesmo**; nunca pede "rode no seu terminal".
5. O motor só enfileira as etapas quando o pré-voo está sem pendência (ou quando a pessoa aceita explicitamente seguir sem as etapas afetadas).
6. **Imprevisto no meio:** a etapa vira `blocked` com o item exato que falta, a fila segue nas independentes, e o item vai para `pendencias-humanas.md` e para o inbox. Ao fim, **uma lista só** de pendências. Se todas as etapas restantes dependem da bloqueada, o motor para com "preciso de X" (não há como contornar sem a pessoa).

## Premissas (a confirmar no Gate 1)
- O pipeline precisa estar ativo (`forge.json`). Sem ele o comando orienta `forge pipeline init`.
- Etapa que **não gera PR** (ex.: produzir vídeos) é concluída pelo Builder com `forge queue complete` e revisada pelo Reviewer lendo a spec e os arquivos produzidos; o ponto exato fica para validar na implementação (ver risco).
- Em máquina sem tmux, vale o worker headless (contexto novo por job); o `/clear` do modo live só existe no tmux.

## Modelo do módulo
**Enxuto** (ARQ-11): `internal/pipeline/goal.go` reaproveita `bus`, `queue`, `Dispatch` e o archive; o subcomando fica em `cmd/forge-sdd/pipeline.go` (ARQ-01, ARQ-06: subcomando de `pipeline`, sem comando de nível raiz).

## Decisões do projeto aplicadas
ADR-0004 (pipeline multi-sessão), ADR-0009 (reset de contexto e auto-archive), ADR-0012 (Go aplica o veredito e atualiza a fila), ADR-0013 (Reviewer revisa a spec no modo avançado). Rules: `sdd-arquitetura`, `sdd-fluxo-e-portao`, `sdd-codigo-e-testes`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Enfileirar jobs de uma iniciativa | `pipeline.Dispatch` (`internal/pipeline/dispatch.go`) | Gate pré-build e um job por iniciativa a partir de `features/` | **estender** (função irmã `EnqueueGoalSteps`) que gera um job por etapa | O gate exige card e seções de feature; etapa de objetivo precisa de gate próprio |
| Ordem entre etapas | `forge bus plan`, `queue.EligibleTasks` | DAG por `**Depende de:**` | **reusar** | Sem ordenação nova |
| Revisão da spec | `spec_review`, `ApplySpecVerdict` (feat-88-04) | Existe na base | **reusar** | Portão da spec sem humano |
| Limpeza de contexto | `deliverContextReset`, `config.Context` | Só nas sessões do pipeline | **reusar** | O Go faz, não o modelo |
| Archive | `pipeline.CheckAndAutoArchive`, Archivist | Dispara com todas as tasks `approved` | **reusar** | Depende do fix-84-03 (já na base) |
| Parada | `pipeline.StopFile`, `HaltReason` | Existe na base | **reusar** | Sem mecanismo novo |
| Bloqueio sem travar a fila | `queue.BlockTasks` | Existe na base | **reusar** | Impasse não para tudo |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`FLX-01`, `FLX-06`, `ARQ-01`, `ARQ-06`, `DUP-01`, `SEG-01`, `SEG-03`, `SEG-06`, `COD-01`, `COD-03`, `TST-01`.
Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-fluxo-e-portao.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
0. Pré-voo: `requisitos.json` (esquema validado, SEG-01), `forge pipeline toca-ai --preflight` e `forge pipeline consent <id>` (subcomandos de `pipeline`), com `consents.json` gravado de forma atômica e sem segredos (SEG-04, SEG-05). Comandos de `requisitos.json` entram como argv explícito, nunca shell (SEG-03).
1. `internal/pipeline/goal.go`: lê `plan.md`, valida as specs das etapas (existem, `**Depende de:**` sem ciclo, IDs validados — SEG-01) e enfileira um job por etapa, idempotente.
2. Portão: `spec_review` no modo avançado antes das etapas; sem modo avançado, para e instrui.
3. Subcomando `forge pipeline toca-ai <iniciativa> [--dry-run]` em `cmd/forge-sdd/pipeline.go`; `--dry-run` não grava nada.
4. Testes: DAG respeitado, reenvio sem duplicar, bloqueio no portão da spec, `--dry-run` sem efeito, ciclo de dependência rejeitado.
5. Spike curto, registrado no PR: confirmar a conclusão e a revisão de uma etapa **sem PR**.

## Fora do escopo
Gerar as specs (feat-89-02); novo mecanismo de limpeza, archive ou parada; mudar o modo avançado.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/pipeline -run 'TestGoal'
go test -v -race ./cmd/forge-sdd -run 'TestPipelineTocaAi'
go vet ./... && go test ./...
forge gate
```
- Plano com 3 etapas e uma dependência vira 3 jobs, na ordem do DAG.
- Rodar duas vezes não duplica jobs.
- No modo avançado, nenhuma etapa é enfileirada antes do `spec_review` aprovado.
- `--dry-run` não grava arquivo nem job.
- Ciclo em `**Depende de:**` é rejeitado com a lista das etapas.
- Pré-voo com ferramenta ausente, arquivo ausente, comando fora da permissão ou consentimento não dado lista **todas** as pendências de uma vez e não enfileira nada.
- Consentimento só é gravado por resposta humana; um job nunca o grava. Etapa sem consentimento gravado não inicia.
- Etapa que bate num imprevisto vira `blocked` com o item exato, as independentes continuam e `pendencias-humanas.md` é uma lista única.
```bash
go test -v -race ./internal/pipeline -run 'TestPreflightGoal|TestConsent|TestPendencias'
```

## Riscos
| Risco | Mitigação |
|---|---|
| O agente pedir interação no meio (caso `marketing`) | Pré-voo único, consentimentos gravados, comando executado pelo próprio agente; imprevisto vira `blocked` + lista única |
| Agente "confirmar" consentimento pela pessoa | Só `forge pipeline consent` (resposta humana) grava; etapa sem registro não inicia |
| `pr_review` espera um PR e a etapa não tem | Spike (escopo 5); se preciso, a revisão de etapa sem PR lê os arquivos produzidos |
| Objetivo sem código passando pelo gate de spec de código | Gate próprio da etapa em `goal.go`, sem exigir card nem seções de feature |
| Sem tmux, sem `/clear` | O worker headless já abre contexto novo por job |
