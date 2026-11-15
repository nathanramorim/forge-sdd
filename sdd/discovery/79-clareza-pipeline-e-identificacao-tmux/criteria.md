# Critérios Técnicos — Discovery 79 (Clareza no Pipeline, Identificação no Tmux, Paralelismo de Filas, Monitor de Processos e Menu)

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Toca a orquestração de processos do tmux (`internal/pipeline`), o motor de filas com grafo acíclico (`internal/queue`), o inspetor nativo de processos (`internal/monitor`) e a interface CLI (`cmd/forge-sdd/`). Exige testes unitários com mocks, runner simulado e zero efeitos colaterais (`ARQ-11`, `COD-01`, `SEG-01`).

## Decisões do projeto aplicadas

| Decisão | Origem | Impacto nesta iniciativa |
|---|---|---|
| ADR-0004 | Pipeline Multi-Sessão | A esteira física tem precedência sobre suposições de subagentes genéricos. |
| ADR-0005 | Menu Navegável e Agente por Sessão | O menu interativo simplifica o acesso às funções diárias, eliminando a hierarquia artificial "pipeline" vs "padrão". |
| ADR-0006 | Visibilidade tmux | Estende a visibilidade dos painéis para incluir a identidade visual do projeto, formato compacto de monitoramento e inspeção de processos. |
| ADR-0008 | Sessões com CLI Aberta (Modo Live) | Preserva a integridade dos buffers e das janelas de controle `forge-ctl`, garantindo que o status do projeto não interfira no envio de prompts. |
| D2 | Framework CLI: Cobra v1.8+ | Menu e subcomando `ps` integrados como comandos Cobra oficiais. |
| D8 | Idioma padrão: pt-BR | Textos de interface, opções de menu e mensagens dos agentes em pt-BR padrão. |

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Construção do plano tmux | `internal/pipeline/tmux.go` (`BuildTmuxPlan`) | Comandos `new-session`, `split-window`, `pane-border-status` | estender | Adicionar opções de sessão (`status-left`, `status-left-length`, `pane-border-format` com nome do projeto e janela `forge:<projeto>`). |
| Nome do projeto saneado | `internal/pipeline/tmux.go` (`TmuxSessionName`) | Sanitização de nome e cálculo de hash de diretório | reusar | A mesma base limpa de nome é usada na rotulagem visual e no filtro de processos. |
| Motor de fila | `internal/queue/queue.go` | `Queue`, `QueueItem`, `Enqueue`, `Next` | estender | Adicionar `DependsOn []string`, parser de spec e método `EligibleTasks` com DAG. |
| Inspeção de painéis | `internal/pipeline/tmux.go` | `TmuxRunner` | estender | Extrair PIDs dos painéis da sessão do projeto para o monitor `forge-sdd ps`. |
| Menu interativo | `cmd/forge-sdd/menu.go` | `runMenu`, submenus separados | unificar | Consolidar em menu único direto por ação, com cabeçalho de status do projeto. |

