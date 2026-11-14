# ADR-0008: Sessões com CLI Aberta no Tmux (Modo Live e Despachante de Entrega)

- **Data:** 2026-11-14
- **Status:** `aprovado`
- **Regras afetadas (IDs da constitution):** `ARQ-01`, `ARQ-11`, `COD-01`, `COD-03`, `SEG-05`, `TST-03`, `TST-05`, `DUP-02`
- **Autor / Responsável:** Líder (decisão do dono do produto no discovery 77)

## Contexto

No pipeline multi-sessão anterior (ADR-0004 e ADR-0006), os workers do Builder e Reviewer operavam exclusivamente em modo `headless`: a cada job reivindicado na fila, uma nova instância do binário da CLI do agente (Claude, Gemini, Copilot, Codex ou Agy) era inicializada com `--prompt`, executava a tarefa e terminava o processo.

Embora esse modelo oferecesse isolamento total de processos, gerava limitações práticas:
1. **Sobrecarga de inicialização:** cada job pagava o custo de spawn, autenticação e boot do runtime do agente.
2. **Perda de visibilidade interativa:** o usuário via apenas a saída em streaming e o acompanhamento por git, sem a interface viva do agente no terminal.
3. **Impossibilidade de contexto compartilhado opcional:** inviabilizava manter a mesma sessão de raciocínio aberta entre jobs relacionados.

Referência: `sdd/discovery/77-sessoes-com-cli-aberta/` e `feat-77-01` a `feat-77-05`.

## Decisão

- **D-1 Modo de sessão e contexto no `forge.json`.** Introduzimos `session_mode` (`live` vs `headless`, padrão `headless` para compatibilidade retroativa) e `context` (`clear` vs `continuous`, padrão `clear`). O modo `live` mantém as CLIs interativas dos agentes abertas de forma contínua nos painéis visíveis do tmux.
- **D-2 Painéis ao vivo e janela de controle (`forge-ctl`).** No modo `live`, o plano do tmux cria a janela principal `forge` com as CLIs interativas abertas com títulos explícitos e argumentos interativos (`interactive_args`), e uma janela de controle oculta (`forge-ctl`) onde rodam os despachantes de entrega (`forge-sdd worker --role ... --deliver <target>`).
- **D-3 Despachante de entrega por buffer tmux.** A entrega de prompts aos painéis ao vivo ocorre via buffers de tmux (`tmux load-buffer -b <name> -` alimentado por stdin bruto, seguido de `paste-buffer -p -t <target>` e `send-keys Enter`). Esse mecanismo preserva newlines, aspas e caracteres especiais sem depender de escape de shell. O despachante aguarda estabilidade da tela (`ready_quiet_seconds`, padrão 3s) antes de colar e implementa retry único de `Enter` se a tela não mudar em 30s.
- **D-4 Conclusão por estado compartilhado (`VerifyJobHandoff`).** Como a CLI interativa não emite código de saída ao final de cada comando, a conclusão de jobs é observada no barramento de estado:
  - Builder conclui quando o PR é registrado com sucesso (`bus.UpdatePR` / `forge-sdd pipeline pr-opened`), enfileirando automaticamente o job de revisão;
  - Reviewer conclui quando o parecer é aplicado via `bus review` (`ApplyVerdict`).
  O despachante renova periodicamente o lease (`RenewEvery`) e aplica timeout caso o job exceda `job_timeout_minutes`.
- **D-5 Limpeza de contexto (`reset_command`).** Quando `context == clear`, o despachante envia o comando de reset do perfil (`/clear` para Claude, Gemini, Copilot e Agy; `/new` para Codex) via buffer após a conclusão do job e aguarda a ociosidade do painel antes do próximo prompt. Se `reset_command` estiver vazio, assume `continuous` emitindo aviso único.
- **D-6 Falhas lidas da tela (`capture-pane`).** Enquanto aguarda a conclusão, o despachante inspeciona as últimas 20 linhas do painel:
  - Limite de uso (`DetectUsageLimit`): devolve o job à fila (`bus.Requeue`) sem queimar tentativas, emite notificação no inbox e pausa até o horário de reset;
  - Agente indisponível (`DetectAgentUnavailable`): devolve o job à fila (`bus.Requeue`), emite alerta de troca de modelo/tier no inbox e encerra o despachante;
  - Painel fechado ou morto (`CapturePane` falha): devolve o job à fila (`bus.Requeue`), notifica o inbox e encerra o despachante.
- **D-7 Contingência e queda suave.** Se `session_mode: live` for configurado mas o tmux não estiver disponível ou o terminal for `native`, o `forge-sdd up` faz fallback transparente para `headless` exibindo aviso informativo de uma linha.
- **D-8 Configuração retrocompatível e menu interativo.** Arquivos `forge.json` legados sem `session_mode` continuam operando em `headless`. O menu de configuração do pipeline oferece alternar entre os modos `live` e `headless` e entre contextos `clear` e `continuous`.

## Consequências

- **Positivas:** Redução dramática da latência de início de jobs em sessões longas; interface interativa do agente visível em tempo real no tmux; entrega resiliente de prompts sem corrupção por shell; tratamento robusto de limites de cota e indisponibilidade sem travar o pipeline.
- **Negativas / Trade-offs:** O modo `live` depende de sessões tmux funcionais (indisponível em terminais puramente nativos de desktop); o término do trabalho depende da sinalização correta do agente via comandos de barramento (`pr-opened` e `bus review`).
