# Plano de Implementação — Iniciativa 84

Plano estruturado de baixo para cima, dividido em ondas com dependências explícitas (limite máximo de 7 tasks por onda).

---

## Onda 1: Fundação & Stream Narrativo nos Workers

### Task 84-01: Formatador de Stream Narrativo para Workers (`ChatFeedFormatter`)
- **Escopo:** Implementar `internal/monitor/chat_feed.go` com estrutura de eventos e gerador de balões/cards de etapas com status humano (💭 Analisando, 🔨 Editando, 🧪 Testando, ✅ Concluído, ❌ Erro). Incluir suporte a cores ANSI suaves e sanitização ASCII.
- **Rules aplicáveis:** `ARQ-01`, `ARQ-04`, `COD-01`, `COD-03`, `DUP-01`, `TST-03`.
- **Dependências:** Nenhuma.
- **Critério Executável:** `go test -v ./internal/monitor -run TestChatFeed`

### Task 84-02: Integração do Stream Narrativo no Ciclo de Vida do Worker
- **Escopo:** Atualizar `cmd/forge-sdd/worker.go` para emitir eventos de ciclo de vida através do `ChatFeedFormatter` durante o polling e processamento de jobs, substituindo a mensagem de heartbeat monótona (`trabalhando há XmYs • Z arquivos alterados`).
- **Rules aplicáveis:** `ARQ-01`, `ARQ-06`, `COD-01`, `COD-03`, `DUP-01`.
- **Dependências:** Task 84-01.
- **Critério Executável:** `go test -v ./cmd/forge-sdd -run TestWorker_ChatFeed`

---

## Onda 2: TUI da Fila Interativa com Drill-Down

### Task 84-03: Componente de Lista de Tarefas com Paginação/Scroll (`QueueDrilldownTUI`)
- **Escopo:** Implementar `internal/survey/queue_drilldown.go` utilizando o framework `huh`, carregando todas as tarefas de `queue.json` sem truncamento e permitindo navegação fluida com setas `↑` e `↓`.
- **Rules aplicáveis:** `ARQ-01`, `ARQ-04`, `COD-01`, `DUP-01`, `TST-03`.
- **Dependências:** Nenhuma (paralelizável com Onda 1).
- **Critério Executável:** `go test -v ./internal/survey -run TestQueueDrilldown_List`

### Task 84-04: Modal de Inspeção e Detalhes da Tarefa Selecionada
- **Escopo:** Implementar a visualização detalhada ao teclar `Enter` sobre uma tarefa (`internal/survey/task_detail.go`), exibindo arquivos modificados, duração do build, commit gerado e parecer do reviewer, retornando à lista com `Esc` ou `q`.
- **Rules aplicáveis:** `ARQ-01`, `ARQ-04`, `COD-01`, `COD-03`, `TST-03`.
- **Dependências:** Task 84-03.
- **Critério Executável:** `go test -v ./internal/survey -run TestQueueDrilldown_Inspect`

---

## Onda 3: Integração CLI, Menu e Validação Final

### Task 84-05: Integração de Comandos CLI e Menu Principal
- **Escopo:** Adicionar a flag `-i/--interactive` no comando `status` (`cmd/forge-sdd/status.go`) e incluir opção navegável no menu interativo `forge` (`cmd/forge-sdd/menu.go`), conectando a TUI de drill-down.
- **Rules aplicáveis:** `ARQ-01`, `ARQ-06`, `COD-01`, `DUP-01`.
- **Dependências:** Tasks 84-02 e 84-04.
- **Critério Executável:** `go test -v ./cmd/forge-sdd -run TestCLI_StatusInteractive`

### Task 84-06: Documentação, Golden Tests e Validação de Regressão
- **Escopo:** Atualizar templates do `LEIA-ME.md` e testdata golden com o novo modo interativo e exemplos visuais do stream narrativo; rodar verificação estática global.
- **Rules aplicáveis:** `COD-02`, `COD-04`, `TST-01`, `TST-05`.
- **Dependências:** Task 84-05.
- **Critério Executável:** `go build ./... && go vet ./... && go test -v -race ./...`
