# Critérios de Aceitação e Arquitetura — Iniciativa 84

## Modelo do Módulo (ARQ-11)
- **Modelo:** `enxuto`.
- **Justificativa:** Extensão de apresentação visual e componentes de terminal em pacotes existentes (`internal/monitor`, `internal/survey` e `cmd/forge-sdd`), sem necessidade de novos daemons em background, persistência de banco de dados ou novos subsistemas de rede (`ARQ-11`).

## Decisões do Projeto Aplicadas
- `ADR-0004`: Pipeline Multi-Sessão (Discovery, Builder, Reviewer). A apresentação dos workers respeita a separação física de papéis e estados do pipeline.
- `ADR-0005`: Menu Navegável e Agente por Sessão. Acesso ao painel interativo pelo menu principal do CLI.
- `ADR-0006`: Visibilidade e tmux. Preservação do layout de 3 painéis divididos no tmux (macOS/Linux/WSL) e no Windows Terminal (Windows nativo).
- `ADR-0008`: Sessões interativas e despachante raw. O stream narrativo dos workers consome os marcos de execução sem interferir no buffer raw de entrega de comandos.
- `ARQ-01`, `ARQ-04`, `ARQ-06`, `ARQ-10`: Separação estrita: adaptadores CLI em `cmd/forge-sdd`, lógica de negócio e renderização em `internal/`, sem vazamento de domínio e sem dependências externas adicionais de runtime.
- `DUP-01`: Um conceito, uma casa: o formatador do feed conversacional reside em `internal/monitor/chat_feed.go` e a TUI de drill-down reside em `internal/survey/queue_drilldown.go`.
- `SEG-03`: Manipulação e renderização puramente in-memory, sem injeção de comandos de shell.

## O Que Já Existe (DUP-02)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Painel ao vivo de status | `internal/monitor/status_live.go` | `RenderLiveStatusLoop`, captura de processos e tokens | estender | Adicionar modo interativo com drill-down e suporte ao feed conversacional unificado. |
| Componentes interativos de terminal | `internal/survey/survey.go` | Wrappers de Charm `huh` para seleção e confirmação | estender | Implementar o componente `QueueDrilldownTUI` aproveitando a stack `huh` já integrada ao projeto. |
| Monitoramento de progresso do worker | `cmd/forge-sdd/worker.go` | Loop do worker com heartbeat a cada 5 segundos | estender | Substituir o log periódico monótono por emissão de cards de etapas conversacionais com status humano. |
| Consulta e estado da fila | `internal/queue/queue.go` | `ListTasks`, `LoadQueue` e inspeção de status | reusar | A TUI de drill-down consome diretamente as estruturas existentes de `queue.json` sem alterar formato. |
| Comandos de status e queue na CLI | `cmd/forge-sdd/status.go`, `cmd/forge-sdd/queue.go` | Comandos cobra para exibição estática e live | estender | Adicionar flag `-i/--interactive` para disparar a visualização navegável com drill-down. |

## Dependências Novas (SEG-07)
- **Nenhuma dependência externa nova.**
- Todo o desenvolvimento reutiliza a stdlib do Go (`fmt`, `strings`, `time`, `os`, `sync`) e o framework de formulários interativos `github.com/charmbracelet/huh` v0.3 já presente no `go.mod` (`ARQ-04`).

## Diagramas C4

### Nível 1: Diagrama de Contexto
```mermaid
flowchart TD
    User["Desenvolvedor / Líder Técnico"]
    
    subgraph ForgeSDD["Forge SDD CLI (v4.7.0)"]
        DiscoverySession["Sessão Discovery (Chat Interativo IA)"]
        BuilderSession["Sessão Builder (Stream Narrativo Chat)"]
        ReviewerSession["Sessão Reviewer (Cards de Governança)"]
        InteractivePanel["Painel Interativo de Fila (Drill-Down TUI)"]
    end
    
    User -->|Interage diretamente| DiscoverySession
    User -->|Acompanha etapas humanas| BuilderSession
    User -->|Acompanha laudo e veredito| ReviewerSession
    User -->|Navega com setas e Enter| InteractivePanel
    
    DiscoverySession -.->|Enfileira tasks| InteractivePanel
    BuilderSession -.->|Atualiza status da task| InteractivePanel
    ReviewerSession -.->|Publica veredito| InteractivePanel
```

