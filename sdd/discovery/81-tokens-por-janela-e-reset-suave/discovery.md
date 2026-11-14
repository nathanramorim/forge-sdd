# Discovery 81 — Monitor de Tokens por Janela, Auto-Archive e Reset Suave de Sessão sem Reabrir Terminal

## Problema
1. **Acúmulo Excessivo de Contexto no Builder e Reviewer:** Conforme a esteira avança por múltiplas tasks, os agentes interativos nos painéis do tmux acumulam milhares de tokens no histórico da sessão (conversas anteriores, diffs passados, saídas de comandos e logs). Esse inchaço de contexto gera três graves problemas:
   - Aumento drástico no custo e latência a cada mensagem enviada.
   - Perda de foco e degradação da precisão do LLM (alucinações ou regras esquecidas).
   - Risco de atingir os limites de contexto das APIs (ex.: 200k tokens).
2. **Falta de Visibilidade do Consumo por Painel:** O desenvolvedor não consegue ver de relance quantos tokens cada janela (Discovery, Builder, Reviewer) já consumiu, dificultando saber a hora ideal de reiniciar.
3. **Fricção ao Reiniciar Sessões:** Fechar o terminal ou matar o painel do tmux causa flashes visuais, quebra a disposição das telas e perde o estado de onde a esteira parou. A reinicialização deve ser **100% in-place e transparente** no mesmo painel tmux, sem jamais fechar a janela.
4. **Perda de Continuidade (Handoff Quebrado):** Se a pessoa limpa o contexto com `/clear` manualmente, o agente perde a memória da task atual, da branch e de onde parou. O sistema precisa arquivar o concluído e reinjetar automaticamente o contexto mínimo da próxima task.

## Para quem
Desenvolvedores e líderes de projeto que utilizam o pipeline multi-sessão do Forge SDD e precisam manter seus agentes rápidos, focados, com consumo de tokens sob controle e sem interrupções visuais na tela.

## Jornada Alvo

### 1. Visibilidade de Tokens por Painel em Tempo Real
1. O desenvolvedor olha para o topo ou rodapé do painel no tmux e vê o consumo consolidado:
   - `[ barbearia-app ❯ forge-builder │ 🪙 42k tok ]`
2. No comando `forge-sdd ps` ou `forge-sdd tokens`, uma coluna dedicada exibe os tokens de entrada e saída por janela.

### 2. Auto-Archive ao Fechar Feature sem Dependências
1. Quando uma feature é concluída com sucesso (critérios atendidos e PR registrado sem pendências):
   - Se a feature não tem dependências ativas e não gera impacto imediato, a esteira consolida e arquiva o estado da feature (`status.md`, telemetria de sessão).
   - O histórico de conversas passadas daquela entrega não precisa mais residir na memória do agente.

### 3. Reset Suave de Contexto In-Place (Sem Fechar Terminal)
1. O despachante oculta o ciclo de renovação:
   - Envia o comando de reset in-place da CLI (`/clear`, `/new` ou reset suave) diretamente no buffer do painel.
   - O terminal permanece aberto, imóvel e estável (zero reabertura de janelas).
2. A CLI limpa a memória RAM e zera o contexto acumulado do modelo.

### 4. Retomada Transparente com Injeção de Contexto Mínimo
1. Assim que o painel volta a ficar ocioso após a limpeza, o orquestrador injeta imediatamente o **Card de Handoff Compacto**:
   ```text
   ▶ Retomando Esteira: Próxima task feat-XX-YY
   • Branch ativa: feat/XX-nome
   • Tarefa anterior (XX-01) concluída e arquivada
   • Contexto mínimo: regras R-01, ADR-0004
   ```
2. O Builder inicia a próxima tarefa imediatamente com a janela de contexto 100% limpa e com 100% de foco no escopo novo.

## Regras de Negócio
- **RN-01 (Sem Reabertura de Janelas):** O reset de sessão e contexto deve ser estritamente in-place, executado por envio de comando de reset ou reexecução suave do processo da CLI no mesmo painel tmux. É proibido destruir ou recriar painéis tmux durante a esteira ativa.
- **RN-02 (Gatilho de Auto-Archive e Reset):** A limpeza deve ocorrer prioritariamente na fronteira de conclusão de tarefas/features independentes (quando a task atinge estado `approved` ou PR aberto), ou quando o teto de contexto ultrapassar um limiar configurável (padrão: 80k tokens).
- **RN-03 (Garantia de Handoff):** Nenhum reset pode ocorrer sem a persistência prévia do estado em disco e sem a geração do prompt de handoff compacto para a tarefa subsequente.
- **RN-04 (Exibição Amigável de Tokens):** A contagem de tokens deve ser expressa em múltiplos amigáveis (ex.: `1.2k`, `45k`, `120k`) por painel.

## Escopo
- **Dentro:**
  - Extração de métricas de tokens das sessões e arquivos de telemetria (`sdd/.metrics/`).
  - Exibição de tokens por painel na borda tmux (`pane-border-format`) e no comando `forge-sdd ps` / `forge-sdd tokens`.
  - Módulo de reset suave in-place no despachante (`internal/pipeline/deliver.go` e `worker.go`).
  - Geração e injeção do card de handoff compacto pós-reset.
  - Gatilho de auto-archive de tarefas finalizadas.
- **Fora:**
  - Criação de novas janelas do sistema operacional.
  - Modificação dos limites de tokens proprietários dos provedores de LLM.

## Riscos e Mitigações
| Risco | Mitigação |
|---|---|
| Agente não interpretar o comando `/clear` a tempo | Monitor de estabilidade de tela (`waitPaneIdle`) aguarda a CLI ficar ociosa antes de injetar o handoff. |
| Injeção do handoff se misturar com o reset | Sequenciamento com confirmação de tela ociosa entre o envio do `/clear` e o envio do prompt da próxima task. |
| Métricas de tokens indisponíveis em tempo real | Interpolação via telemetria de sessão e tamanho de buffers trafegados. |
