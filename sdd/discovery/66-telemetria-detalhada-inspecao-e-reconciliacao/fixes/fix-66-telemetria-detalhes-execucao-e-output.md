# fix-66 — Enriquecimento de Detalhes de Execução, Comando e Saída no Modal de Telemetria

**Branch:** `fix/66-telemetria-detalhes-execucao-e-output`
**Fase:** fix-66
**Tamanho:** `M`
**Depende de:** `feat-66-04`
**Status:** `ready`

## Feature Correlacionada

- **Iniciativa de Origem:** Feature 66 — Telemetria Detalhada, Inspeção Passo a Passo, Reconciliação de Atividades Órfãs e Estrutura Unificada de Discovery (`sdd/discovery/66-telemetria-detalhada-inspecao-e-reconciliacao/`).
- **Arquivos Afetados de Origem:**
  - `internal/studio/server.go`: endpoints `handleTaskExecute` e `runAutoRunAsync`.
  - `internal/studio/assets/studio.js`: função `openActivityModal`.
  - `internal/studio/assets/index.html` e `studio.css`: layout e campos do modal.
- **Causa-Raiz do Defeito:**
  1. No backend (`internal/studio/server.go`), as execuções de tarefas e portões (como `go test -race ./...`, verificações de governança e handoffs) chamavam `queue.RecordActivity(...)` sem preencher `Command`, `FilePath`, `Output`, `DurationMs` e `CompletedAt`, ou apenas concatenavam uma string resumida no campo `details`. A saída de terminal gerada por `exec.Command` e a duração medida eram descartadas em vez de serem persistidas em `queue.RecordActivityDetailed`.
  2. No frontend (`internal/studio/assets/studio.js`), a função `openActivityModal` utilizava um fallback agressivo no bloco "Comando Executado": se `act.command` estivesse vazio, exibia `act.details || act.action`, fazendo com que textos descritivos como *"Auditando conformidade constitucional"* aparecessem erroneamente rotulados como o comando executado no terminal.
  3. Campos como "Arquivo Afetado", "Duração" e "Saída / Logs do Terminal" exibiam `--` e `(Nenhuma saída capturada no terminal)` porque o objeto retornado por `/api/state` não trazia esses metadados preenchidos.
- **Evidência do Defeito:**
  - Captura de tela do usuário exibindo o modal `#activityModal` com:
    - *Comando Executado:* "Auditando conformidade constitucional" (texto da ação, não comando)
    - *Arquivo Afetado:* `--`
    - *Duração:* `--`
    - *Término:* `--`
    - *Saída / Logs do Terminal:* `(Nenhuma saída capturada no terminal)`

---

## Modelo do módulo

- **Modelo:** `enxuto` em `internal/studio` e `internal/studio/assets/`.
- **Justificativa:** Correção cirúrgica de persistência de telemetria e renderização de dados no frontend sem introdução de novas camadas ou dependências externas (`ARQ-03`, `ARQ-04`).

---

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Modelo de Telemetria | `internal/queue/activity.go` | `ActivityItem`, `RecordActivityDetailed` | `reusar` | O modelo já suporta `Command`, `FilePath`, `DurationMs`, `Output`, `CompletedAt`. Basta preenchê-los. |
| Execução de Tarefas no Studio | `internal/studio/server.go` | `handleTaskExecute` e `runAutoRunAsync` | `estender` | Coletar `start := time.Now()`, `time.Since(start).Milliseconds()`, capturar saída completa e chamar `RecordActivityDetailed`. |
| Modal de Telemetria no Frontend | `internal/studio/assets/studio.js` | `openActivityModal()` | `estender` | Separar explicitamente comando shell de descrições conceituais, calcular duração entre timestamps se `duration_ms` vier zerado, e renderizar contexto rico para ferramentas de auditoria e governança. |
| Estrutura Visual | `internal/studio/assets/index.html` e `studio.css` | `#activityModal` | `estender` | Adicionar badges de contexto quando o evento for uma tool de auditoria, handoff ou portão em vez de comando CLI simples. |

---

## Dependências novas

- **Nenhuma dependência externa nova.** (`ARQ-03`, `ARQ-04`, `SEG-07`).

---

## Rules e IDs da Constituição Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-06`, `SEG-01`, `SEG-06`, `DUP-02`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-04`, `FLX-02`.

---

## Critérios de Aceitação Executáveis

```gherkin
Cenário: Execução de comando do Builder registra telemetria completa com stdout e duração
  Dado que o Studio dispara uma execução de task do Builder com comando "go test -race ./..."
  Quando a execução terminar
  Então a telemetria gravada em activity.json deve conter command="go test -race ./..."
  E deve conter output com a saída real do terminal (PASS ou FAIL)
  E deve conter duration_ms maior que zero
  E deve conter completed_at preenchido

Cenário: Execução de Auditoria da Governança registra arquivo auditado e achados
  Dado que a estação de Governança executa a auditoria de uma task
  Quando o parecer for emitido
  Então a telemetria gravada deve conter file_path apontando para o arquivo de spec/artefato da task
  E o output deve conter o resumo do parecer e o status constitucional
  E o modal deve exibir claramente o arquivo afetado e os detalhes de conformidade

Cenário: Frontend exibe metadados mesmo com registros parciais legados
  Dado uma atividade com timestamps de início e término mas duration_ms omitido
  Quando o usuário abre o modal de inspeção dessa atividade
  Então a duração deve ser calculada dinamicamente a partir dos timestamps
  E se command for omitido para ferramentas conceituais (ex: audit, handoff), o bloco deve indicar o tipo de operação claramente
```

---

## Plano de Implementação

1. **Backend (`internal/studio/server.go`):**
   - Em `handleTaskExecute`: substituir chamadas rasas de `RecordActivity` por `RecordActivityDetailed` populando `Command`, `FilePath`, `DurationMs`, `Output`, `CompletedAt`.
   - Em `runAutoRunAsync`: coletar o tempo real de execução de cada passo (`time.Since(start).Milliseconds()`), capturar saída do `go test` e do parecer de auditoria e persistir detalhadamente.
2. **Frontend (`internal/studio/assets/studio.js`):**
   - Ajustar `openActivityModal`:
     - Se `act.command` existir, exibir no terminal com cópia. Se não existir, verificar se é uma tool de sistema (`audit`, `handoff`, `spec_approve`), exibindo um painel contextual descritivo em vez de colocar o texto no terminal.
     - Resolver `file_path` caso venha vazio, inferindo de `act.action` ou do artefato correspondente.
     - Calcular duração se `duration_ms` for 0 mas houver `timestamp` e `completed_at`.
     - Exibir no bloco de saída o parecer ou logs capturados.
3. **Testes Automatizados:**
   - Adicionar casos de teste em `internal/studio/telemetry_reconcile_test.go` verificando que execuções reais persistem `Command`, `Output`, `DurationMs` e `FilePath`.
   - Validar com `go test -count=1 -race ./...` e `go vet ./...`.
4. **Recompilar binário e reiniciar Studio.**
