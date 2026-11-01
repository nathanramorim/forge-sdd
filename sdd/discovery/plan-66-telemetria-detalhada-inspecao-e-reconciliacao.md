# Plano de Discovery 66 — Telemetria Detalhada, Inspeção Passo a Passo e Reconciliação de Atividades Órfãs

Roadmap preliminar, arquitetura de fatiamento e mapeamento de governança para execução da iniciativa `feat-66-telemetria-detalhada-inspecao-e-reconciliacao`.

A execução será agrupada na subpasta única `sdd/features/feat-66-telemetria-detalhada-inspecao-e-reconciliacao/` (Regra 13), utilizando uma **única branch de trabalho** (`feat/66-telemetria-detalhada-inspecao-e-reconciliacao`, conforme Regra 15 e `FLX-02`). O fatiamento respeita a **Rule of 7** (4 subtarefas de baixo para cima com escopos claros, atômicos e não sobrepostos).

---

## 1. Quebra de Tarefas / Subfeatures (`P / M / G`)

### Fase 1 — Núcleo do Modelo de Dados e Reconciliador de Atividades em Go stdlib

- [ ] **`feat-66-01` — Extensão do Modelo de Telemetria e Reconciliador de Atividades Órfãs em Go (`internal/queue/activity.go`)** · **Tamanho: `M`**
  - **Depende de:** — (reutiliza `internal/queue` existente)
  - **Escopo:**
    1. Estender a struct `ActivityItem` em `internal/queue/activity.go` com os campos enriquecidos: `ID`, `Command`, `Args`, `FilePath`, `Output`, `Error`, `DurationMs`, `CompletedAt` e `MascotState`, preservando retrocompatibilidade estrita via tags JSON `omitempty` (`ARQ-06`, `DUP-02`).
    2. Implementar a função `RecordActivityDetailed(dir string, item ActivityItem) error` com aplicação de truncamento seguro no campo `Output` (máx. 8192 bytes) para prevenção de inchaço do arquivo JSON.
    3. Atualizar `RecordActivity(...)` para que gere automaticamente um `ID` determinístico único e invoque a reconciliação automática de atividades anteriores da mesma estação que estejam pendentes em `running`.
    4. Implementar a função `ReconcileStationActivities(dir, station, newStatus string) error` para fechar e calcular durações decorridas de micro-ações.
    5. Implementar a função `SanitizeOrphanActivities(dir string, maxAge time.Duration) (int, error)` para varredura e saneamento de registros órfãos expirados.
    6. Criar suite de testes em `internal/queue/activity_test.go` validando persistência, truncamento de output, reconciliação de múltiplos itens e garantia de concorrência com detecção de corrida (`go test -race`) (`TST-01`).
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestActivity" ./internal/queue/... && go vet ./...
    ```

---

### Fase 2 — Integração do Reconciliador no Aggregator do Studio, SessionPool e CLI

- [ ] **`feat-66-02` — Integração do Reconciliador no Aggregator do Studio, Transições de Standby e CLI (`internal/studio/state.go`, `cmd/forge-sdd/queue.go`)** · **Tamanho: `M`**
  - **Depende de:** `feat-66-01`
  - **Escopo:**
    1. Integrar a chamada a `SanitizeOrphanActivities` na inicialização do servidor HTTP do Studio (`internal/studio/server.go`) e no construtor `NewAggregator`.
    2. No método `Aggregator.GetSnapshot()` em `internal/studio/state.go`, validar que se uma estação possui `MascotState == "idle"` ou `SessionState == "standby"`, suas atividades recentes não exibem `running` órfão, disparando reconciliação determinística sob demanda.
    3. Ajustar `internal/acp/attach.go` para que a emissão de logs via `AttachBridge` reconcilie ou vincule de forma limpa micro-ações sem criar registros órfãos soltos.
    4. Expandir o subcomando CLI `forge-sdd queue activity` em `cmd/forge-sdd/queue.go` com suporte a `--sanitize` para saneamento explícito via terminal e exibição das novas colunas de telemetria.
    5. Testes unitários em `internal/studio/server_test.go` e `internal/studio/studio_test.go` validando que o snapshot não contém atividades `running` quando os agentes estão ociosos.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestAggregator|TestStudio" ./internal/studio/... && go vet ./...
    ```

---

### Fase 3 — UI do Forge Studio: Modal/Drawer de Inspeção de Telemetria Interativo