## Dependências novas
- **Nenhuma.** Permanece 100% aderente à stdlib Go, `cobra` e `huh`, além do utilitário `tmux` e comandos padrão de SO (`ps` / `pgrep`) (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- **ARQ-01 / ARQ-10:** Separação entre CLI (`cmd/forge-sdd`) e lógica de criação/execução (`internal/pipeline`, `internal/queue`, `internal/monitor`).
- **ARQ-03:** Templates de instruções embutidos via `embed.FS` atualizados mantendo sincronismo.
- **COD-01 / COD-05:** Responsabilidade única por função e idioma pt-BR padrão.
- **DUP-01 / DUP-02:** Um conceito, uma casa; reuso do saneador de nomes de projetos e status existente.
- **FLX-01 / GOV-01:** Concepção foca na Spec e na arquitetura, respeitando as decisões de pipeline.
- **SEG-01 / SEG-07:** Isolamento de processos; inspeção restrita exclusivamente ao diretório e sessão do projeto.

---

## Protótipos de Interface

### Protótipo 1: Monitor de Processos e Consumo (`forge-sdd ps`)

O comando `forge-sdd ps` lista exatamente o que está rodando atrelado ao projeto, consumo de máquina e tempo ativo:

```text
Processos Ativos — barbearia-app (Sessão tmux: forge-barbearia-app-a1b2c3)

PID    PAPEL / COMPONENTE            ESTADO      CPU %   MEM (RSS)   TEMPO ATIVO   COMANDO
─────────────────────────────────────────────────────────────────────────────────────────────
48120  tmux session (forge)          ativo       0.2%    18 MB       2h 15m        tmux new-session ...
48125  forge-discovery (gemini)      ocioso      0.0%    82 MB       2h 15m        agy -i
48130  forge-builder (claude)        executando  14.2%   145 MB      18m 32s       claude -n forge-builder ...
48201  └── subagente tests (go test) executando  45.0%   110 MB      42s           go test -v ./...
48135  forge-reviewer (gemini)       ocioso      0.0%    78 MB       2h 15m        agy -i
48142  forge-ctl (despachante)       ocioso      0.1%    14 MB       2h 15m        forge-sdd worker ...
─────────────────────────────────────────────────────────────────────────────────────────────
Total do Projeto: 6 processos ativos │ CPU: 59.5% │ Memória: 447 MB

[Dica: Use 'forge-sdd ps --kill-zombies' para limpar processos orfãos ou presos]
```

---

### Protótipo 2: Tela no tmux com Identificação, Terminal Livre e Mini-Monitor

```text
┌─ [ barbearia-app ❯ forge-discovery ] ────────────────┬─ [ barbearia-app ❯ terminal livre / dev ] ─────────┐
│                                                      │ $ git status                                       │
│ > Sessão forge-discovery                             │ On branch feat/79-clareza-pipeline                 │
│                                                      │ $ npm test                                         │
│ A esteira está em repouso passivo.                   │                                                    │
│ Pronto para conceber nova feature.                   │ (O terminal direito fica livre para você!)         │
│                                                      │                                                    │
│                                                      ├─ [ barbearia-app ❯ ⚙️  monitor da esteira ] ────────┤
│                                                      │  Builder:  [OCIOSO]  — pronto para receber jobs    │
│                                                      │  Reviewer: [OCIOSO]  — sem PRs pendentes           │
│                                                      │  Fila:     0 queued · 0 review_pending             │
│                                                      │  Consumo:  CPU 0.3% │ RAM 192 MB │ 4 processos     │
└──────────────────────────────────────────────────────┴────────────────────────────────────────────────────┘
 [ 📦 barbearia-app │ feat/12-auth ]  0:forge:barbearia-app*                14:25 | 3 sessões: Disc · Bld · Rev 
```

---

### Protótipo 3: Execução Paralela de Fila por DAG (Ondas)

Visualização no mini-monitor quando a fila libera tasks independentes simultaneamente:

```text
┌─ [ barbearia-app ❯ ⚙️  esteira: execução paralela por ondas ] ─────────────────────────────────────────────┐
│ ├── 🌊 Onda 1 (Em execução simultânea - 2/3 workers ocupados):                                             │
│ │   ├── 🚀 Task 79-01 (Builder - Gemini) [EXECUTANDO - 52s]                                               │
│ │   │   └── ✍️  Editando templates de prompt em internal/scaffold...                                      │
│ │   └── 🚀 Task 79-02 (Builder - Claude) [EXECUTANDO - 34s]                                               │
│ │       └── 🧪 Compilando internal/pipeline/tmux.go...                                                    │
│ └── ⏳ Onda 2 (Bloqueada até aprovação da Onda 1):                                                         │
│     ├── 🔒 Task 79-03 (Depende de 79-02)                                                                  │
│     └── 🔒 Task 79-04 (Depende de 79-02)                                                                  │
└───────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Critérios de Aceitação Executáveis

- **CA-01 (Identificação tmux nos testes):** `go test ./internal/pipeline -run TmuxPlanProjectIdentity` passa: o plano gerado por `BuildTmuxPlan` configura `status-left` com o nome do projeto saneado, `status-left-length 60`, `pane-border-format` contendo `<projeto> ❯ #{pane_title}` e nome da janela como `forge:<projeto>`.
- **CA-02 (Escopo estrito do tmux):** Todas as diretivas de customização de status e bordas em `internal/pipeline/tmux.go` usam o alvo exato da sessão (`-t =name`), sem aplicar opções globais (`-g`), preservando o ambiente global do usuário.
- **CA-03 (Menu unificado e direto):** `go test ./cmd/forge-sdd -run Menu` passa: o menu raiz não divide mais entre "pipeline" e "estrutura padrão", apresentando as opções principais de forma direta e mantendo a ação padrão de iniciar o pipeline quando instalado e pronto.
- **CA-04 (Prompts e regras Pipeline-First):** As menções a "pergunte se vai para um subagente" em `GEMINI.md`, `CLAUDE.md` e `.agents/skills/sdd-concepcao` são substituídas por diretrizes claras de inspeção da esteira (`forge-sdd status`) e priorização de handoff para sessões de Builder/Reviewer.
- **CA-05 (Paralelismo de fila por DAG):** `go test ./internal/queue -run TestQueueDAGParallelEligibility` passa: `QueueItem` persiste `depends_on`, identifica tasks sem bloqueio como elegíveis para build paralelo, respeita `max_parallel_tasks` e detecta ciclos como erro.
- **CA-06 (Monitor de processos do projeto):** `go test ./internal/monitor -run TestProjectProcessInspection` passa: mapeia os processos filhos e painéis da sessão do projeto, reporta PID, CPU, memória RSS e tempo decorrido, e `forge-sdd ps` exibe a tabela formatada.
- **CA-07 (Integridade da suíte):** `go build ./... && go vet ./... && go test -race ./...` e `forge-sdd gate` executam com 100% de sucesso.
