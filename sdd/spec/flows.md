# Fluxos Operacionais — forge-sdd (v4.9.1-beta + iniciativas 89 e 90, na feat/88)

## 1. Pipeline Multi-Sessão: da ideia ao PR aprovado (ADR-0004)

```mermaid
sequenceDiagram
    autonumber
    actor H as Humano
    participant D as Sessão Discovery
    participant BUS as Barramento (.forge/runtime)
    participant B as Sessão Builder (worker)
    participant GA as Camada A (forge gate)
    participant GO as Go (forge pipeline pr)
    participant R as Sessão Reviewer (worker, só leitura)
    participant Q as queue.json da iniciativa

    H->>D: /discovery e /split-features (aprova a spec, FLX-01)
    D->>BUS: forge dispatch: gate pré-build + job feature_build (Initiative + TaskIDs)
    B->>BUS: claim do job, constrói em ondas paralelas (depends_on)
    B->>GO: commit feito, forge pipeline pr (Camada A, push, PR na base da iniciativa)
    GO->>Q: registra o PR, tasks do job -> review_pending
    B->>BUS: complete + enfileira job pr_review
    R->>GA: Camada A determinística (P-xx)
    R->>R: Camada B semântica: diff x spec x ADRs x rules
    R->>BUS: forge bus review --verdict-file (veredito JSON)
    alt conforme
        BUS->>Q: tasks do job -> approved (ADR-0012)
        BUS->>D: notificação "PR pronto" (inbox, toast, e-mail, webhook)
        BUS->>D: última task approved: aviso "pronta para arquivar" (só a fila do /toca-ai arquiva sozinha)
        H->>D: aprovação final do PR
    else violação (tentativas restantes)
        BUS->>B: job pr_rework com feedback e tentativa + 1
    else violação no limite ou dúvida
        BUS->>D: escalação para a pessoa
    end
```

O Reviewer não escreve no repositório (`SEG-06`). Quem guarda o veredito em `sdd/discovery/<id>/verdicts/<job>.json` e atualiza a fila é o Go, dentro de `ApplyVerdict`. O push e o PR também são do Go (`forge pipeline pr`, ADR-0014): o Builder só faz o commit. Quem abre o PR à mão usa `forge pipeline pr-opened`.

A base do PR é a da iniciativa: `queue.ResolveBase` usa, nesta ordem, a flag `--base`, o `base_branch` da fila, a origem da branch no git e o `base_branch` do `forge.json`. `forge dispatch <iniciativa> --fix fix-NN-NN` leva um fix pelo mesmo caminho, com a base na branch da iniciativa.

---

## 2. Estados do job no barramento (`.forge/runtime/<papel>/<estado>/`)

```mermaid
stateDiagram-v2
    [*] --> pending: bus.Enqueue (feature_build, pr_rework, pr_review, spec_review)
    pending --> processing: bus.Claim (lease com posse do worker)
    processing --> processing: bus.Renew / UpdatePR
    processing --> done: bus.Complete
    processing --> failed: bus.Fail
    processing --> pending: bus.Recover / Requeue (lease vencido ou limite de uso)
    done --> [*]
```

A Discovery usa os estados `inbox` e `archive` para as notificações.

---

## 3. Estados das tasks da fila (`sdd/discovery/<id>/queue.json`)

```mermaid
stateDiagram-v2
    [*] --> queued: forge queue enqueue
    queued --> building: forge queue next (modo manual)
    building --> review_pending: forge queue complete
    queued --> review_pending: pr-opened do Builder (modo pipeline)
    review_pending --> approved: veredito conforme (queue.Review ou ApplyVerdict)
    queued --> blocked: spec escalada ou no limite (modo avançado) ou pendência humana do /toca-ai
    review_pending --> changes_requested: veredito violacao
    changes_requested --> review_pending: PR do rework aberto
    changes_requested --> approved: rework aprovado
    approved --> [*]
```

No modo pipeline o job trata um bloco de tasks: `SetTaskStatuses` aplica a transição a todas as tasks de `job.TaskIDs`, na iniciativa `job.Initiative`. Uma task `approved` nunca regride. Uma task `blocked` guarda o motivo e só sai dela por decisão da pessoa; as tasks independentes seguem (`BlockTasks`).

---

## 4. Tokens por janela e reset suave de contexto (ADR-0009)

```mermaid
flowchart LR
    M["monitor.TokenMetricsPerRole<br/>tokens de Discovery, Builder e Reviewer"] --> BDG["badge no tmux (status-right)<br/>para decidir a hora de limpar"]
    JOB["novo job entregue ao painel vivo"] --> CM{"live.context = clear?<br/>(padrão)"}
    CM -- sim --> RST["deliverContextReset: profile.ResetCommand<br/>(/clear ou /new) no mesmo painel"]
    CM -- continuous --> KEEP["mantém o contexto contínuo"]
    RST --> HC["handoff: cartão com task, branch e onde parou"]
    HC --> SEND["prompt do job"]
    KEEP --> SEND
    APR["aprovação do último job"] --> ARQ["aviso: pronta para arquivar<br/>(só o /toca-ai arquiva sozinho)"]
```

