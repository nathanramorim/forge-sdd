# feat-84-07 — UX da Fila (Separação Ativas vs. Concluídas, Pipeline Discovery→Builder→Reviewer, Tokens por Tarefa/Etapa e Dicas de Atalhos no Tmux)

**Iniciativa:** `84-painel-interativo-chat-e-fila-drilldown`
**Branch:** `feat/84-painel-interativo-chat-e-fila-drilldown`
**Tamanho:** `M`
**Depende de:** `feat-84-03`, `feat-84-04`, `feat-84-05`, `fix-84-01`

## Objetivo
Refinar a experiência do painel interativo da fila (`forge queue -i`, `forge status -i` e popup `Ctrl+b q` no tmux) e da barra inferior do tmux sem impactar o funcionamento atual:
1. **Separação em 2 Abas na Fila (`Aba 1: ⏳ Na Fila / Em Andamento` vs. `Aba 2: ✅ Concluídas`):** cada aba lista exclusivamente seu conjunto de tarefas com opção nativa de alternância (`⇄ Ver Aba 2: Concluídas` / `⇄ Ver Aba 1: Na Fila`), sem colunas fixas de texto que estouram quando o `TaskID` é grande.
2. **Dicas dinâmicas de atalhos na barra inferior direita do tmux (`status-right`):** exibe dicas rotativas nativas (`💡 Dica: Ctrl+b q abre Fila`, `💡 Dica: Ctrl+b Setas navega · z zoom`, e guia imediato `⌨️ [q] Fila/Detalhes · [Setas] Painel · [z] Zoom` ao pressionar `Ctrl+b`) ao lado do resumo de tokens, usando `-t <session>` sem prefixo `=`.
3. **UX do card de detalhes da tarefa (`TaskDetailView`):**
   - Layout de blocos com borda apenas à esquerda (`╭───`, `│`, `├───`, `╰───`), **sem fechar caixa com `│` na direita**, eliminando qualquer quebra de alinhamento causada por largura de emojis/fontes ou redimensionamento de terminal.
   - Seção visual da esteira da tarefa (`[1. Discovery] ──► [2. Builder] ──► [3. Reviewer]`).
   - Contabilidade de tokens consumidos em cada etapa (`Discovery`, `Builder`, `Reviewer`) e **Total da Tarefa**, além de gravação automática da estimativa de tokens pelo worker ao concluir jobs.
   - Eliminação dos botões `"Voltar"` duplicados: em `ShowTaskDetail`, os dois botões agora têm ações distintas e rotuladas (`← Voltar à lista` para retornar à lista de tarefas e `✕ Fechar painel` para encerrar o modal direto); em `Pause` (`internal/survey/select.go`), o segundo botão idêntico `"Voltar ao menu"` foi removido, deixando apenas uma opção `"← Voltar ao menu"`.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Evolução coesa dos pacotes existentes `internal/survey`, `internal/monitor`, `internal/pipeline` e `cmd/forge-sdd`, sem novos pacotes ou concorrência adicional (`ARQ-11`).

## Decisões do projeto aplicadas
- **ADRs:** `ADR-0004` (Pipeline Multi-Sessão Discovery, Builder e Reviewer), `ADR-0006` (Visibilidade e ergonomia no tmux), `ADR-0008` (Sessões no tmux).
- **IDs da Constitution:** `ARQ-01`, `ARQ-04`, `ARQ-10`, `ARQ-11`, `COD-01`, `COD-02`, `COD-05`, `DUP-01`, `DUP-02`, `SEG-03`, `TST-03`.
- **Módulos afetados:** `internal/monitor/tokens.go`, `internal/survey/queue_drilldown.go`, `internal/survey/task_detail.go`, `internal/pipeline/tmux.go`, `cmd/forge-sdd/worker.go`.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Lista interativa da fila | `internal/survey/queue_drilldown.go` | `NewQueueDrilldownTUI`, `BuildOptions`, `SelectTask` | estender | Separar e ordenar tarefas `Ativas / Na Fila` antes das `Concluídas`, exibindo a etapa da pipeline (`Discovery → Builder → Reviewer`) e tokens por task. |
| Card de detalhes da task | `internal/survey/task_detail.go` | `TaskDetailView`, `BuildTaskDetailView`, `FormatTaskDetailCard` | estender | Incluir seção visual da pipeline (`Discovery → Builder → Reviewer`), consumo de tokens por etapa + total da task e alinhamento exato das bordas com largura visual Unicode. |
| Cálculo de tokens por papel | `internal/monitor/tokens.go` | `CalculateTokenMetrics`, `FormatTokensShort` | estender | Adicionar `CalculateTaskTokenMetrics` para apurar tokens de `Discovery`, `Builder`, `Reviewer` e `Total` por tarefa a partir de `sdd/.metrics/` (com fallback proporcional/estimado). |
| Telemetria automática do worker | `cmd/forge-sdd/worker.go` | `recordJobTelemetry` | estender | Estimar e informar `--tokens-input`, `--tokens-output` e `--duration-seconds` a partir do prompt/spec e log do job para não gravar `0` tokens nos jobs do Builder e Reviewer. |
| Barra inferior direita do tmux | `internal/pipeline/tmux.go` | `BuildTmuxPlan` (`status-right`) | estender | Exibir dicas rotativas de atalhos (`Ctrl+b q`, `Ctrl+b Setas`, `Ctrl+b z`) + resumo de tokens em `status-right` com `-t <name>` (sem `=`). |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-10`, `ARQ-11`: Domínio em `internal/`, stdlib/pacotes existentes, modelo enxuto.
- `COD-01`, `COD-02`, `COD-03`, `COD-05`: Responsabilidade única, `go vet ./...` limpo, idioma `pt-BR`.
- `DUP-01`, `DUP-02`: Reuso do cálculo de tokens em `internal/monitor`.
- `SEG-03`: Comandos tmux com argv explícito e alvo de sessão sem `=` em `set-option`.
- `TST-03`, `TST-04`, `TST-05`: Testes unitários e de integração real com tmux (`FORGE_TMUX_REAL=1`).

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./...
FORGE_TMUX_REAL=1 go test -v -race ./internal/monitor/... ./internal/survey/... ./internal/pipeline/... ./cmd/forge-sdd/...
```
1. Na fila interativa (`QueueDrilldownTUI`), tarefas na fila/em andamento aparecem agrupadas no topo antes das tarefas concluídas, exibindo a etapa correspondente (`Discovery → Builder → Reviewer`).
2. No card de detalhes (`FormatTaskDetailCard`), todas as linhas possuem largura visual idêntica na moldura, exibem o estágio na pipeline (`Discovery ──► Builder ──► Reviewer`) e o gasto de tokens em cada etapa (`Discovery`, `Builder`, `Reviewer`) + `Total da Tarefa`.
3. Na barra `status-right` do tmux (`BuildTmuxPlan`), as dicas de atalhos (`Ctrl+b q`, `Setas`, `z`) aparecem junto aos tokens usando `-t <name>` (sem `=`), funcionando tanto quando `totalTokens == 0` quanto quando `totalTokens > 0`.