- [ ] **`feat-66-03` — Interface Web do Forge Studio: Modal de Inspeção Passo a Passo e Linhas Clicáveis (`internal/studio/assets/`)** · **Tamanho: `M`**
  - **Depende de:** `feat-66-01`, `feat-66-02`
  - **Escopo:**
    1. Adicionar o container modal `#activityModal` em `internal/studio/assets/index.html` com cabeçalho de mascote, detalhes de execução, container de terminal escuro para `stdout`/`stderr` e rodapé de navegação.
    2. Atualizar a função `renderConsoleLogs()` em `internal/studio/assets/studio.js` para renderizar as linhas com a classe clicável `.log-row-clickable`, atributos de acessibilidade (`tabindex="0"`, `role="button"`) e evento `onclick="openActivityModal('{act.id}')"`.
    3. Implementar as funções de controle em `studio.js`:
       - `openActivityModal(actId)`: localiza a atividade por ID, popula campos de comando, argumentos, caminho, output e mascote, e exibe o modal com animação suave.
       - `closeActivityModal()`: fecha o modal e restaura o foco da interface.
       - `navigateActivity(delta)`: avança ou retrocede na lista sequencial de atividades.
       - `copyCommand()` e `copyFilePath()`: copia para o clipboard com feedback visual temporário ("Copiado!").
    4. Adicionar suporte a eventos de teclado globais no modal: tecla `Escape` para fechar e setas `ArrowLeft` / `ArrowRight` para navegação contínua entre micro-ações.
    5. Adicionar regras CSS em `internal/studio/assets/studio.css` para o modal responsivo, highlight de hover nas linhas, layout estilo terminal dark para saída de comando e botões de cópia.
    6. Atualizar testes de integridade de assets em `internal/studio/assets_test.go`.
  - **Critério Executável:**
    ```bash
    go test -v -race -run "TestAssets" ./internal/studio/... && go vet ./...
    ```

---

### Fase 4 — Integração E2E, Sanitização da Base Legada e Validação de Portão

- [ ] **`feat-66-04` — Testes E2E, Saneamento do `activity.json` Legado e Validação Constitucional Completa** · **Tamanho: `M`**
  - **Depende de:** `feat-66-01`, `feat-66-02`, `feat-66-03`
  - **Escopo:**
    1. Aplicar a sanitização no arquivo de dados real do repositório `sdd/.queue/activity.json`, eliminando permanentemente qualquer atividade histórica órfã que estivesse marcada com `running`.
    2. Desenvolver teste E2E em `internal/studio/studio_test.go` simulando o ciclo completo: registro de ação com output e duração -> reconciliação no handoff -> transmissão via SSE -> inspeção no snapshot.
    3. Validar concorrência pesada com múltiplos goroutines inserindo e reconciliando atividades simultaneamente sob a flag `-race`.
    4. Garantir aprovação de 100% da suíte de testes e linters do projeto (`go test -race ./...` e `go vet ./...`).
    5. Atualizar documentação e histórico de releases em conformidade com `FLX-07`.
  - **Critério Executável:**
    ```bash
    go test -v -race ./... && go vet ./...
    ```

---

## 2. Matriz de Rastreabilidade Constitucional

| Subfeature | Regras Constitucionais Cobertas |
|---|---|
| `feat-66-01` | `ARQ-01`, `ARQ-03`, `ARQ-04`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-04`, `DUP-02` |
| `feat-66-02` | `ARQ-02`, `ARQ-06`, `SEG-01`, `SEG-06`, `FLX-04`, `TST-01`, `TST-05` |
| `feat-66-03` | `ARQ-03`, `COD-01`, `COD-05`, `SEG-01`, `TST-02`, `FLX-01` |
| `feat-66-04` | `ARQ-04`, `SEG-04`, `SEG-05`, `TST-01`, `TST-04`, `TST-05`, `FLX-02`, `FLX-07`, `GOV-03` |

---

## 3. Próximo Passo

> [!NOTE]
> Concluída a concepção da tríade de Discovery pela **Estação 02 (PM & Arquiteto)**, os artefatos estão submetidos à validação humana (Portão **`FLX-01`**).
> Após a aprovação do usuário, a esteira avança para a **Estação 03 (Planejador)** para enfileiramento das tasks detalhadas na fila `sdd/.queue/stations/planner.json`.