O reset acontece **por job**, na entrega, e nunca fecha o painel do tmux. O modo é `clear` (padrão) ou `continuous` (feat 77). Perfis sem `reset_command` mantêm o contexto contínuo e avisam no terminal.

---

## 5. Acompanhamento: painel vivo e fila com drill-down (ADR-0011)

```mermaid
flowchart LR
    W["workers (Builder e Reviewer)"] --> CF["monitor.ChatFeedFormatter<br/>stream narrativo em linhas de chat"]
    W --> ACT["queue activity.json"]
    ACT --> LIVE["forge status --live<br/>pipeline.FormatRichDashboard"]
    Q["queue.json"] --> LIVE
    Q --> DR["survey.QueueDrilldownTUI<br/>abas Na Fila e Concluídas"]
    DR --> DET["cartão de detalhe da task<br/>arquivos, laudo do Reviewer, tempo"]
```

---

## 6. Ciclo de Vida de Auto-Skills (`DUP-05` / `GOV-03`)

```mermaid
flowchart LR
    E["3+ ocorrências do padrão<br/>(Regra de Três DUP-05)"] --> P["forge skills propose<br/>.agents/skills/slug/SKILL.md<br/>(status: candidate)"]
    P --> R["aprovação no fechamento de onda<br/>(GOV-03)"]
    R --> A["forge skills approve<br/>(status: active + sdd/skills/index.md)"]
    A --> S["forge skills sync<br/>propaga para .claude/skills/ e .gemini/skills/"]
```

---

## 7. Modo avançado: o Reviewer revisa a spec (ADR-0013)

```mermaid
sequenceDiagram
    autonumber
    participant D as Discovery
    participant GO as Go (forge dispatch)
    participant R as Reviewer (só leitura)
    participant Q as queue.json
    D->>GO: spec pronta (autonomy: advanced)
    GO->>GO: gate pré-build (Camada A) sem pendência
    GO->>R: job spec_review
    R-->>GO: parecer (conforme, violacao ou duvida)
    alt conforme
        GO->>GO: grava review-spec.md e enfileira o feature_build
    else violacao e tentativas restantes
        GO->>D: feedback pelo inbox (ajustar a spec)
    else duvida ou limite de tentativas
        GO->>Q: tasks blocked com o motivo
        GO->>D: escalação para a pessoa
    end
```

O modo exige `permissions: full` e `isolated_confirmed`. Os tetos de tokens, tempo e jobs por dia (`safeguards`) são opcionais e nascem desligados; `.forge/runtime/STOP` para a esteira de forma limpa.

---

## 8. `/toca-ai`: pré-voo e uma etapa por job (iniciativa 89)

```mermaid
flowchart LR
    L["Líder: plan.md, uma spec por etapa<br/>e requisitos.json"] --> PF["forge pipeline toca-ai --preflight<br/>ferramentas, arquivos, comandos e consentimentos"]
    PF -->|pendência| H["pessoa responde uma vez<br/>forge pipeline consent"]
    H --> PF
    PF -->|sem pendência| SG{"spec aprovada?"}
    SG -->|"modo avançado: Reviewer; normal: Gate 1 humano"| EN["EnqueueGoalSteps:<br/>um job por etapa, na ordem do DAG"]
    EN --> B["Builder executa cada etapa"]
    B --> AP["última etapa approved"]
    AP --> AR["archive automático (origem: toca-ai),<br/>se a branch atual for a da iniciativa"]
```

O consentimento só é gravado em terminal interativo (`.forge/runtime/consents.json`): o agente nunca o concede. Reenviar o comando não duplica jobs. Cada etapa sai com a branch `feat/<iniciativa>-<etapa>`. O archive automático só ocorre com a branch da iniciativa em uso (`branchGuard`) e `layout.FromBranch` não reconhece a branch de etapa como a da iniciativa. Limitação conhecida: `ApplyVerdict` descarta o resultado de `CheckAndAutoArchive`, então a recusa é silenciosa, sem aviso no inbox.

---

## 9. Uso real por job e relatório (ADR-0014)

```mermaid
flowchart LR
    LOG["log do job (stream-json)"] --> RES["evento result:<br/>tokens, cache, turnos, duração, modelo, custo de tabela"]
    RES --> REC["monitor.ReadJobUsage<br/>grava sdd/.metrics (medicao: real)"]
    LOG -->|"sem result"| EST["EstimateJobUsage<br/>(medicao: estimado)"]
    EST --> REC
    REC --> REP["forge report --initiative ID --jobs<br/>real e estimado em colunas separadas"]
```

Só o `claude` informa o uso real; nas outras CLIs o registro segue estimado. O custo é preço de tabela da API, útil para comparar jobs.