### Nível 2: Diagrama de Contêineres
```mermaid
flowchart LR
    subgraph TerminalMultiplexer["Ambiente de Terminal (Tmux / Windows Terminal)"]
        subgraph Pane1["Painel 0: forge-discovery"]
            AgyAgent["CLI Interativa do Agente (agy / claude)"]
        end
        subgraph Pane2["Painel 1: forge-builder"]
            WorkerBuilder["forge worker --role builder"]
            ChatFeedBuilder["ChatFeedFormatter (Balões & Etapas)"]
            WorkerBuilder --> ChatFeedBuilder
        end
        subgraph Pane3["Painel 2: forge-reviewer"]
            WorkerReviewer["forge worker --role reviewer"]
            ChatFeedReviewer["ChatFeedFormatter (Laudos & Etapas)"]
            WorkerReviewer --> ChatFeedReviewer
        end
    end

    subgraph InternalCore["internal/ (Núcleo Go)"]
        QueueStore["queue.json (Fila de Tarefas)"]
        QueueDrilldown["QueueDrilldownTUI (internal/survey)"]
        MetricsStore["sdd/.metrics/ (Telemetria)"]
    end

    ChatFeedBuilder -->|Lê/Grava Estado| QueueStore
    ChatFeedReviewer -->|Grava Veredito| QueueStore
    QueueDrilldown -->|Carrega Lista Completa| QueueStore
    QueueDrilldown -->|Carrega Histórico e Tokens| MetricsStore
```

### Diagrama de Sequência: Stream Narrativo do Worker e Drill-Down
```mermaid
sequenceDiagram
    autonumber
    actor Dev as Usuário
    participant Worker as Worker (Builder/Reviewer)
    participant Feed as ChatFeedFormatter
    participant Queue as QueueStore (queue.json)
    participant TUI as QueueDrilldownTUI

    Note over Worker,Feed: Ciclo de Execução Humanizado no Worker
    Worker->>Feed: StartTask("fix-abcc-01", "Telas no frontend-admin")
    Feed-->>Dev: 💭 "Analisando requisitos e critérios da task fix-abcc-01..."
    
    Worker->>Feed: StepAction("edit", "shared/ui/src/components/BottomNav.tsx")
    Feed-->>Dev: 🔨 "Editando BottomNav.tsx e componentes de navegação..."
    
    Worker->>Feed: StepAction("test", "npm run test")
    Feed-->>Dev: 🧪 "Executando suite de testes automatizados (vitest)..."
    
    Worker->>Feed: CompleteTask(status="ok", duration="1m15s")
    Feed-->>Dev: ✅ "Task concluída com sucesso! Handoff despachado para o Reviewer."
    Worker->>Queue: Atualiza status da task para review_pending

    Note over Dev,TUI: Inspeção da Fila com Drill-Down
    Dev->>TUI: Executa 'forge status -i'
    TUI->>Queue: Carrega todas as tasks (sem truncamento)
    TUI-->>Dev: Renderiza lista com badges [✓] Concluído, [▶] Em andamento
    Dev->>TUI: Pressiona 'Enter' sobre fix-abcc-01
    TUI-->>Dev: Abre modal com arquivos alterados, tempo de build e commit
    Dev->>TUI: Pressiona 'Esc' / 'q'
    TUI-->>Dev: Retorna para a lista de tarefas
```

## Critérios de Aceitação Executáveis

```bash
go test -v -race ./internal/monitor/... ./internal/survey/... ./cmd/forge-sdd/...
```

1. **TestChatFeed_RenderStepCards:**
   - O `ChatFeedFormatter` recebe eventos de início, ação em arquivo, execução de teste e conclusão.
   - A saída gerada contém formatação estruturada com badges, quebras de linha adequadas, horário relativo e texto humanizado (ex: `💭 Analisando...`, `🔨 Editando...`, `🧪 Testando...`, `✅ Concluído...`), sem dumps brutos.
2. **TestChatFeed_ErrorHighlighting:**
   - Em caso de falha de teste ou compilação, o `ChatFeedFormatter` destaca o erro em bloco `❌ Falha na etapa [teste]`, incluindo os últimos 5 logs relevantes de erro sem poluir o buffer.
3. **TestQueueDrilldown_ListAllTasksWithoutTruncation:**
   - Em uma fila com 25 tarefas (como a evidenciada no projeto `barbearia-app`), o componente `QueueDrilldownTUI` carrega e exibe todas as 25 tarefas navegáveis, sem omitir tarefas por reticências.
4. **TestQueueDrilldown_CardDetailInspection:**
   - Ao selecionar uma tarefa concluída ou com falha, o método `InspectTaskDetails(taskID)` retorna a estrutura completa de inspeção (tempo total decorrido, lista de arquivos alterados, commit gerado e parecer do reviewer).
5. **TestCLI_StatusInteractiveFlag:**
   - `forge status --interactive` (ou `-i`) inicializa a visualização navegável interativa em vez do loop de texto puro do modo padrão.
6. **Compilação e Higiene:**
   - `go build ./... && go vet ./...`: 100% verde sem erros, sem warnings e com zero dependências externas novas.
