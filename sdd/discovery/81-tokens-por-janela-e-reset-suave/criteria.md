# Critérios Técnicos — Discovery 81 (Monitor de Tokens por Janela e Reset Suave de Sessão)

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica a entrega de comandos e reset no tmux (`internal/pipeline/deliver.go`), a contabilidade de telemetria de tokens (`internal/monitor`, `internal/layout`), a persistência de handoffs compactos e comandos da CLI (`ARQ-01`, `ARQ-11`, `COD-01`, `SEG-01`).

## Decisões do projeto aplicadas

| Decisão | Origem | Impacto nesta iniciativa |
|---|---|---|
| ADR-0003 | Orçamento de Contexto e Skills por Estação | Alinhado com o princípio central de manter o contexto enxuto e estritamente relevante. |
| ADR-0006 | Visibilidade tmux | Exibição da métrica de tokens acumulados diretamente nas bordas dos painéis. |
| ADR-0008 | Sessões com CLI Aberta (Modo Live) | Entrega por buffers in-place sem jamais fechar o terminal nem reiniciar a janela do tmux. |
| D8 | Idioma padrão: pt-BR | Textos de interface e cards de handoff em português padrão. |

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Entrega de reset in-place | `internal/pipeline/deliver.go` | `deliverContextReset` | estender | Integrar com gatilho de auto-archive e sequenciamento de card de handoff. |
| Telemetria de sessão | `cmd/forge-sdd/session.go`, `sdd/.metrics/` | Arquivos JSON de sessão | estender | Agregar tokens de entrada/saída por papel e janela. |
| Monitor de processos | `cmd/forge-sdd/ps.go`, `internal/monitor/` | Comando `forge-sdd ps` | estender | Adicionar coluna de tokens acumulados por componente. |
| Bordas de painel tmux | `internal/pipeline/tmux.go` | `pane-border-format` | estender | Incluir badge dinâmico de tokens: `🪙 <n>k tok`. |
| Arquivamento de features | `cmd/forge-sdd/archive.go`, `.agents/skills/sdd-operacao/references/archive.md` | Lógica de arquivamento | reusar | Acionar auto-archive de features sem dependências após aprovação. |

## Dependências novas
- **Nenhuma.** stdlib Go, `tmux` e arquivos de telemetria já existentes (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- **ARQ-01 / ARQ-10:** Separação entre CLI e lógica de ciclo de vida de sessões.
- **COD-01 / COD-03:** Tratamento resiliente de erros na entrega de prompts e parsing de métricas.
- **DUP-01 / DUP-10:** Reuso do leitor de métricas e documentação completa da mecânica de auto-reset.
- **FLX-01 / GOV-01:** Concepção da feature respeitando os portões de governança.

---

## Protótipos de Interface

### Protótipo 1: Visualização de Tokens por Painel no tmux

As bordas dos painéis agora exibem a contabilidade de tokens acumulados, alertando visualmente quando o contexto estiver elevado:

```text
┌─ [ barbearia-app ❯ forge-discovery │ 🪙 12k tok ] ──┬─ [ barbearia-app ❯ forge-builder │ 🪙 78k tok ⚠️ ] ──┐
│                                                      │ > $ go test -v ./...                               │
│ > forge-discovery                                    │ ok  github.com/meu-app/auth 0.42s                  │
│ Contexto enxuto e focado.                            │                                                    │
│ Próxima demanda pronta para planejar.                │ (Contexto próximo do limiar de 80k: auto-reset      │
│                                                      │  agendado para o término desta feature)            │
│                                                      ├─ [ barbearia-app ❯ forge-reviewer │ 🪙 28k tok ] ───┤
│                                                      │                                                    │
│                                                      │ > forge-reviewer                                   │
│                                                      │ Aguardando próximo PR para revisão...              │
└──────────────────────────────────────────────────────┴────────────────────────────────────────────────────┘
 [ 📦 barbearia-app │ feat/79-clareza ]  0:forge:barbearia-app*      15:30 │ Tokens Totais: 118k (Disc: 12k · Bld: 78k · Rev: 28k)
```

---

### Protótipo 2: O Ciclo de Auto-Archive & Reset Suave In-Place (Sem Piscar Janela)

Quando a task termina com sucesso:

#### Passo A: Conclusão & Auto-Archive
A esteira arquiva a feature concluída e limpa o contexto in-place via buffer:
```text
[forge-ctl] ✓ Task 79-01 aprovada. Auto-archive concluído.
[forge-ctl] Enviando /clear in-place ao painel forge.1...
[forge-ctl] Aguardando painel ocioso (3s)...
```

#### Passo B: Injeção Imediata do Card de Handoff
No **mesmo painel**, sem reabrir nem piscar o terminal, o Builder recebe:

```text
┌─ [ barbearia-app ❯ forge-builder │ 🪙 0.8k tok 🟢 ] ────────────────────────────────────────────────────────┐
│                                                                                                           │
│ > /clear                                                                                                  │
│ Contexto limpo com sucesso. Memória liberada.                                                             │
│                                                                                                           │
│ ────────────────────────────────────────────────────────────────────────                                  │
│ ▶ CONTINUAÇÃO AUTOMÁTICA DA ESTEIRA — HANDOFF DE CONTEXTO                                                  │
│ ────────────────────────────────────────────────────────────────────────                                  │
│ • Iniciativa: 79-clareza-pipeline-e-identificacao-tmux                                                    │
│ • Branch:     feat/79-clareza-pipeline-e-identificacao-tmux                                               │
│ • Status:     Task 79-01 concluída e arquivada com sucesso.                                               │
│ • Próxima:    Task 79-02 (Identificação Visual e Ergonomia no Tmux)                                       │
│ • Escopo:     Estender BuildTmuxPlan para configurar status-left e pane-border-format                     │
│ • Contexto:   Carregados IDs ARQ-01, COD-01 e sdd-arquitetura.md                                          │
│                                                                                                           │
│ Builder, continue a implementação da task 79-02 a partir deste ponto.                                     │
│                                                                                                           │
└───────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Critérios de Aceitação Executáveis

- **CA-01 (Métricas de Tokens por Painel):** `go test -v ./internal/monitor -run TestTokenMetricsPerRole` passa: calcula o total acumulado de tokens (in, out, total) para cada papel (`discovery`, `builder`, `reviewer`) lendo os registros de telemetria da sessão.
- **CA-02 (Borda tmux com Badge de Tokens):** `internal/pipeline/tmux.go` formata o título dos painéis incluindo `🪙 <n>k tok` e atualiza a barra de status com o total consumido.
- **CA-03 (Reset Suave In-Place sem Fechar Terminal):** O teste com runner simulado em `internal/pipeline/deliver_test.go` confirma que o envio de `/clear` e o handoff ocorrem no mesmo painel sem invocar `kill-pane` nem `new-window`.
- **CA-04 (Auto-Archive de Tarefas Concluídas):** Quando a task atinge estado final aprovado e não possui dependências pendentes, os artefatos locais são consolidados e registrados na telemetria sem exigir intervenção manual.
- **CA-05 (Card de Handoff Compacto):** O prompt de handoff gerado pós-reset contém a identificação da task anterior, status da branch e escopo da próxima task em menos de 800 tokens de entrada.
- **CA-06 (Integridade da Suíte):** `go build ./... && go vet ./... && go test -race ./...` e `forge-sdd gate` executam com 100% de sucesso.
